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
package io.cellery.impl;

import com.google.gson.Gson;
import io.cellery.CelleryConstants;
import io.cellery.exception.BallerinaCelleryException;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.AutoScalingResourceMetric;
import io.cellery.models.Cell;
import io.cellery.models.CellSpec;
import io.cellery.models.ClusterIngress;
import io.cellery.models.Component;
import io.cellery.models.ComponentSpec;
import io.cellery.models.Composite;
import io.cellery.models.CompositeSpec;
import io.cellery.models.Destination;
import io.cellery.models.Extension;
import io.cellery.models.GRPC;
import io.cellery.models.Gateway;
import io.cellery.models.GatewaySpec;
import io.cellery.models.GlobalApiPublisher;
import io.cellery.models.HPA;
import io.cellery.models.Ingress;
import io.cellery.models.KPA;
import io.cellery.models.Port;
import io.cellery.models.Resource;
import io.cellery.models.STSTemplate;
import io.cellery.models.STSTemplateSpec;
import io.cellery.models.ScalingPolicy;
import io.cellery.models.TCP;
import io.cellery.models.TLS;
import io.cellery.models.VolumeClaim;
import io.cellery.models.Web;
import io.cellery.models.internal.Dependency;
import io.cellery.models.internal.Image;
import io.cellery.models.internal.ImageComponent;
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
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;
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
import java.util.List;
import java.util.Locale;
import java.util.stream.Collectors;
import java.util.stream.IntStream;

import static io.cellery.CelleryConstants.BTYPE_STRING;
import static io.cellery.CelleryConstants.CELL;
import static io.cellery.CelleryConstants.CELLS;
import static io.cellery.CelleryConstants.COMPONENTS;
import static io.cellery.CelleryConstants.COMPOSITE;
import static io.cellery.CelleryConstants.COMPOSITES;
import static io.cellery.CelleryConstants.DEPENDENCIES;
import static io.cellery.CelleryConstants.ENV_VARS;
import static io.cellery.CelleryConstants.IMAGE_SOURCE;
import static io.cellery.CelleryConstants.INGRESSES;
import static io.cellery.CelleryConstants.KIND;
import static io.cellery.CelleryConstants.LABELS;
import static io.cellery.CelleryConstants.METADATA_FILE_NAME;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.POD_RESOURCES;
import static io.cellery.CelleryConstants.PROBES;
import static io.cellery.CelleryConstants.SCALING_POLICY;
import static io.cellery.CelleryConstants.VERSION;
import static io.cellery.CelleryUtils.copyResourceToTarget;
import static io.cellery.CelleryUtils.getApi;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.processEnvVars;
import static io.cellery.CelleryUtils.processProbes;
import static io.cellery.CelleryUtils.processResources;
import static io.cellery.CelleryUtils.processVolumes;
import static io.cellery.CelleryUtils.processWebIngress;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * Native function cellery:createImage.
 */
public class CreateCellImage {
    private static final String OUTPUT_DIRECTORY =
            System.getProperty("user.dir") + File.separator + CelleryConstants.TARGET;
    private static final Logger log = LoggerFactory.getLogger(CreateCellImage.class);

    private static Image image = new Image();

    public static void createCellImage(MapValue cellImage, MapValue imageName) throws BallerinaCelleryException {
        image.setOrgName((imageName.getStringValue(ORG)));
        image.setCellName((imageName.getStringValue(NAME)));
        image.setCellVersion((imageName.getStringValue(VERSION)));
        image.setCompositeImage("Composite".equals(cellImage.getType().getName()));
        MapValue<?, ?> components = cellImage.getMapValue("components");
        try {
            processComponents(components);
            generateCellReference();
            generateMetadataFile(components);
            if (image.isCompositeImage()) {
                generateComposite(image);
            } else {
                if (cellImage.containsKey("globalPublisher")) {
                    processGlobalAPIPublisher(cellImage.getMapValue("globalPublisher"));
                }
                generateCell(image);
            }
        } catch (BallerinaException e) {
            throw new BallerinaCelleryException(e.getMessage());
        }
    }

    private static void processGlobalAPIPublisher(MapValue apiPublisherMap) {
        GlobalApiPublisher globalApiPublisher = new GlobalApiPublisher();
        if (apiPublisherMap.containsKey("apiVersion")) {
            globalApiPublisher.setVersion(apiPublisherMap.getStringValue("apiVersion"));
        }
        if (apiPublisherMap.containsKey(CelleryConstants.CONTEXT)) {
            globalApiPublisher.setContext(apiPublisherMap.getStringValue(CelleryConstants.CONTEXT));
        }
        image.setGlobalApiPublisher(globalApiPublisher);
    }

