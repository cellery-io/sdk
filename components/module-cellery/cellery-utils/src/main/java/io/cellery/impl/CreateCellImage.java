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
import io.cellery.BallerinaCelleryException;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.AutoScaling;
import io.cellery.models.AutoScalingPolicy;
import io.cellery.models.AutoScalingResourceMetric;
import io.cellery.models.Cell;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.Composite;
import io.cellery.models.CompositeSpec;
import io.cellery.models.Dependency;
import io.cellery.models.GRPC;
import io.cellery.models.GatewaySpec;
import io.cellery.models.GatewayTemplate;
import io.cellery.models.Image;
import io.cellery.models.Resource;
import io.cellery.models.STSTemplate;
import io.cellery.models.STSTemplateSpec;
import io.cellery.models.ServiceTemplate;
import io.cellery.models.ServiceTemplateSpec;
import io.cellery.models.TCP;
import io.cellery.models.Web;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.jvm.values.ObjectValue;
import org.ballerinalang.model.values.BInteger;
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
import java.util.HashSet;
import java.util.List;
import java.util.Set;
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
import static io.cellery.CelleryConstants.ENVOY_GATEWAY;
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
import static io.cellery.CelleryConstants.MICRO_GATEWAY;
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
public class CreateCellImage {
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir") + File.separator + TARGET;
    private static final Logger log = LoggerFactory.getLogger(CreateCellImage.class);

