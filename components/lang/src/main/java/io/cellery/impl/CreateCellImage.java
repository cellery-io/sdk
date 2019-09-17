/*
 *   Copyright (c) 2018, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  WSO2 Inc. licenses this file to you under the Apache License,
 *  Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package io.cellery.impl;

import com.google.gson.Gson;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.AutoScaling;
import io.cellery.models.AutoScalingResourceMetric;
import io.cellery.models.Cell;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.ComponentSpec;
import io.cellery.models.ComponentTemplate;
import io.cellery.models.Extension;
import io.cellery.models.GRPC;
import io.cellery.models.Gateway;
import io.cellery.models.GatewaySpec;
import io.cellery.models.Ingress;
import io.cellery.models.Resource;
import io.cellery.models.ScalingPolicy;
import io.cellery.models.TCP;
import io.cellery.models.internal.Dependency;
import io.cellery.models.internal.Image;
import io.cellery.models.internal.ImageComponent;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
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
import java.util.LinkedHashMap;
import java.util.List;
import java.util.stream.IntStream;

import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_NAME;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_ORG;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION;
import static io.cellery.CelleryConstants.AUTO_SCALING_METRIC_RESOURCE;
import static io.cellery.CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_CPU;
import static io.cellery.CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_MEMORY;
import static io.cellery.CelleryConstants.BTYPE_STRING;
import static io.cellery.CelleryConstants.CELL;
import static io.cellery.CelleryConstants.CELLS;
import static io.cellery.CelleryConstants.COMPONENTS;
import static io.cellery.CelleryConstants.COMPOSITE;
import static io.cellery.CelleryConstants.COMPOSITES;
import static io.cellery.CelleryConstants.CONCURRENCY_TARGET;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PORT;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PROTOCOL;
import static io.cellery.CelleryConstants.DEPENDENCIES;
import static io.cellery.CelleryConstants.ENV_VARS;
import static io.cellery.CelleryConstants.EXPOSE;
import static io.cellery.CelleryConstants.GATEWAY_PORT;
import static io.cellery.CelleryConstants.GATEWAY_SERVICE;
import static io.cellery.CelleryConstants.IMAGE_SOURCE;
import static io.cellery.CelleryConstants.INGRESSES;
import static io.cellery.CelleryConstants.INSTANCE_NAME_PLACEHOLDER;
import static io.cellery.CelleryConstants.KIND;
import static io.cellery.CelleryConstants.LABELS;
import static io.cellery.CelleryConstants.MAX_REPLICAS;
import static io.cellery.CelleryConstants.METADATA_FILE_NAME;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.POD_RESOURCES;
import static io.cellery.CelleryConstants.PROBES;
import static io.cellery.CelleryConstants.PROTOCOL_GRPC;
import static io.cellery.CelleryConstants.PROTOCOL_TCP;
import static io.cellery.CelleryConstants.PROTO_FILE;
import static io.cellery.CelleryConstants.REFERENCE_FILE_NAME;
import static io.cellery.CelleryConstants.SCALING_POLICY;
import static io.cellery.CelleryConstants.TARGET;
import static io.cellery.CelleryConstants.VERSION;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.copyResourceToTarget;
import static io.cellery.CelleryUtils.getApi;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.processEnvVars;
import static io.cellery.CelleryUtils.processProbes;
import static io.cellery.CelleryUtils.processResources;
import static io.cellery.CelleryUtils.processWebIngress;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

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
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir") + File.separator + TARGET;
    private static final Logger log = LoggerFactory.getLogger(CreateCellImage.class);

    private Image image = new Image();

    public void execute(Context ctx) {
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        image.setOrgName(((BString) nameStruct.get(ORG)).stringValue());
        image.setCellName(((BString) nameStruct.get(NAME)).stringValue());
        image.setCellVersion(((BString) nameStruct.get(VERSION)).stringValue());
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        image.setCompositeImage("Composite".equals(refArgument.getType().getName()));
        LinkedHashMap<?, ?> components = ((BMap) refArgument.getMap().get("components")).getMap();
        try {
            processComponents(components);
            generateCellReference();
            generateMetadataFile(components);
            if (image.isCompositeImage()) {
//                generateComposite(image);
            } else {
                generateCell(image);
            }
        } catch (BallerinaException e) {
            ctx.setReturnValues(BLangVMErrors.createError(ctx, e.getMessage()));
        }
    }

    private void processComponents(LinkedHashMap<?, ?> components) {
        components.forEach((componentKey, componentValue) -> {
            ImageComponent component = new ImageComponent();
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            // Set mandatory fields.
            component.setName(((BString) attributeMap.get("name")).stringValue());
            component.setReplicas((int) (((BInteger) attributeMap.get("replicas")).intValue()));
            component.setService(component.getName());
            component.setType(((BString) attributeMap.get("type")).stringValue());
            processSource(component, attributeMap);
            //Process Optional fields
            if (attributeMap.containsKey(INGRESSES)) {
                processIngress(((BMap<?, ?>) attributeMap.get(INGRESSES)).getMap(), component);
            }
            if (attributeMap.containsKey(LABELS)) {
                ((BMap<?, ?>) attributeMap.get(LABELS)).getMap().forEach((labelKey, labelValue) ->
                        component.addLabel(labelKey.toString(), labelValue.toString()));
            }
            if (attributeMap.containsKey(SCALING_POLICY)) {
                processAutoScalePolicy(((BMap<?, ?>) attributeMap.get(SCALING_POLICY)), component);
            }
            if (attributeMap.containsKey(ENV_VARS)) {
                processEnvVars(((BMap<?, ?>) attributeMap.get(ENV_VARS)).getMap(), component);
            }
            if (attributeMap.containsKey(PROBES)) {
                processProbes(((BMap<?, ?>) attributeMap.get(PROBES)).getMap(), component);
            }
            if (attributeMap.containsKey(POD_RESOURCES)) {
                processResources(((BMap<?, ?>) attributeMap.get(POD_RESOURCES)).getMap(), component);
            }
            image.addComponent(component);
        });
    }

    private void processSource(ImageComponent component, LinkedHashMap attributeMap) {
        if ("ImageSource".equals(((BValue) attributeMap.get(IMAGE_SOURCE)).getType().getName())) {
            //Image Source
            component.setSource(((BString) ((BMap) attributeMap.get(IMAGE_SOURCE)).getMap()
                    .get("image")).stringValue());
        } else {
            // Docker Source
            LinkedHashMap dockerSourceMap = ((BMap) attributeMap.get(IMAGE_SOURCE)).getMap();
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
                    processHttpIngress(component, attributeMap);
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
        if (attributeMap.containsKey(GATEWAY_PORT)) {
            grpc.setPort((int) ((BInteger) attributeMap.get(GATEWAY_PORT)).intValue());
        }
        grpc.setBackendPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        if (attributeMap.containsKey(PROTO_FILE)) {
            String protoFile = ((BString) attributeMap.get(PROTO_FILE)).stringValue();
            if (!protoFile.isEmpty()) {
                copyResourceToTarget(protoFile);
            }
        }
        component.setProtocol(PROTOCOL_GRPC);
        component.setContainerPort(grpc.getBackendPort());
        grpc.setBackendHost(component.getService());
        component.addGRPC(grpc);
    }

    private void processTCPIngress(ImageComponent component, LinkedHashMap attributeMap) {
        TCP tcp = new TCP();
        if (attributeMap.containsKey(GATEWAY_PORT)) {
            tcp.setPort((int) ((BInteger) attributeMap.get(GATEWAY_PORT)).intValue());
        }
        tcp.setBackendPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        component.setProtocol(PROTOCOL_TCP);
        component.setContainerPort(tcp.getBackendPort());
        tcp.setBackendHost(component.getService());
        component.addTCP(tcp);
    }

    private void processHttpIngress(ImageComponent component, LinkedHashMap attributeMap) {
        API httpAPI = getApi(component, attributeMap);
        component.setProtocol(DEFAULT_GATEWAY_PROTOCOL);
        // Process optional attributes
        if (attributeMap.containsKey("context")) {
            httpAPI.setContext(((BString) attributeMap.get("context")).stringValue());
        }

        if (attributeMap.containsKey(EXPOSE)) {
            httpAPI.setAuthenticate(((BBoolean) attributeMap.get("authenticate")).booleanValue());
            if (!httpAPI.isAuthenticate()) {
                String context = httpAPI.getContext();
                if (!context.startsWith("/")) {
                    context = "/" + context;
                }
                component.addUnsecuredPaths(context);
            }
            if ("global".equals(((BString) attributeMap.get(EXPOSE)).stringValue())) {
                httpAPI.setGlobal(true);
                httpAPI.setBackend(component.getService());
            } else if ("local".equals(((BString) attributeMap.get(EXPOSE)).stringValue())) {
                httpAPI.setGlobal(false);
                httpAPI.setBackend(component.getService());
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
        LinkedHashMap bScalePolicy = scalePolicy.getMap();
        boolean bOverridable = false;
        if ("AutoScalingPolicy".equals(scalePolicy.getType().getName())) {
            // Autoscaling
            image.setAutoScaling(true);
            scalingPolicy.setMinReplicas(((BInteger) (bScalePolicy.get("minReplicas"))).intValue());
            scalingPolicy.setMaxReplicas(((BInteger) bScalePolicy.get(MAX_REPLICAS)).intValue());
            bOverridable = ((BBoolean) bScalePolicy.get("overridable")).booleanValue();
            LinkedHashMap metricsMap = ((BMap) bScalePolicy.get("metrics")).getMap();
            if (metricsMap.containsKey(AUTO_SCALING_METRIC_RESOURCE_CPU)) {
                extractMetrics(scalingPolicy, metricsMap, AUTO_SCALING_METRIC_RESOURCE_CPU);
            }
            if (metricsMap.containsKey(AUTO_SCALING_METRIC_RESOURCE_MEMORY)) {
                extractMetrics(scalingPolicy, metricsMap, AUTO_SCALING_METRIC_RESOURCE_MEMORY);
            }

        } else {
            //Zero Scaling
            image.setZeroScaling(true);
            scalingPolicy.setMinReplicas(0);
            if (bScalePolicy.containsKey(MAX_REPLICAS)) {
                scalingPolicy.setMaxReplicas(((BInteger) bScalePolicy.get(MAX_REPLICAS)).intValue());
            }
            if (bScalePolicy.containsKey(CONCURRENCY_TARGET)) {
                scalingPolicy.setConcurrency(((BInteger) bScalePolicy.get(CONCURRENCY_TARGET)).intValue());
            }
        }
        component.setAutoscaling(new AutoScaling(scalingPolicy, bOverridable));
    }

    private void extractMetrics(ScalingPolicy scalingPolicy, LinkedHashMap metricsMap,
                                String autoScalingMetricResourceMemory) {
        AutoScalingResourceMetric scalingResourceMetric = new AutoScalingResourceMetric();
        scalingResourceMetric.setType(AUTO_SCALING_METRIC_RESOURCE);
        Resource resource = new Resource();
        resource.setName(autoScalingMetricResourceMemory);
        scalingResourceMetric.setResource(resource);
        scalingPolicy.addAutoScalingResourceMetric(scalingResourceMetric);
        final BValue bValue = (BValue) ((BMap) metricsMap.get(autoScalingMetricResourceMemory)).getMap()
                .get("threshold");
        if (BTYPE_STRING.equals(bValue.getType().getName())) {
            resource.setTargetAverageValue(bValue.stringValue());
        } else {
            resource.setTargetAverageUtilization((int) ((BInteger) bValue).intValue());
        }
    }

//    private void generateComposite(Image image) {
//        List<ImageComponent> components =
//                new ArrayList<>(image.getComponentNameToComponentMap().values());
//        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
//        for (ImageComponent component : components) {
//            ServiceTemplateSpec templateSpec = getServiceTemplateSpec(new GatewaySpec(), component);
//            buildTemplateSpec(serviceTemplateList, component, templateSpec);
//        }
//
//        CompositeSpec compositeSpec = new CompositeSpec();
//        compositeSpec.setServicesTemplates(serviceTemplateList);
//        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(image.getCellName()))
//                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, image.getOrgName())
//                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, image.getCellName())
//                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, image.getCellVersion())
//                .addToAnnotations(ANNOTATION_CELL_IMAGE_DEPENDENCIES, new Gson().toJson(image.getDependencies()))
//                .build();
//        Composite composite = new Composite(objectMeta, compositeSpec);
//        String targetPath =
//                OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + image.getCellName() + YAML;
//        try {
//            writeToFile(toYaml(composite), targetPath);
//        } catch (IOException e) {
//            String errMsg = "Error occurred while writing composite yaml " + targetPath;
//            log.error(errMsg, e);
//            throw new BallerinaException(errMsg);
//        }
//    }

//    private void buildTemplateSpec(List<ServiceTemplate> serviceTemplateList, ImageComponent component,
//                                   ServiceTemplateSpec templateSpec) {
//        templateSpec.setReplicas(component.getReplicas());
//        templateSpec.setProtocol(component.getProtocol());
//        List<EnvVar> envVarList = getEnvVars(component);
//        templateSpec.setContainer(getContainer(component, envVarList));
//        AutoScaling autoScaling = component.getAutoscaling();
//        if (autoScaling != null) {
//            templateSpec.setAutoscaling(autoScaling);
//        }
//        ServiceTemplate serviceTemplate = new ServiceTemplate();
//        serviceTemplate.setMetadata(new ObjectMetaBuilder()
//                .withName(component.getService())
//                .withLabels(component.getLabels())
//                .build());
//        serviceTemplate.setSpec(templateSpec);
//        serviceTemplateList.add(serviceTemplate);
//    }

    private void generateCell(Image image) {
        List<ImageComponent> components =
                new ArrayList<>(image.getComponentNameToComponentMap().values());
//        List<String> unsecuredPaths = new ArrayList<>();
//        STSTemplate stsTemplate = new STSTemplate();
//        STSTemplateSpec stsTemplateSpec = new STSTemplateSpec();
        Ingress ingress = new Ingress();
        Extension extension = new Extension();
        CellSpec cellSpec = new CellSpec();
        components.forEach(imageComponent -> {
            ingress.addHttpAPI(imageComponent.getApis());
            ingress.addGRPC(imageComponent.getGrpcList());
            ingress.addTCP(imageComponent.getTcpList());
            if (imageComponent.getWebList().size() > 0) {
                extension.setOidc(imageComponent.getWebList().get(0).getOidc());
            }
            Component component = new Component();
            component.setMetadata(new ObjectMetaBuilder().withName(imageComponent.getName()).build());
            ComponentTemplate componentTemplate = new ComponentTemplate();
            componentTemplate.addContainer(getContainer(imageComponent, getEnvVars(imageComponent)));
            ComponentSpec componentSpec = new ComponentSpec();
            componentSpec.setTemplate(componentTemplate);
            component.setSpec(componentSpec);
            cellSpec.addComponent(component);
        });

        Gateway gateway = new Gateway();
        GatewaySpec gatewaySpec = new GatewaySpec();
        gatewaySpec.setIngress(ingress);
        gateway.setSpec(gatewaySpec);

        ingress.setExtensions(extension);
        cellSpec.setGateway(gateway);

//        stsTemplateSpec.setUnsecuredPaths(unsecuredPaths);
//        stsTemplate.setSpec(stsTemplateSpec);
//        Gateway gateway = new Gateway();
//        gateway.setSpec(gatewaySpec);
//
//        CellSpec cellSpec = new CellSpec();
//        cellSpec.setGateway(gateway);
//        cellSpec.setServicesTemplates(serviceTemplateList);
//        cellSpec.setStsTemplate(stsTemplate);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(image.getCellName()))
                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, image.getOrgName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, image.getCellName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, image.getCellVersion())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_DEPENDENCIES, new Gson().toJson(image.getDependencies()))
                .build();
        Cell cell = new Cell(objectMeta, cellSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + image.getCellName() + YAML;
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

//    private ServiceTemplateSpec getServiceTemplateSpec(GatewaySpec gatewaySpec, ImageComponent component) {
//        ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();
//        if (component.getWebList().size() > 0) {
//            templateSpec.setServicePort(DEFAULT_GATEWAY_PORT);
//            gatewaySpec.setType(ENVOY_GATEWAY);
//            // Only Single web ingress is supported for 0.2.0
//            // Therefore we only process the 0th element
//            Web webIngress = component.getWebList().get(0);
//            gatewaySpec.addHttpAPI(Collections.singletonList(webIngress.getHttpAPI()));
//            gatewaySpec.setHost(webIngress.getVhost());
//            gatewaySpec.setOidc(webIngress.getOidc());
//        } else if (component.getApis().size() > 0) {
//            // HTTP ingress
//            templateSpec.setServicePort(DEFAULT_GATEWAY_PORT);
//            gatewaySpec.setType(MICRO_GATEWAY);
//            gatewaySpec.addHttpAPI(component.getApis());
//        } else if (component.getTcpList().size() > 0) {
//            // Only Single TCP ingress is supported for 0.2.0
//            // Therefore we only process the 0th element
//            gatewaySpec.setType(ENVOY_GATEWAY);
//            if (component.getTcpList().get(0).getPort() != 0) {
//                gatewaySpec.addTCP(component.getTcpList());
//            }
//            templateSpec.setServicePort(component.getTcpList().get(0).getBackendPort());
//        } else if (component.getGrpcList().size() > 0) {
//            gatewaySpec.setType(ENVOY_GATEWAY);
//            templateSpec.setServicePort(component.getGrpcList().get(0).getBackendPort());
//            if (component.getGrpcList().get(0).getPort() != 0) {
//                gatewaySpec.addGRPC(component.getGrpcList());
//            }
//        }
//        return templateSpec;
//    }

    /**
     * Generate a Cell Reference that can be used by other cells.
     */
    private void generateCellReference() {
        JSONObject json = new JSONObject();
        if (image.isCompositeImage()) {
            image.getComponentNameToComponentMap().forEach((componentName, component) -> {
                json.put(componentName + "_host", INSTANCE_NAME_PLACEHOLDER + "--" + componentName + "-service");
                if (component.getApis().size() > 0) {
                    json.put(componentName + "_port", DEFAULT_GATEWAY_PORT);
                }
                component.getTcpList().forEach(tcp -> json.put(componentName + "_tcp_port", tcp.getPort()));
                component.getGrpcList().forEach(grpc -> json.put(componentName + "_grpc_port", grpc.getPort()));
            });
        } else {
            image.getComponentNameToComponentMap().forEach((componentName, component) -> {
                component.getApis().forEach(api -> {
                    String context = api.getContext();
                    if (StringUtils.isNotEmpty(context)) {
                        String url =
                                DEFAULT_GATEWAY_PROTOCOL + "://" + INSTANCE_NAME_PLACEHOLDER + GATEWAY_SERVICE + ":"
                                        + DEFAULT_GATEWAY_PORT + "/" + context;
                        if ("/".equals(context)) {
                            json.put(componentName + "_api_url", url.replaceAll("(?<!http:)//", "/"));
                        } else {
                            json.put(context + "_api_url", url.replaceAll("(?<!http:)//", "/"));
                        }
                    }
                });
                component.getTcpList().forEach(tcp -> json.put(componentName + "_tcp_port", tcp.getPort()));
                component.getGrpcList().forEach(grpc -> json.put(componentName + "_grpc_port", grpc.getPort()));
            });
            json.put("gateway_host", INSTANCE_NAME_PLACEHOLDER + GATEWAY_SERVICE);
        }
        String targetFileNameWithPath =
                OUTPUT_DIRECTORY + File.separator + "ref" + File.separator + REFERENCE_FILE_NAME;
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
        jsonObject.put(KIND, image.isCompositeImage() ? "Composite" : "Cell");
        jsonObject.put(ORG, image.getOrgName());
        jsonObject.put(NAME, image.getCellName());
        jsonObject.put(VERSION, image.getCellVersion());
        jsonObject.put("zeroScalingRequired", image.isZeroScaling());
        jsonObject.put("autoScalingRequired", image.isAutoScaling());

        JSONObject componentsJsonObject = new JSONObject();
        components.forEach((key, componentValue) -> {
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            String componentName = ((BString) attributeMap.get("name")).stringValue();
            JSONObject componentJson = new JSONObject();

            JSONObject labelsJsonObject = new JSONObject();
            if (attributeMap.containsKey(LABELS)) {
                ((BMap<?, ?>) attributeMap.get(LABELS)).getMap().forEach((labelKey, labelValue) ->
                        labelsJsonObject.put(labelKey.toString(), labelValue.toString()));
            }
            componentJson.put("labels", labelsJsonObject);

            JSONObject cellDependenciesJsonObject = new JSONObject();
            JSONObject compositeDependenciesJsonObject = new JSONObject();
            JSONArray componentDependenciesJsonArray = new JSONArray();
            if (attributeMap.containsKey(DEPENDENCIES)) {
                LinkedHashMap<?, ?> dependencies = ((BMap<?, ?>) attributeMap.get(DEPENDENCIES)).getMap();
                if (dependencies.containsKey(CELLS)) {
                    LinkedHashMap<?, ?> cellDependencies = ((BMap) dependencies.get(CELLS)).getMap();
                    extractDependencies(cellDependenciesJsonObject, cellDependencies, CELL);
                }
                if (dependencies.containsKey(COMPOSITES)) {
                    LinkedHashMap<?, ?> compositeDependencies = ((BMap) dependencies.get(COMPOSITES)).getMap();
                    extractDependencies(compositeDependenciesJsonObject, compositeDependencies, COMPOSITE);
                }
                if (dependencies.containsKey(COMPONENTS)) {
                    BValueArray componentsArray = ((BValueArray) dependencies.get(COMPONENTS));
                    IntStream.range(0, (int) componentsArray.size()).forEach(componentIndex -> {
                        LinkedHashMap component = ((BMap) componentsArray.getBValue(componentIndex)).getMap();
                        componentDependenciesJsonArray.put(((BString) component.get("name")).stringValue());
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

    private void extractDependencies(JSONObject dependenciesJsonObject, LinkedHashMap<?, ?> cellDependencies,
                                     String kind) {
        cellDependencies.forEach((alias, dependencyValue) -> {
            JSONObject dependencyJsonObject = new JSONObject();
            String org, name, version;
            if (BTYPE_STRING.equals(((BValue) dependencyValue).getType().getName())) {
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
                org = ((BString) dependency.get(ORG)).stringValue();
                name = ((BString) dependency.get(NAME)).stringValue();
                version = ((BString) dependency.get(VERSION)).stringValue();
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
