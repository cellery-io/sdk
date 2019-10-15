/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */
package org.cellery.impl;

import com.google.gson.Gson;
import org.cellery.CelleryConstants;
import org.cellery.models.API;
import org.cellery.models.APIDefinition;
import org.cellery.models.AutoScalingResourceMetric;
import org.cellery.models.Cell;
import org.cellery.models.CellSpec;
import org.cellery.models.ClusterIngress;
import org.cellery.models.Component;
import org.cellery.models.ComponentSpec;
import org.cellery.models.Composite;
import org.cellery.models.CompositeSpec;
import org.cellery.models.Destination;
import org.cellery.models.Extension;
import org.cellery.models.GRPC;
import org.cellery.models.Gateway;
import org.cellery.models.GatewaySpec;
import org.cellery.models.GlobalApiPublisher;
import org.cellery.models.HPA;
import org.cellery.models.Ingress;
import org.cellery.models.KPA;
import org.cellery.models.Port;
import org.cellery.models.Resource;
import org.cellery.models.STSTemplate;
import org.cellery.models.STSTemplateSpec;
import org.cellery.models.ScalingPolicy;
import org.cellery.models.TCP;
import org.cellery.models.TLS;
import org.cellery.models.VolumeClaim;
import org.cellery.models.Web;
import org.cellery.models.internal.Dependency;
import org.cellery.models.internal.Image;
import org.cellery.models.internal.ImageComponent;
import io.fabric8.kubernetes.api.model.ConfigMap;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaim;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.Volume;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMount;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BBoolean;
import org.ballerinalang.model.values.BInteger;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.ballerinax.docker.generator.DockerArtifactHandler;
import org.ballerinax.docker.generator.exceptions.DockerGenException;
import org.ballerinax.docker.generator.models.DockerModel;
import org.json.JSONArray;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.IOException;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.stream.Collectors;
import java.util.stream.IntStream;

import static org.cellery.CelleryUtils.copyResourceToTarget;
import static org.cellery.CelleryUtils.getApi;
import static org.cellery.CelleryUtils.getValidName;
import static org.cellery.CelleryUtils.printWarning;
import static org.cellery.CelleryUtils.processEnvVars;
import static org.cellery.CelleryUtils.processProbes;
import static org.cellery.CelleryUtils.processResources;
import static org.cellery.CelleryUtils.processVolumes;
import static org.cellery.CelleryUtils.processWebIngress;
import static org.cellery.CelleryUtils.toYaml;
import static org.cellery.CelleryUtils.writeToFile;