    private static void processComponents(MapValue<?, ?> components) {
        components.forEach((componentKey, componentValue) -> {
            ImageComponent component = new ImageComponent();
            MapValue attributeMap = ((MapValue) componentValue);
            // Set mandatory fields.
            component.setName(attributeMap.getStringValue("name"));
            component.setReplicas(Math.toIntExact(attributeMap.getIntValue("replicas")));
            component.setService(component.getName());
            component.setType(attributeMap.getStringValue("type"));
            processSource(component, attributeMap);
            //Process Optional fields
            if (attributeMap.containsKey(INGRESSES)) {
                processIngress(((MapValue<?, ?>) attributeMap.getMapValue(INGRESSES)), component);
            }
            if (attributeMap.containsKey(LABELS)) {
                ((MapValue<?, ?>) attributeMap.getMapValue(LABELS)).forEach((labelKey, labelValue) ->
                        component.addLabel(labelKey.toString(), labelValue.toString()));
            }
            if (attributeMap.containsKey(SCALING_POLICY)) {
                processAutoScalePolicy(attributeMap.getMapValue(SCALING_POLICY), component);
            }
            if (attributeMap.containsKey(ENV_VARS)) {
                processEnvVars(attributeMap.getMapValue(ENV_VARS), component);
            }
            if (attributeMap.containsKey(PROBES)) {
                processProbes(attributeMap.getMapValue(PROBES), component);
            }
            if (attributeMap.containsKey(POD_RESOURCES)) {
                processResources(attributeMap.getMapValue(POD_RESOURCES), component);
            }
            if (attributeMap.containsKey(CelleryConstants.VOLUMES)) {
                processVolumes(attributeMap.getMapValue(POD_RESOURCES), component);
            }
            image.addComponent(component);
        });
    }