    private static Image image = new Image();
    private static Set<String> exposedComponents = new HashSet<>();

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
                generateCell(image);
            }
        } catch (BallerinaException e) {
            throw new BallerinaCelleryException(e.getMessage());
        }
    }

    private static void processComponents(MapValue<?, ?> components) {
        components.forEach((componentKey, componentValue) -> {
            Component component = new Component();
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
            image.addComponent(component);
        });
    }

    private static void processSource(Component component, MapValue attributeMap) {
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
    private static void processIngress(MapValue<?, ?> ingressMap, Component component) {
        ingressMap.forEach((key, ingressValues) -> {
            MapValue ingressValueMap = ((MapValue) ingressValues);
            switch (ingressValueMap.getType().getName()) {
                case "HttpApiIngress":
                case "HttpPortIngress":
                case "HttpsPortIngress":
                    processHttpIngress(component, ingressValueMap);
                    break;
                case "TCPIngress":
                    processTCPIngress(component, ingressValueMap);
                    break;
                case "GRPCIngress":
                    processGRPCIngress(component, ingressValueMap);
                    break;
                case "WebIngress":
                    processWebIngress(component, ingressValueMap);
                    exposedComponents.add(component.getName());
                    break;

                default:
                    break;
            }
        });
    }

    private static void processGRPCIngress(Component component, MapValue attributeMap) {
        GRPC grpc = new GRPC();
        if (attributeMap.containsKey(GATEWAY_PORT)) {
            grpc.setPort(Math.toIntExact(attributeMap.getIntValue(GATEWAY_PORT)));
            // Component is exposed via ingress
            exposedComponents.add(component.getName());
        }
        grpc.setBackendPort(Math.toIntExact(attributeMap.getIntValue("backendPort")));
        if (attributeMap.containsKey(PROTO_FILE)) {
            String protoFile = attributeMap.getStringValue(PROTO_FILE);
            if (!protoFile.isEmpty()) {
                copyResourceToTarget(protoFile);
            }
        }
        component.setProtocol(PROTOCOL_GRPC);
        component.setContainerPort(grpc.getBackendPort());
        grpc.setBackendHost(component.getService());
        component.addGRPC(grpc);
    }

    private static void processTCPIngress(Component component, MapValue attributeMap) {
        TCP tcp = new TCP();
        if (attributeMap.containsKey(GATEWAY_PORT)) {
            tcp.setPort(Math.toIntExact(attributeMap.getIntValue(GATEWAY_PORT)));
            // Component is exposed via ingress
            exposedComponents.add(component.getName());
        }
        tcp.setBackendPort(Math.toIntExact(attributeMap.getIntValue("backendPort")));
        component.setProtocol(PROTOCOL_TCP);
        component.setContainerPort(tcp.getBackendPort());
        tcp.setBackendHost(component.getService());
        component.addTCP(tcp);
    }

    private static void processHttpIngress(Component component, MapValue attributeMap) {
        API httpAPI = getApi(component, attributeMap);
        component.setProtocol(DEFAULT_GATEWAY_PROTOCOL);
        // Process optional attributes
        if (attributeMap.containsKey("context")) {
            httpAPI.setContext(attributeMap.getStringValue("context"));
        }

        if (attributeMap.containsKey(EXPOSE)) {
            httpAPI.setAuthenticate(attributeMap.getBooleanValue("authenticate"));
            if (!httpAPI.isAuthenticate()) {
                String context = httpAPI.getContext();
                if (!context.startsWith("/")) {
                    context = "/" + context;
                }
                component.addUnsecuredPaths(context);
            }
            if ("global".equals(attributeMap.getStringValue(EXPOSE))) {
                httpAPI.setGlobal(true);
                httpAPI.setBackend(component.getService());
                exposedComponents.add(component.getName());
            } else if ("local".equals(attributeMap.getStringValue(EXPOSE))) {
                httpAPI.setGlobal(false);
                httpAPI.setBackend(component.getService());
                // Component is exposed via ingress
                exposedComponents.add(component.getName());
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
        component.addApi(httpAPI);
    }


    /**
     * Extract the scale policy.
     *
     * @param scalePolicy Scale policy to be processed
     * @param component   current component
     */
    private static void processAutoScalePolicy(MapValue scalePolicy, Component component) {
        AutoScalingPolicy autoScalingPolicy = new AutoScalingPolicy();
        boolean bOverridable = false;
        if ("AutoScalingPolicy".equals(scalePolicy.getType().getName())) {
            // Autoscaling
            image.setAutoScaling(true);
            autoScalingPolicy.setMinReplicas(scalePolicy.getIntValue("minReplicas"));
            autoScalingPolicy.setMaxReplicas(scalePolicy.getIntValue(MAX_REPLICAS));
            bOverridable = scalePolicy.getBooleanValue("overridable");
            MapValue metricsMap = (scalePolicy.getMapValue("metrics"));
            if (metricsMap.containsKey(AUTO_SCALING_METRIC_RESOURCE_CPU)) {
                extractMetrics(autoScalingPolicy, metricsMap, AUTO_SCALING_METRIC_RESOURCE_CPU);
            }
            if (metricsMap.containsKey(AUTO_SCALING_METRIC_RESOURCE_MEMORY)) {
                extractMetrics(autoScalingPolicy, metricsMap, AUTO_SCALING_METRIC_RESOURCE_MEMORY);
            }

        } else {
            //Zero Scaling
            image.setZeroScaling(true);
            autoScalingPolicy.setMinReplicas(0);
            if (scalePolicy.containsKey(MAX_REPLICAS)) {
                autoScalingPolicy.setMaxReplicas(((BInteger) scalePolicy.get(MAX_REPLICAS)).intValue());
            }
            if (scalePolicy.containsKey(CONCURRENCY_TARGET)) {
                autoScalingPolicy.setConcurrency(((BInteger) scalePolicy.get(CONCURRENCY_TARGET)).intValue());
            }
        }
        component.setAutoscaling(new AutoScaling(autoScalingPolicy, bOverridable));
    }

    private static void extractMetrics(AutoScalingPolicy autoScalingPolicy, MapValue metricsMap,
                                       String autoScalingMetricResourceMemory) {
        AutoScalingResourceMetric scalingResourceMetric = new AutoScalingResourceMetric();
        scalingResourceMetric.setType(AUTO_SCALING_METRIC_RESOURCE);
        Resource resource = new Resource();
        resource.setName(autoScalingMetricResourceMemory);
        scalingResourceMetric.setResource(resource);
        autoScalingPolicy.addAutoScalingResourceMetric(scalingResourceMetric);
        final ObjectValue bValue = metricsMap.getMapValue(autoScalingMetricResourceMemory).getObjectValue("threshold");
        if (BTYPE_STRING.equals(bValue.getType().getName())) {
            resource.setTargetAverageValue(bValue.stringValue());
        } else {
            resource.setTargetAverageUtilization(Integer.parseInt(bValue.stringValue()));
        }
    }

    private static void generateComposite(Image image) {
        List<Component> components =
                new ArrayList<>(image.getComponentNameToComponentMap().values());
        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
        for (Component component : components) {
            ServiceTemplateSpec templateSpec = getServiceTemplateSpec(new GatewaySpec(), component);
            buildTemplateSpec(serviceTemplateList, component, templateSpec);
        }

        CompositeSpec compositeSpec = new CompositeSpec();
        compositeSpec.setServicesTemplates(serviceTemplateList);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(image.getCellName()))
                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, image.getOrgName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, image.getCellName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, image.getCellVersion())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_DEPENDENCIES, new Gson().toJson(image.getDependencies()))
                .build();
        Composite composite = new Composite(objectMeta, compositeSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + image.getCellName() + YAML;
        try {
            writeToFile(toYaml(composite), targetPath);
        } catch (IOException e) {
            String errMsg = "Error occurred while writing composite yaml " + targetPath;
            log.error(errMsg, e);
            throw new BallerinaException(errMsg);
        }
    }

    private static void buildTemplateSpec(List<ServiceTemplate> serviceTemplateList, Component component,
                                          ServiceTemplateSpec templateSpec) {
        templateSpec.setReplicas(component.getReplicas());
        templateSpec.setProtocol(component.getProtocol());
        List<EnvVar> envVarList = getEnvVars(component);
        templateSpec.setContainer(getContainer(component, envVarList));
        AutoScaling autoScaling = component.getAutoscaling();
        if (autoScaling != null) {
            templateSpec.setAutoscaling(autoScaling);
        }
        ServiceTemplate serviceTemplate = new ServiceTemplate();
        serviceTemplate.setMetadata(new ObjectMetaBuilder()
                .withName(component.getService())
                .withLabels(component.getLabels())
                .build());
        serviceTemplate.setSpec(templateSpec);
        serviceTemplateList.add(serviceTemplate);
    }

    private static void generateCell(Image image) {
        List<Component> components =
                new ArrayList<>(image.getComponentNameToComponentMap().values());
        GatewaySpec gatewaySpec = new GatewaySpec();
        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
        List<String> unsecuredPaths = new ArrayList<>();
        STSTemplate stsTemplate = new STSTemplate();
        STSTemplateSpec stsTemplateSpec = new STSTemplateSpec();
        for (Component component : components) {
            ServiceTemplateSpec templateSpec = getServiceTemplateSpec(gatewaySpec, component);
            unsecuredPaths.addAll(component.getUnsecuredPaths());
            buildTemplateSpec(serviceTemplateList, component, templateSpec);
        }
        stsTemplateSpec.setUnsecuredPaths(unsecuredPaths);
        stsTemplate.setSpec(stsTemplateSpec);
        GatewayTemplate gatewayTemplate = new GatewayTemplate();
        gatewayTemplate.setSpec(gatewaySpec);

        CellSpec cellSpec = new CellSpec();
        cellSpec.setGatewayTemplate(gatewayTemplate);
        cellSpec.setServicesTemplates(serviceTemplateList);
        cellSpec.setStsTemplate(stsTemplate);
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

    private static Container getContainer(Component component, List<EnvVar> envVarList) {
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

    private static List<EnvVar> getEnvVars(Component component) {
        List<EnvVar> envVarList = new ArrayList<>();
        component.getEnvVars().forEach((key, value) -> {
            if (StringUtils.isEmpty(value)) {
                printWarning("Value is empty for environment variable \"" + key + "\"");
            }
            envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
        });
        return envVarList;
    }

    private static ServiceTemplateSpec getServiceTemplateSpec(GatewaySpec gatewaySpec, Component component) {
        ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();
        if (component.getWebList().size() > 0) {
            templateSpec.setServicePort(DEFAULT_GATEWAY_PORT);
            gatewaySpec.setType(ENVOY_GATEWAY);
            // Only Single web ingress is supported for 0.2.0
            // Therefore we only process the 0th element
            Web webIngress = component.getWebList().get(0);
            gatewaySpec.addHttpAPI(Collections.singletonList(webIngress.getHttpAPI()));
            gatewaySpec.setHost(webIngress.getVhost());
            gatewaySpec.setOidc(webIngress.getOidc());
        } else if (component.getApis().size() > 0) {
            // HTTP ingress
            templateSpec.setServicePort(DEFAULT_GATEWAY_PORT);
            gatewaySpec.setType(MICRO_GATEWAY);
            gatewaySpec.addHttpAPI(component.getApis());
        } else if (component.getTcpList().size() > 0) {
            // Only Single TCP ingress is supported for 0.2.0
            // Therefore we only process the 0th element
            gatewaySpec.setType(ENVOY_GATEWAY);
            if (component.getTcpList().get(0).getPort() != 0) {
                gatewaySpec.addTCP(component.getTcpList());
            }
            templateSpec.setServicePort(component.getTcpList().get(0).getBackendPort());
        } else if (component.getGrpcList().size() > 0) {
            gatewaySpec.setType(ENVOY_GATEWAY);
            templateSpec.setServicePort(component.getGrpcList().get(0).getBackendPort());
            if (component.getGrpcList().get(0).getPort() != 0) {
                gatewaySpec.addGRPC(component.getGrpcList());
            }
        }
        return templateSpec;
    }

    /**
     * Generate a Cell Reference that can be used by other cells.
     */
    private static void generateCellReference() {
        JSONObject json = new JSONObject();
        if (image.isCompositeImage()) {
            image.getComponentNameToComponentMap().forEach((componentName, component) -> {
                json.put(componentName + "_host", INSTANCE_NAME_PLACEHOLDER + "-" + componentName + "--service");
                json.put(componentName + "_port", component.getContainerPort());
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

            componentJson.put("exposed", exposedComponents.contains(componentName));
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