/**
 * Native function cellery:createImage.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createCellImage",
        args = {@Argument(name = "image", type = TypeKind.RECORD),
                @Argument(name = "iName", type = TypeKind.RECORD)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class CreateCellImage extends BlockingNativeCallableUnit {
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir") + File.separator + CelleryConstants.TARGET;
    private static final Logger log = LoggerFactory.getLogger(CreateCellImage.class);

    private Image image = new Image();

    public void execute(Context ctx) {
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        image.setOrgName(((BString) nameStruct.get(CelleryConstants.ORG)).stringValue());
        image.setCellName(((BString) nameStruct.get(CelleryConstants.NAME)).stringValue());
        image.setCellVersion(((BString) nameStruct.get(CelleryConstants.VERSION)).stringValue());
        final BMap cellImageArg = (BMap) ctx.getNullableRefArgument(0);
        image.setCompositeImage("Composite".equals(cellImageArg.getType().getName()));
        LinkedHashMap<?, ?> components = ((BMap) cellImageArg.getMap().get("components")).getMap();
        try {
            processComponents(components);
            generateCellReference();
            generateMetadataFile(components);
            if (image.isCompositeImage()) {
                generateComposite(image);
            } else {
                if (cellImageArg.getMap().containsKey("globalPublisher")) {
                    processGlobalAPIPublisher(((BMap) cellImageArg.getMap().get("globalPublisher")).getMap());
                }
                generateCell(image);
            }
        } catch (BallerinaException e) {
            ctx.setReturnValues(BLangVMErrors.createError(ctx, e.getMessage()));
        }
    }

    private void processGlobalAPIPublisher(LinkedHashMap apiPublisherMap) {
        GlobalApiPublisher globalApiPublisher = new GlobalApiPublisher();
        if (apiPublisherMap.containsKey("apiVersion")) {
            globalApiPublisher.setVersion(((BString) apiPublisherMap.get("apiVersion")).stringValue());
        }
        if (apiPublisherMap.containsKey(CelleryConstants.CONTEXT)) {
            globalApiPublisher.setContext(((BString) apiPublisherMap.get(CelleryConstants.CONTEXT)).stringValue());
        }
        image.setGlobalApiPublisher(globalApiPublisher);
    }

    private void processComponents(LinkedHashMap<?, ?> components) {
        components.forEach((componentKey, componentValue) -> {
            ImageComponent component = new ImageComponent();
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            // Set mandatory fields.
            component.setName(((BString) attributeMap.get("name")).stringValue());
            component.setReplicas((int) (((BInteger) attributeMap.get("replicas")).intValue()));
            component.setService(component.getName());
            component.setType(((BString) attributeMap.get("componentType")).stringValue());
            processSource(component, attributeMap);
            //Process Optional fields
            if (attributeMap.containsKey(CelleryConstants.INGRESSES)) {
                processIngress(((BMap<?, ?>) attributeMap.get(CelleryConstants.INGRESSES)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.LABELS)) {
                ((BMap<?, ?>) attributeMap.get(CelleryConstants.LABELS)).getMap().forEach((labelKey, labelValue) ->
                        component.addLabel(labelKey.toString(), labelValue.toString()));
            }
            if (attributeMap.containsKey(CelleryConstants.SCALING_POLICY)) {
                processAutoScalePolicy(((BMap<?, ?>) attributeMap.get(CelleryConstants.SCALING_POLICY)), component);
            }
            if (attributeMap.containsKey(CelleryConstants.ENV_VARS)) {
                processEnvVars(((BMap<?, ?>) attributeMap.get(CelleryConstants.ENV_VARS)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.PROBES)) {
                processProbes(((BMap<?, ?>) attributeMap.get(CelleryConstants.PROBES)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.POD_RESOURCES)) {
                processResources(((BMap<?, ?>) attributeMap.get(CelleryConstants.POD_RESOURCES)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.VOLUMES)) {
                processVolumes(((BMap<?, ?>) attributeMap.get(CelleryConstants.VOLUMES)).getMap(), component);
            }
            image.addComponent(component);
        });
    }

    private void processSource(ImageComponent component, LinkedHashMap attributeMap) {
        if ("ImageSource".equals(((BValue) attributeMap.get(CelleryConstants.IMAGE_SOURCE)).getType().getName())) {
            //Image Source
            component.setSource(((BString) ((BMap) attributeMap.get(CelleryConstants.IMAGE_SOURCE)).getMap()
                    .get("image")).stringValue());
        } else {
            // Docker Source
            LinkedHashMap dockerSourceMap = ((BMap) attributeMap.get(CelleryConstants.IMAGE_SOURCE)).getMap();
            String tag = ((BString) dockerSourceMap.get("tag")).stringValue();
            if (!tag.matches("[^/]+")) {
                // <IMAGE_NAME>:1.0.0
                throw new BallerinaException("Invalid docker tag: " + tag + ". Repository name is not supported when " +
                        "building from Dockerfile");
            }
            tag = image.getOrgName() + "/" + tag;
            createDockerImage(tag, ((BString) dockerSourceMap.get("dockerDir")).stringValue());
            component.setDockerPushRequired(true);
            component.setSource(tag);
        }
    }

    /**
     * Extract the ingresses.
     *
     * @param ingressMap list of ingresses defined
     * @param component  current component
     */
    private void processIngress(LinkedHashMap<?, ?> ingressMap, ImageComponent component) {
        ingressMap.forEach((key, ingressValues) -> {
            BMap ingressValueMap = ((BMap) ingressValues);
            LinkedHashMap attributeMap = ingressValueMap.getMap();
            switch (ingressValueMap.getType().getName()) {
                case "HttpApiIngress":
                case "HttpPortIngress":
                case "HttpsPortIngress":
                    processHttpIngress(component, attributeMap, key.toString());
                    break;
                case "TCPIngress":
                    processTCPIngress(component, attributeMap);
                    break;
                case "GRPCIngress":
                    processGRPCIngress(component, attributeMap);
                    break;
                case "WebIngress":
                    processWebIngress(component, attributeMap);
                    break;

                default:
                    break;
            }
        });
    }

    private void processGRPCIngress(ImageComponent component, LinkedHashMap attributeMap) {
        GRPC grpc = new GRPC();
        Destination destination = new Destination();
        destination.setPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        destination.setHost(component.getService());
        if (attributeMap.containsKey(CelleryConstants.PROTO_FILE)) {
            String protoFile = ((BString) attributeMap.get(CelleryConstants.PROTO_FILE)).stringValue();
            if (!protoFile.isEmpty()) {
                copyResourceToTarget(protoFile);
            }
        }
        component.setContainerPort(destination.getPort());
        Port port = new Port();
        port.setName(component.getName());
        port.setPort(destination.getPort());
        port.setProtocol(CelleryConstants.PROTOCOL_GRPC);
        port.setTargetContainer(component.getName());
        port.setTargetPort(component.getContainerPort());
        grpc.setDestination(destination);
        if (attributeMap.containsKey(CelleryConstants.GATEWAY_PORT)) {
            grpc.setPort((int) ((BInteger) attributeMap.get(CelleryConstants.GATEWAY_PORT)).intValue());
        }
        component.addPort(port);
        component.addGRPC(grpc);
    }

    private void processTCPIngress(ImageComponent component, LinkedHashMap attributeMap) {
        TCP tcp = new TCP();
        Destination destination = new Destination();
        destination.setHost(component.getService());
        destination.setPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        tcp.setDestination(destination);
        component.setProtocol(CelleryConstants.PROTOCOL_TCP);
        component.setContainerPort(destination.getPort());
        Port port = new Port();
        port.setName(component.getName());
        port.setPort(destination.getPort());
        port.setProtocol(CelleryConstants.PROTOCOL_TCP);
        port.setTargetContainer(component.getName());
        port.setTargetPort(component.getContainerPort());
        if (attributeMap.containsKey(CelleryConstants.GATEWAY_PORT)) {
            tcp.setPort((int) ((BInteger) attributeMap.get(CelleryConstants.GATEWAY_PORT)).intValue());
        }
        component.addPort(port);
        component.addTCP(tcp);
    }

    private void processHttpIngress(ImageComponent component, LinkedHashMap attributeMap, String key) {
        API httpAPI = getApi(component, attributeMap);
        httpAPI.setName(key);
        httpAPI.setPort(CelleryConstants.DEFAULT_GATEWAY_PORT);
        component.setProtocol(CelleryConstants.DEFAULT_GATEWAY_PROTOCOL);
        // Process optional attributes
        if (attributeMap.containsKey(CelleryConstants.CONTEXT)) {
            String context = ((BString) attributeMap.get(CelleryConstants.CONTEXT)).stringValue();
            if (!context.startsWith("/")) {
                context = "/" + context;
            }
            httpAPI.setContext(context);
        }
        if (attributeMap.containsKey("authenticate")) {
            httpAPI.setAuthenticate(((BBoolean) attributeMap.get("authenticate")).booleanValue());
        }
        if (attributeMap.containsKey(CelleryConstants.EXPOSE)) {
            if (!httpAPI.isAuthenticate()) {
                component.addUnsecuredPaths(httpAPI.getContext());
            }
            if ("global".equals(((BString) attributeMap.get(CelleryConstants.EXPOSE)).stringValue())) {
                httpAPI.setGlobal(true);
            } else if ("local".equals(((BString) attributeMap.get(CelleryConstants.EXPOSE)).stringValue())) {
                httpAPI.setGlobal(false);
            }
            if (attributeMap.containsKey("definition")) {
                List<APIDefinition> apiDefinitions = new ArrayList<>();
                BValueArray resourceDefs =
                        (BValueArray) ((BMap<?, ?>) attributeMap.get("definition")).getMap().get("resources");
                IntStream.range(0, (int) resourceDefs.size()).forEach(resourceIndex -> {
                    APIDefinition apiDefinition = new APIDefinition();
                    LinkedHashMap definitions = ((BMap) resourceDefs.getBValue(resourceIndex)).getMap();
                    apiDefinition.setPath(((BString) definitions.get("path")).stringValue());
                    apiDefinition.setMethod(((BString) definitions.get("method")).stringValue());
                    apiDefinitions.add(apiDefinition);
                });
                if (apiDefinitions.size() > 0) {
                    httpAPI.setDefinitions(apiDefinitions);
                }
            }
        }
        Destination destination = new Destination();
        destination.setHost(component.getName());
        destination.setPort(CelleryConstants.DEFAULT_GATEWAY_PORT);
        httpAPI.setDestination(destination);
        Port port = new Port();
        port.setName(component.getName());
        port.setPort(CelleryConstants.DEFAULT_GATEWAY_PORT);
        port.setProtocol(CelleryConstants.DEFAULT_GATEWAY_PROTOCOL);
        port.setTargetContainer(component.getName());
        port.setTargetPort(component.getContainerPort());
        component.addPort(port);
        component.addApi(httpAPI);
    }


    /**
     * Extract the scale policy.
     *
     * @param scalePolicy Scale policy to be processed
     * @param component   current component
     */
    private void processAutoScalePolicy(BMap<?, ?> scalePolicy, ImageComponent component) {
        ScalingPolicy scalingPolicy = new ScalingPolicy();
        scalingPolicy.setReplicas(component.getReplicas());
        LinkedHashMap bScalePolicy = scalePolicy.getMap();
        boolean bOverridable = false;
        if ("AutoScalingPolicy".equals(scalePolicy.getType().getName())) {
            // Autoscaling
            image.setAutoScaling(true);
            HPA hpa = new HPA();
            hpa.setMaxReplicas(((BInteger) bScalePolicy.get(CelleryConstants.MAX_REPLICAS)).intValue());
            hpa.setMinReplicas(((BInteger) (bScalePolicy.get("minReplicas"))).intValue());
            bOverridable = ((BBoolean) bScalePolicy.get("overridable")).booleanValue();
            LinkedHashMap metricsMap = ((BMap) bScalePolicy.get("metrics")).getMap();
            if (metricsMap.containsKey(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_CPU)) {
                hpa.addMetric(extractMetrics(metricsMap, CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_CPU));
            }
            if (metricsMap.containsKey(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_MEMORY)) {
                hpa.addMetric(extractMetrics(metricsMap, CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_MEMORY));
            }
            scalingPolicy.setHpa(hpa);
        } else {
            //Zero Scaling
            image.setZeroScaling(true);
            KPA kpa = new KPA();
            kpa.setMinReplicas(0);
            if (bScalePolicy.containsKey(CelleryConstants.MAX_REPLICAS)) {
                kpa.setMaxReplicas(((BInteger) bScalePolicy.get(CelleryConstants.MAX_REPLICAS)).intValue());
            }
            if (bScalePolicy.containsKey(CelleryConstants.CONCURRENCY_TARGET)) {
                kpa.setConcurrency(((BInteger) bScalePolicy.get(CelleryConstants.CONCURRENCY_TARGET)).intValue());
            }
            scalingPolicy.setKpa(kpa);
        }
        scalingPolicy.setOverridable(bOverridable);
        component.setScalingPolicy(scalingPolicy);
    }

    private AutoScalingResourceMetric extractMetrics(LinkedHashMap metricsMap, String resourceName) {
        AutoScalingResourceMetric scalingResourceMetric = new AutoScalingResourceMetric();
        scalingResourceMetric.setType(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE);
        Resource resource = new Resource();
        resource.setName(resourceName);
        final BValue bValue = (BValue) ((BMap) metricsMap.get(resourceName)).getMap().get("threshold");
        if (CelleryConstants.BTYPE_STRING.equals(bValue.getType().getName())) {
            HashMap<String, Object> hs = new HashMap<>();
            hs.put("type", "AverageValue");
            hs.put("averageValue", bValue.stringValue());
            resource.setTarget(hs);
        } else {
            HashMap<String, Object> hs = new HashMap<>();
            hs.put("type", "Utilization");
            hs.put("averageUtilization", Integer.parseInt(bValue.stringValue()));
            resource.setTarget(hs);
        }
        scalingResourceMetric.setResource(resource);
        return scalingResourceMetric;
    }

    private void generateComposite(Image image) {
        List<ImageComponent> components =
                new ArrayList<>(image.getComponentNameToComponentMap().values());
        CompositeSpec compositeSpec = new CompositeSpec();
        components.forEach(imageComponent -> {
            Component component = new Component();
            ComponentSpec componentSpec = getSpec(imageComponent, component);
            component.setSpec(componentSpec);
            compositeSpec.addComponent(component);
        });
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(image.getCellName()))
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_ORG, image.getOrgName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_NAME, image.getCellName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION, image.getCellVersion())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES, new Gson().toJson(image.getDependencies()))
                .build();
        Composite composite = new Composite(objectMeta, compositeSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + CelleryConstants.CELLERY + File.separator
                        + image.getCellName() + CelleryConstants.YAML;
        try {
            writeToFile(toYaml(composite), targetPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while writing composite yaml " + targetPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private ComponentSpec getSpec(ImageComponent imageComponent, Component component) {
        component.setMetadata(new ObjectMetaBuilder().withName(imageComponent.getName())
                .withLabels(imageComponent.getLabels())
                .build());
        PodSpec componentTemplate = new PodSpec();
        Container container = getContainer(imageComponent, getEnvVars(imageComponent));
        ComponentSpec componentSpec = new ComponentSpec();
        componentSpec.setType(imageComponent.getType());
        componentSpec.setScalingPolicy(imageComponent.getScalingPolicy());
        componentSpec.setPorts(imageComponent.getPorts());
        List<VolumeMount> volumeMounts = new ArrayList<>();
        List<Volume> volumes = new ArrayList<>();
        imageComponent.getVolumes().forEach(volumeInfo -> {
            String name;
            if (volumeInfo.getVolume() instanceof ConfigMap) {
                name = ((ConfigMap) volumeInfo.getVolume())
                        .getMetadata().getName();
                Volume volume = new VolumeBuilder()
                        .withName(name)
                        .withNewConfigMap()
                        .withName(name)
                        .endConfigMap()
                        .build();
                if (!volumeInfo.isShared()) {
                    componentSpec.addConfiguration((ConfigMap) volumeInfo.getVolume());
                } else {
                    volumes.add(volume);
                }
            } else if (volumeInfo.getVolume() instanceof Secret) {
                name = ((Secret) volumeInfo.getVolume())
                        .getMetadata().getName();
                Volume volume = new VolumeBuilder()
                        .withName(name)
                        .withNewSecret()
                        .withSecretName(name)
                        .endSecret()
                        .build();
                if (!volumeInfo.isShared()) {
                    componentSpec.addSecrets((Secret) volumeInfo.getVolume());
                } else {
                    volumes.add(volume);
                }
            } else {
                name = ((PersistentVolumeClaim) volumeInfo.getVolume()).getMetadata().getName();
                Volume volume = new VolumeBuilder()
                        .withName(name)
                        .withNewPersistentVolumeClaim()
                        .withClaimName(name)
                        .endPersistentVolumeClaim()
                        .build();
                if (!volumeInfo.isShared()) {
                    VolumeClaim sharedVolumeClaim = new VolumeClaim();
                    sharedVolumeClaim.setShared(true);
                    sharedVolumeClaim.setTemplate((PersistentVolumeClaim) volumeInfo.getVolume());
                    sharedVolumeClaim.setName(name);
                    componentSpec.addVolumeClaim(sharedVolumeClaim);
                } else {
                    volumes.add(volume);
                }
            }
            VolumeMount volumeMount = new VolumeMountBuilder()
                    .withName(name)
                    .withMountPath(volumeInfo.getPath())
                    .withReadOnly(volumeInfo.isReadOnly()).build();
            volumeMounts.add(volumeMount);
        });
        container.setVolumeMounts(volumeMounts);
        componentTemplate.setVolumes(volumes);
        componentTemplate.setContainers(Collections.singletonList(container));
        componentSpec.setTemplate(componentTemplate);
        return componentSpec;
    }

    private void generateCell(Image image) {
        List<ImageComponent> components = new ArrayList<>(image.getComponentNameToComponentMap().values());
        Ingress ingress = new Ingress();
        Extension extension = new Extension();
        extension.setApiPublisher(image.getGlobalApiPublisher());
        CellSpec cellSpec = new CellSpec();
        List<String> unsecuredPaths = new ArrayList<>();
        components.forEach(imageComponent -> {
            unsecuredPaths.addAll(imageComponent.getUnsecuredPaths());
            ingress.addHttpAPI(imageComponent.getApis());
            ingress.addGRPC(imageComponent.getGrpcList().stream()
                    .filter(a -> a.getPort() > 0).collect(Collectors.toList()));
            ingress.addTCP(imageComponent.getTcpList().stream()
                    .filter(a -> a.getPort() > 0).collect(Collectors.toList()));
            final Web web = imageComponent.getWeb();
            if (web != null) {
                extension.setOidc(web.getOidc());
                ClusterIngress clusterIngress = new ClusterIngress();
                TLS tls = new TLS();
                tls.setCert(web.getTlsCert());
                tls.setKey(web.getTlsKey());
                clusterIngress.setTls(tls);
                clusterIngress.setHost(web.getVhost());
                extension.setOidc(web.getOidc());
                extension.setClusterIngress(clusterIngress);
                ingress.addHttpAPI((Collections.singletonList(web.getHttpAPI())));
            }
            Component component = new Component();
            ComponentSpec componentSpec = getSpec(imageComponent, component);
            component.setSpec(componentSpec);
            cellSpec.addComponent(component);
        });
        STSTemplate stsTemplate = new STSTemplate();
        STSTemplateSpec stsTemplateSpec = new STSTemplateSpec();
        stsTemplateSpec.setUnsecuredPaths(unsecuredPaths);
        stsTemplate.setSpec(stsTemplateSpec);
        Gateway gateway = new Gateway();
        GatewaySpec gatewaySpec = new GatewaySpec();
        gatewaySpec.setIngress(ingress);
        gateway.setSpec(gatewaySpec);

        ingress.setExtensions(extension);
        cellSpec.setGateway(gateway);
        cellSpec.setSts(stsTemplate);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(image.getCellName()))
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_ORG, image.getOrgName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_NAME, image.getCellName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION, image.getCellVersion())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES, new Gson().toJson(image.getDependencies()))
                .build();
        Cell cell = new Cell(objectMeta, cellSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + CelleryConstants.CELLERY + File.separator + image.getCellName() + CelleryConstants.YAML;
        try {
            writeToFile(toYaml(cell), targetPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while writing cell yaml " + targetPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private Container getContainer(ImageComponent component, List<EnvVar> envVarList) {
        ContainerBuilder containerBuilder = new ContainerBuilder()
                .withName(component.getName())
                .withImage(component.getSource())
                .withEnv(envVarList)
                .withResources(component.getResources())
                .withReadinessProbe(component.getReadinessProbe())
                .withLivenessProbe(component.getLivenessProbe());
        if (component.getContainerPort() != 0) {
            containerBuilder.withPorts(new ContainerPortBuilder()
                    .withContainerPort(component.getContainerPort()).build());
        }
        return containerBuilder.build();
    }

    private List<EnvVar> getEnvVars(ImageComponent component) {
        List<EnvVar> envVarList = new ArrayList<>();
        component.getEnvVars().forEach((key, value) -> {
            if (StringUtils.isEmpty(value)) {
                printWarning("Value is empty for environment variable \"" + key + "\"");
            }
            envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
        });
        return envVarList;
    }

    /**
     * Generate a Cell Reference that can be used by other cells.
     */
    private void generateCellReference() {
        JSONObject json = new JSONObject();
        if (image.isCompositeImage()) {
            image.getComponentNameToComponentMap().forEach((name, component) -> {
                String componentName = getValidName(component.getName());
                json.put(componentName + "_host", CelleryConstants.INSTANCE_NAME_PLACEHOLDER + "--" + componentName + "-service");
                if (component.getApis().size() > 0) {
                    json.put(componentName + "_port", CelleryConstants.DEFAULT_GATEWAY_PORT);
                }
                component.getTcpList().forEach(tcp -> json.put(componentName + "_tcp_port", tcp.getPort()));
                component.getGrpcList().forEach(grpc -> json.put(componentName + "_grpc_port", grpc.getPort()));
            });
        } else {
            image.getComponentNameToComponentMap().forEach((componentName, component) -> {
                String recordName = getValidRecordName(componentName);
                component.getApis().forEach(api -> {
                    String context = api.getContext();
                    if (StringUtils.isNotEmpty(context)) {
                        String url =
                                CelleryConstants.DEFAULT_GATEWAY_PROTOCOL + "://" + CelleryConstants.INSTANCE_NAME_PLACEHOLDER + CelleryConstants.GATEWAY_SERVICE + ":"
                                        + CelleryConstants.DEFAULT_GATEWAY_PORT + "/" + context;
                        json.put(recordName + "_" + getValidRecordName(api.getName()) + "_api_url",
                                url.replaceAll("(?<!http:)//", "/"));
                    }
                });
                component.getTcpList().forEach(tcp -> json.put(recordName + "_tcp_port", tcp.getPort()));
                component.getGrpcList().forEach(grpc -> json.put(recordName + "_grpc_port", grpc.getPort()));
            });
            json.put("gateway_host", CelleryConstants.INSTANCE_NAME_PLACEHOLDER + CelleryConstants.GATEWAY_SERVICE);
        }
        String targetFileNameWithPath =
                OUTPUT_DIRECTORY + File.separator + "ref" + File.separator + CelleryConstants.REFERENCE_FILE_NAME;
        try {
            writeToFile(json.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while generating reference file " + targetFileNameWithPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    /**
     * Generate the metadata json without dependencies.
     *
     * @param components Components from which data should be extracted for metadata
     */
    private void generateMetadataFile(LinkedHashMap<?, ?> components) {
        JSONObject jsonObject = new JSONObject();
        jsonObject.put(CelleryConstants.KIND, image.isCompositeImage() ? "Composite" : "Cell");
        jsonObject.put(CelleryConstants.ORG, image.getOrgName());
        jsonObject.put(CelleryConstants.NAME, image.getCellName());
        jsonObject.put(CelleryConstants.VERSION, image.getCellVersion());
        jsonObject.put("zeroScalingRequired", image.isZeroScaling());
        jsonObject.put("autoScalingRequired", image.isAutoScaling());

        JSONObject componentsJsonObject = new JSONObject();
        components.forEach((key, componentValue) -> {
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            String componentName = ((BString) attributeMap.get("name")).stringValue();
            JSONObject componentJson = new JSONObject();

            JSONObject labelsJsonObject = new JSONObject();
            if (attributeMap.containsKey(CelleryConstants.LABELS)) {
                ((BMap<?, ?>) attributeMap.get(CelleryConstants.LABELS)).getMap().forEach((labelKey, labelValue) ->
                        labelsJsonObject.put(labelKey.toString(), labelValue.toString()));
            }
            componentJson.put("labels", labelsJsonObject);

            JSONObject cellDependenciesJsonObject = new JSONObject();
            JSONObject compositeDependenciesJsonObject = new JSONObject();
            JSONArray componentDependenciesJsonArray = new JSONArray();
            if (attributeMap.containsKey(CelleryConstants.DEPENDENCIES)) {
                LinkedHashMap<?, ?> dependencies = ((BMap<?, ?>) attributeMap.get(CelleryConstants.DEPENDENCIES)).getMap();
                if (dependencies.containsKey(CelleryConstants.CELLS)) {
                    LinkedHashMap<?, ?> cellDependencies = ((BMap) dependencies.get(CelleryConstants.CELLS)).getMap();
                    extractDependencies(cellDependenciesJsonObject, cellDependencies, CelleryConstants.CELL);
                }
                if (dependencies.containsKey(CelleryConstants.COMPOSITES)) {
                    LinkedHashMap<?, ?> compositeDependencies = ((BMap) dependencies.get(CelleryConstants.COMPOSITES)).getMap();
                    extractDependencies(compositeDependenciesJsonObject, compositeDependencies, CelleryConstants.COMPOSITE);
                }
                if (dependencies.containsKey(CelleryConstants.COMPONENTS)) {
                    BValueArray componentsArray = ((BValueArray) dependencies.get(CelleryConstants.COMPONENTS));
                    IntStream.range(0, (int) componentsArray.size()).forEach(componentIndex -> {
                        LinkedHashMap component = ((BMap) componentsArray.getBValue(componentIndex)).getMap();
                        componentDependenciesJsonArray.put(((BString) component.get("name")).stringValue());
                    });
                }
            }
            JSONObject dependenciesJsonObject = new JSONObject();
            dependenciesJsonObject.put(CelleryConstants.CELLS, cellDependenciesJsonObject);
            dependenciesJsonObject.put(CelleryConstants.COMPOSITES, compositeDependenciesJsonObject);
            dependenciesJsonObject.put(CelleryConstants.COMPONENTS, componentDependenciesJsonArray);
            componentJson.put("dependencies", dependenciesJsonObject);

            componentsJsonObject.put(componentName, componentJson);
        });
        image.getComponentNameToComponentMap().forEach((componentName, component) -> {
            JSONObject componentJsonObject = componentsJsonObject.getJSONObject(componentName);
            componentJsonObject.put("dockerImage", component.getSource());
            componentJsonObject.put("isDockerPushRequired", component.isDockerPushRequired());
        });
        jsonObject.put("components", componentsJsonObject);

        String targetFileNameWithPath =
                OUTPUT_DIRECTORY + File.separator + CelleryConstants.CELLERY + File.separator + CelleryConstants.METADATA_FILE_NAME;
        try {
            writeToFile(jsonObject.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while generating metadata file " + targetFileNameWithPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private String getValidRecordName(String name) {
        return name.toLowerCase(Locale.getDefault()).replaceAll("\\P{Alnum}", "");
    }

    private void extractDependencies(JSONObject dependenciesJsonObject, LinkedHashMap<?, ?> cellDependencies,
                                     String kind) {
        cellDependencies.forEach((alias, dependencyValue) -> {
            JSONObject dependencyJsonObject = new JSONObject();
            String org, name, version;
            if (CelleryConstants.BTYPE_STRING.equals(((BValue) dependencyValue).getType().getName())) {
                String dependency = ((BString) (dependencyValue)).stringValue();
                // Validate dependency text
                if (dependency.matches("^([^/:]*)/([^/:]*):([^/:]*)$")) {
                    String[] dependencyVersionSplit = dependency.split(":");
                    String[] dependencySplit = dependencyVersionSplit[0].split("/");
                    org = dependencySplit[0];
                    name = dependencySplit[1];
                    version = dependencyVersionSplit[1];
                } else {
                    throw new BallerinaException("expects <organization>/<cell-image>:<version> " +
                            "as the dependency, received " + dependency);
                }
            } else {
                LinkedHashMap dependency = ((BMap) dependencyValue).getMap();
                org = ((BString) dependency.get(CelleryConstants.ORG)).stringValue();
                name = ((BString) dependency.get(CelleryConstants.NAME)).stringValue();
                version = ((BString) dependency.get(CelleryConstants.VERSION)).stringValue();
            }
            dependencyJsonObject.put(CelleryConstants.ORG, org);
            dependencyJsonObject.put(CelleryConstants.NAME, name);
            dependencyJsonObject.put(CelleryConstants.VERSION, version);
            dependencyJsonObject.put("alias", alias.toString());
            dependencyJsonObject.put(CelleryConstants.KIND, kind);
            image.addDependency(new Dependency(org, name, version, alias.toString(), kind));
            dependenciesJsonObject.put(alias.toString(), dependencyJsonObject);
        });
    }

    /**
     * Create a Docker Image from Dockerfile.
     *
     * @param dockerImageTag Tag for docker image
     * @param dockerDir      Path to docker Directory
     */
    private void createDockerImage(String dockerImageTag, String dockerDir) {
        DockerModel dockerModel = new DockerModel();
        dockerModel.setName(dockerImageTag);
        try {
            DockerArtifactHandler dockerArtifactHandler = new DockerArtifactHandler(dockerModel);
            dockerArtifactHandler.buildImage(dockerModel, Paths.get(dockerDir));
        } catch (DockerGenException | InterruptedException e) {
            String errMsg = "Error occurred while building Docker image ";
            log.error(errMsg, e);
            throw new BallerinaException(errMsg + e.getMessage());
        }
    }
}