    private static void processSource(ImageComponent component, MapValue attributeMap) {
        if ("ImageSource".equals(attributeMap.getMapValue(IMAGE_SOURCE).getType().getName())) {
            //Image Source
            component.setSource(attributeMap.getMapValue(IMAGE_SOURCE).getStringValue("image"));
        } else {
            // Docker Source
            MapValue dockerSourceMap = attributeMap.getMapValue(IMAGE_SOURCE);
            String tag = dockerSourceMap.getStringValue("tag");
            if (!tag.matches("[^/]+")) {
                // <IMAGE_NAME>:1.0.0
                throw new BallerinaException("Invalid docker tag: " + tag + ". Repository name is not supported when " +
                        "building from Dockerfile");
            }
            tag = image.getOrgName() + "/" + tag;
            createDockerImage(tag, dockerSourceMap.getStringValue("dockerDir"));
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
    private static void processIngress(MapValue<?, ?> ingressMap, ImageComponent component) {
        ingressMap.forEach((key, ingressValues) -> {
            MapValue ingressValueMap = ((MapValue) ingressValues);
            switch (ingressValueMap.getType().getName()) {
                case "HttpApiIngress":
                case "HttpPortIngress":
                case "HttpsPortIngress":
                    processHttpIngress(component, ingressValueMap, key.toString());
                    break;
                case "TCPIngress":
                    processTCPIngress(component, ingressValueMap);
                    break;
                case "GRPCIngress":
                    processGRPCIngress(component, ingressValueMap);
                    break;
                case "WebIngress":
                    processWebIngress(component, ingressValueMap);
                    break;

                default:
                    break;
            }
        });
    }

    private static void processGRPCIngress(ImageComponent component, MapValue attributeMap) {
        GRPC grpc = new GRPC();
        Destination destination = new Destination();
        destination.setPort(Math.toIntExact(attributeMap.getIntValue("backendPort")));
        destination.setHost(component.getService());
        if (attributeMap.containsKey(CelleryConstants.PROTO_FILE)) {
            String protoFile = (attributeMap.getStringValue(CelleryConstants.PROTO_FILE));
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
            grpc.setPort(Math.toIntExact(attributeMap.getIntValue(CelleryConstants.GATEWAY_PORT)));
        }
        component.addPort(port);
        component.addGRPC(grpc);
    }

    private static void processTCPIngress(ImageComponent component, MapValue attributeMap) {
        TCP tcp = new TCP();
        Destination destination = new Destination();
        destination.setHost(component.getService());
        destination.setPort(Math.toIntExact(attributeMap.getIntValue("backendPort")));
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
            tcp.setPort(Math.toIntExact(attributeMap.getIntValue(CelleryConstants.GATEWAY_PORT)));
        }
        component.addPort(port);
        component.addTCP(tcp);
    }

    private static void processHttpIngress(ImageComponent component, MapValue attributeMap, String key) {
        API httpAPI = getApi(component, attributeMap);
        httpAPI.setName(key);
        httpAPI.setPort(CelleryConstants.DEFAULT_GATEWAY_PORT);
        component.setProtocol(CelleryConstants.DEFAULT_GATEWAY_PROTOCOL);
        // Process optional attributes
        if (attributeMap.containsKey(CelleryConstants.CONTEXT)) {
            String context = (attributeMap.getStringValue(CelleryConstants.CONTEXT));
            if (!context.startsWith("/")) {
                context = "/" + context;
            }
            httpAPI.setContext(context);
        }
        if (attributeMap.containsKey("authenticate")) {
            httpAPI.setAuthenticate(attributeMap.getBooleanValue("authenticate"));
        }
        if (attributeMap.containsKey(CelleryConstants.EXPOSE)) {
            if (!httpAPI.isAuthenticate()) {
                component.addUnsecuredPaths(httpAPI.getContext());
            }
            if ("global".equals(attributeMap.getStringValue(CelleryConstants.EXPOSE))) {
                httpAPI.setGlobal(true);
            } else if ("local".equals(attributeMap.getStringValue(CelleryConstants.EXPOSE))) {
                httpAPI.setGlobal(false);
            }
            if (attributeMap.containsKey("definition")) {
                List<APIDefinition> apiDefinitions = new ArrayList<>();
                ArrayValue resourceDefs = attributeMap.getMapValue("definition").getArrayValue("resources");
                IntStream.range(0, resourceDefs.size()).forEach(resourceIndex -> {
                    APIDefinition apiDefinition = new APIDefinition();
                    MapValue definitions = (MapValue) resourceDefs.get(resourceIndex);
                    apiDefinition.setPath(definitions.getStringValue("path"));
                    apiDefinition.setMethod(definitions.getStringValue("method"));
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
     * @param bScalePolicy Scale policy to be processed
     * @param component    current component
     */
    private static void processAutoScalePolicy(MapValue<?, ?> bScalePolicy, ImageComponent component) {
        ScalingPolicy scalingPolicy = new ScalingPolicy();
        scalingPolicy.setReplicas(component.getReplicas());
        boolean bOverridable = false;
        if ("AutoScalingPolicy".equals(bScalePolicy.getType().getName())) {
            // Autoscaling
            image.setAutoScaling(true);
            HPA hpa = new HPA();
            hpa.setMaxReplicas(bScalePolicy.getIntValue(CelleryConstants.MAX_REPLICAS));
            hpa.setMinReplicas(bScalePolicy.getIntValue("minReplicas"));
            bOverridable = (bScalePolicy.getBooleanValue("overridable"));
            MapValue metricsMap = (bScalePolicy.getMapValue("metrics"));
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
                kpa.setMaxReplicas(bScalePolicy.getIntValue(CelleryConstants.MAX_REPLICAS));
            }
            if (bScalePolicy.containsKey(CelleryConstants.CONCURRENCY_TARGET)) {
                kpa.setConcurrency(bScalePolicy.getIntValue(CelleryConstants.CONCURRENCY_TARGET));
            }
            scalingPolicy.setKpa(kpa);
        }
        scalingPolicy.setOverridable(bOverridable);
        component.setScalingPolicy(scalingPolicy);
    }

    private static AutoScalingResourceMetric extractMetrics(MapValue metricsMap, String resourceName) {
        AutoScalingResourceMetric scalingResourceMetric = new AutoScalingResourceMetric();
        scalingResourceMetric.setType(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE);
        Resource resource = new Resource();
        resource.setName(resourceName);
        final MapValue resourceMap = metricsMap.getMapValue(resourceName);
        final String threshold = resourceMap.getStringValue("threshold");
        if (BTYPE_STRING.equals(resourceMap.getType().getName())) {
            HashMap<String, Object> hs = new HashMap<>();
            hs.put("type", "AverageValue");
            hs.put("averageValue", threshold);
            resource.setTarget(hs);
        } else {
            HashMap<String, Object> hs = new HashMap<>();
            hs.put("type", "Utilization");
            hs.put("averageUtilization", Integer.parseInt(threshold));
            resource.setTarget(hs);
        }
        scalingResourceMetric.setResource(resource);
        return scalingResourceMetric;
    }

    private static void generateComposite(Image image) {
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
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES,
                        new Gson().toJson(image.getDependencies()))
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

    private static ComponentSpec getSpec(ImageComponent imageComponent, Component component) {
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

    private static void generateCell(Image image) {
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
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES,
                        new Gson().toJson(image.getDependencies()))
                .build();
        Cell cell = new Cell(objectMeta, cellSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + CelleryConstants.CELLERY + File.separator
                        + image.getCellName() + CelleryConstants.YAML;
        try {
            writeToFile(toYaml(cell), targetPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while writing cell yaml " + targetPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private static Container getContainer(ImageComponent component, List<EnvVar> envVarList) {
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

    private static List<EnvVar> getEnvVars(ImageComponent component) {
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
    private static void generateCellReference() {
        JSONObject json = new JSONObject();
        if (image.isCompositeImage()) {
            image.getComponentNameToComponentMap().forEach((name, component) -> {
                String componentName = getValidName(component.getName());
                json.put(componentName + "_host",
                        CelleryConstants.INSTANCE_NAME_PLACEHOLDER + "--" + componentName + "-service");
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
                        String url = CelleryConstants.DEFAULT_GATEWAY_PROTOCOL + "://" +
                                CelleryConstants.INSTANCE_NAME_PLACEHOLDER +
                                CelleryConstants.GATEWAY_SERVICE + ":"
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
    private static void generateMetadataFile(MapValue<?, ?> components) {
        JSONObject jsonObject = new JSONObject();
        jsonObject.put(KIND, image.isCompositeImage() ? "Composite" : "Cell");
        jsonObject.put(ORG, image.getOrgName());
        jsonObject.put(NAME, image.getCellName());
        jsonObject.put(VERSION, image.getCellVersion());
        jsonObject.put("zeroScalingRequired", image.isZeroScaling());
        jsonObject.put("autoScalingRequired", image.isAutoScaling());

        JSONObject componentsJsonObject = new JSONObject();
        components.forEach((key, componentValue) -> {
            MapValue attributeMap = ((MapValue) componentValue);
            String componentName = attributeMap.getStringValue("name");
            JSONObject componentJson = new JSONObject();

            JSONObject labelsJsonObject = new JSONObject();
            if (attributeMap.containsKey(LABELS)) {
                ((MapValue<?, ?>) attributeMap.getMapValue(LABELS)).forEach((labelKey, labelValue) ->
                        labelsJsonObject.put(labelKey.toString(), labelValue.toString()));
            }
            componentJson.put("labels", labelsJsonObject);

            JSONObject cellDependenciesJsonObject = new JSONObject();
            JSONObject compositeDependenciesJsonObject = new JSONObject();
            JSONArray componentDependenciesJsonArray = new JSONArray();
            if (attributeMap.containsKey(DEPENDENCIES)) {
                MapValue<?, ?> dependencies = attributeMap.getMapValue(DEPENDENCIES);
                if (dependencies.containsKey(CELLS)) {
                    MapValue<?, ?> cellDependencies = dependencies.getMapValue(CELLS);
                    extractDependencies(cellDependenciesJsonObject, cellDependencies, CELL);
                }
                if (dependencies.containsKey(COMPOSITES)) {
                    MapValue<?, ?> compositeDependencies = dependencies.getMapValue(COMPOSITES);
                    extractDependencies(compositeDependenciesJsonObject, compositeDependencies, COMPOSITE);
                }
                if (dependencies.containsKey(COMPONENTS)) {
                    ArrayValue componentsArray = dependencies.getArrayValue(COMPONENTS);
                    IntStream.range(0, componentsArray.size()).forEach(componentIndex -> {
                        MapValue component = ((MapValue) componentsArray.get(componentIndex));
                        componentDependenciesJsonArray.put(component.getStringValue("name"));
                    });
                }
            }
            JSONObject dependenciesJsonObject = new JSONObject();
            dependenciesJsonObject.put(CELLS, cellDependenciesJsonObject);
            dependenciesJsonObject.put(COMPOSITES, compositeDependenciesJsonObject);
            dependenciesJsonObject.put(COMPONENTS, componentDependenciesJsonArray);
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
                OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + METADATA_FILE_NAME;
        try {
            writeToFile(jsonObject.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while generating metadata file " + targetFileNameWithPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private static String getValidRecordName(String name) {
        return name.toLowerCase(Locale.getDefault()).replaceAll("\\P{Alnum}", "");
    }

    private static void extractDependencies(JSONObject dependenciesJsonObject, MapValue<?, ?> cellDependencies,
                                            String kind) {
        cellDependencies.forEach((alias, dependencyValue) -> {
            JSONObject dependencyJsonObject = new JSONObject();
            String org, name, version;
            if (BTYPE_STRING.equals((((MapValue) dependencyValue).getType().getName()))) {
                String dependency = ((MapValue) dependencyValue).stringValue();
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
                MapValue dependency = (MapValue) dependencyValue;
                org = dependency.getStringValue(ORG);
                name = dependency.getStringValue(NAME);
                version = dependency.getStringValue(VERSION);
            }
            dependencyJsonObject.put(ORG, org);
            dependencyJsonObject.put(NAME, name);
            dependencyJsonObject.put(VERSION, version);
            dependencyJsonObject.put("alias", alias.toString());
            dependencyJsonObject.put(KIND, kind);
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
    private static void createDockerImage(String dockerImageTag, String dockerDir) {
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
