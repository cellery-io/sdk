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

import io.cellery.CelleryConstants;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.AutoScaling;
import io.cellery.models.AutoScalingPolicy;
import io.cellery.models.AutoScalingResourceMetric;
import io.cellery.models.AutoScalingSpec;
import io.cellery.models.Cell;
import io.cellery.models.CellImage;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.GRPC;
import io.cellery.models.GatewaySpec;
import io.cellery.models.GatewayTemplate;
import io.cellery.models.ServiceTemplate;
import io.cellery.models.ServiceTemplateSpec;
import io.cellery.models.TCP;
import io.cellery.models.Web;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.HorizontalPodAutoscalerSpecBuilder;
import io.fabric8.kubernetes.api.model.MetricSpecBuilder;
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
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.stream.IntStream;

import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_NAME;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_ORG;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PORT;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PROTOCOL;
import static io.cellery.CelleryConstants.ENVOY_GATEWAY;
import static io.cellery.CelleryConstants.INSTANCE_NAME_PLACEHOLDER;
import static io.cellery.CelleryConstants.MICRO_GATEWAY;
import static io.cellery.CelleryConstants.PROTOCOL_GRPC;
import static io.cellery.CelleryConstants.PROTOCOL_TCP;
import static io.cellery.CelleryConstants.REFERENCE_FILE_NAME;
import static io.cellery.CelleryConstants.TARGET;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.copyResourceToTarget;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.processEnvVars;
import static io.cellery.CelleryUtils.processOidc;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * Native function cellery:createImage.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createImage",
        args = {@Argument(name = "cellImage", type = TypeKind.RECORD),
                @Argument(name = "iName", type = TypeKind.RECORD)},
        returnType = {@ReturnType(type = TypeKind.BOOLEAN), @ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class CreateImage extends BlockingNativeCallableUnit {
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir") + File.separator + TARGET;

    private CellImage cellImage = new CellImage();

    public void execute(Context ctx) {
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        cellImage.setOrgName(((BString) nameStruct.get("org")).stringValue());
        cellImage.setCellName(((BString) nameStruct.get("name")).stringValue());
        cellImage.setCellVersion(((BString) nameStruct.get("ver")).stringValue());
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        LinkedHashMap<?, ?> components = ((BMap) refArgument.getMap().get("components")).getMap();
        try {
            processComponents(components);
            generateCell();
            generateCellReference();
        } catch (BallerinaException e) {
            ctx.setReturnValues(BLangVMErrors.createError(ctx, e.getMessage()));
            return;
        }
        ctx.setReturnValues(new BBoolean(true));
    }

    private void processComponents(LinkedHashMap<?, ?> components) {
        components.forEach((componentKey, componentValue) -> {
            Component component = new Component();
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            // Set mandatory fields.
            component.setName(((BString) attributeMap.get("name")).stringValue());
            component.setReplicas((int) (((BInteger) attributeMap.get("replicas")).intValue()));
            component.setService(component.getName());
            component.setSource(((BString) ((BMap) attributeMap.get("source")).getMap().get("image")).stringValue());

            //Process Optional fields
            if (attributeMap.containsKey("ingresses")) {
                processIngress(((BMap<?, ?>) attributeMap.get("ingresses")).getMap(), component);
            }
            if (attributeMap.containsKey("labels")) {
                ((BMap<?, ?>) attributeMap.get("labels")).getMap().forEach((labelKey, labelValue) ->
                        component.addLabel(labelKey.toString(), labelValue.toString()));
            }
            if (attributeMap.containsKey("autoscaling")) {
                processAutoScalePolicy(((BMap<?, ?>) attributeMap.get("autoscaling")).getMap(), component);
            }
            if (attributeMap.containsKey("dependencies")) {
                generateDependenciesFile(((BMap<?, ?>) attributeMap.get("dependencies")).getMap());
            }
            if (attributeMap.containsKey("envVars")) {
                processEnvVars(((BMap<?, ?>) attributeMap.get("envVars")).getMap(), component);
            }
            cellImage.addComponent(component);
        });
    }

    /**
     * Extract the ingresses.
     *
     * @param ingressMap list of ingresses defined
     * @param component  current component
     */
    private void processIngress(LinkedHashMap<?, ?> ingressMap, Component component) {
        ingressMap.forEach((key, ingressValues) -> {
            BMap ingressValueMap = ((BMap) ingressValues);
            LinkedHashMap attributeMap = ingressValueMap.getMap();
            switch (ingressValueMap.getType().getName()) {
                case "HttpApiIngress":
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

    private void processGRPCIngress(Component component, LinkedHashMap attributeMap) {
        GRPC grpc = new GRPC();
        grpc.setPort((int) ((BInteger) attributeMap.get("gatewayPort")).intValue());
        grpc.setBackendPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        String protoFile = ((BString) attributeMap.get("protoFile")).stringValue();
        if (!protoFile.isEmpty()) {
            copyResourceToTarget(protoFile);
        }
        component.setProtocol(PROTOCOL_GRPC);
        component.setContainerPort(grpc.getBackendPort());
        grpc.setBackendHost(component.getService());
        component.addGRPC(grpc);
    }

    private void processTCPIngress(Component component, LinkedHashMap attributeMap) {
        TCP tcp = new TCP();
        tcp.setPort((int) ((BInteger) attributeMap.get("gatewayPort")).intValue());
        tcp.setBackendPort((int) ((BInteger) attributeMap.get("backendPort")).intValue());
        component.setProtocol(PROTOCOL_TCP);
        component.setContainerPort(tcp.getBackendPort());
        component.addTCP(tcp);
    }

    private void processHttpIngress(Component component, LinkedHashMap attributeMap) {
        API httpAPI = new API();
        int containerPort = (int) ((BInteger) attributeMap.get("port")).intValue();
        // Validate the container port is same for all the ingresses.
        if (component.getContainerPort() > 0 && containerPort != component.getContainerPort()) {
            throw new BallerinaException("Invalid container port" + containerPort + ". Multiple container ports are " +
                    "not supported.");
        }
        component.setContainerPort(containerPort);
        // Process optional attributes
        if (attributeMap.containsKey("context")) {
            httpAPI.setContext(((BString) attributeMap.get("context")).stringValue());
        }

        if (attributeMap.containsKey("expose")) {
            if ("global".equals(((BString) attributeMap.get("expose")).stringValue())) {
                httpAPI.setGlobal(true);
                httpAPI.setBackend(component.getService());
            } else if ("local".equals(((BString) attributeMap.get("expose")).stringValue())) {
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

    private void processWebIngress(Component component, LinkedHashMap attributeMap) {
        Web webIngress = new Web();
        LinkedHashMap gatewayConfig = ((BMap) attributeMap.get("gatewayConfig")).getMap();
        API httpAPI = new API();
        int containerPort = (int) ((BInteger) attributeMap.get("port")).intValue();
        // Validate the container port is same for all the ingresses.
        if (component.getContainerPort() > 0 && containerPort != component.getContainerPort()) {
            throw new BallerinaException("Invalid container port" + containerPort + ". Multiple container ports are " +
                    "not supported.");
        }
        component.setContainerPort(containerPort);
        httpAPI.setGlobal(true);
        httpAPI.setBackend(component.getService());
        httpAPI.setContext(((BString) gatewayConfig.get("context")).stringValue());
        webIngress.setHttpAPI(httpAPI);
        webIngress.setVhost(((BString) gatewayConfig.get("vhost")).stringValue());
        if (gatewayConfig.containsKey("tls")) {
            // TLS enabled
            LinkedHashMap tlsConfig = ((BMap) gatewayConfig.get("tls")).getMap();
            webIngress.setTlsKey(((BString) tlsConfig.get("key")).stringValue());
            webIngress.setTlsCert(((BString) tlsConfig.get("cert")).stringValue());
        }
        if (gatewayConfig.containsKey("oidc")) {
            // OIDC enabled
            webIngress.setOidc(processOidc(((BMap) gatewayConfig.get("oidc")).getMap()));
        }
        component.addWeb(webIngress);
    }


    /**
     * Extract the scale policy.
     *
     * @param scalePolicy Scale policy to be processed
     * @param component   current component
     */
    private void processAutoScalePolicy(LinkedHashMap<?, ?> scalePolicy, Component component) {
        LinkedHashMap bScalePolicy = ((BMap) scalePolicy.get("policy")).getMap();
        boolean bOverridable = ((BBoolean) scalePolicy.get("overridable")).booleanValue();

        List<AutoScalingResourceMetric> autoScalingResourceMetrics = new ArrayList<>();
        LinkedHashMap cpuPercentage = (((BMap) bScalePolicy.get("cpuPercentage")).getMap());
        long percentage = ((BInteger) cpuPercentage.get("percentage")).intValue();
        AutoScalingResourceMetric autoScalingResourceMetric
                = new AutoScalingResourceMetric(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_CPU, (int) percentage);
        autoScalingResourceMetrics.add(autoScalingResourceMetric);

        AutoScalingPolicy autoScalingPolicy = new AutoScalingPolicy();
        autoScalingPolicy.setMinReplicas(((BInteger) bScalePolicy.get("minReplicas")).intValue());
        autoScalingPolicy.setMaxReplicas(((BInteger) bScalePolicy.get("maxReplicas")).intValue());
        autoScalingPolicy.setMetrics(autoScalingResourceMetrics);
        component.setAutoScaling(new AutoScaling(autoScalingPolicy, bOverridable));
    }


    private void generateDependenciesFile(LinkedHashMap<?, ?> dependencies) {
        StringBuffer buffer = new StringBuffer();
        dependencies.forEach((key, value) -> {
            LinkedHashMap attributeMap = ((BMap) value).getMap();
            final String depText = ((BString) attributeMap.get("org")).stringValue() + "/"
                    + ((BString) attributeMap.get("name")).stringValue() + ":"
                    + ((BString) attributeMap.get("ver")).stringValue();
            buffer.append(key.toString()).append("=").append(depText).append("\n");
        });
        String targetFileNameWithPath =
                OUTPUT_DIRECTORY + File.separator + "tmp" + File.separator + "dependencies.properties";
        try {
            writeToFile(buffer.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            throw new BallerinaException("Error occurred while generating reference file " + targetFileNameWithPath);
        }
    }

    private void generateCell() {
        List<Component> components =
                new ArrayList<>(cellImage.getComponentNameToComponentMap().values());
        GatewaySpec gatewaySpec = new GatewaySpec();
        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
        for (Component component : components) {
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
                gatewaySpec.addTCP(component.getTcpList());
                templateSpec.setServicePort(component.getTcpList().get(0).getPort());
            } else if (component.getGrpcList().size() > 0) {
                gatewaySpec.setType(ENVOY_GATEWAY);
                templateSpec.setServicePort(component.getGrpcList().get(0).getPort());
                gatewaySpec.addGRPC(component.getGrpcList());
            }
            templateSpec.setReplicas(component.getReplicas());
            templateSpec.setProtocol(component.getProtocol());
            List<EnvVar> envVarList = new ArrayList<>();
            component.getEnvVars().forEach((key, value) -> {
                if (StringUtils.isEmpty(value)) {
                    printWarning("Value is empty for environment variable \"" + key + "\"");
                }
                envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
            });
            templateSpec.setContainer(new ContainerBuilder()
                    .withImage(component.getSource())
                    .withPorts(new ContainerPortBuilder()
                            .withContainerPort(component.getContainerPort())
                            .build())
                    .withEnv(envVarList)
                    .build());

            AutoScaling autoScaling = component.getAutoScaling();
            if (autoScaling != null) {
                templateSpec.setAutoscaling(generateAutoScaling(autoScaling));
            }
            ServiceTemplate serviceTemplate = new ServiceTemplate();
            serviceTemplate.setMetadata(new ObjectMetaBuilder()
                    .withName(component.getService())
                    .withLabels(component.getLabels())
                    .build());
            serviceTemplate.setSpec(templateSpec);
            serviceTemplateList.add(serviceTemplate);
        }

        GatewayTemplate gatewayTemplate = new GatewayTemplate();
        gatewayTemplate.setSpec(gatewaySpec);

        CellSpec cellSpec = new CellSpec();
        cellSpec.setGatewayTemplate(gatewayTemplate);
        cellSpec.setServicesTemplates(serviceTemplateList);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(cellImage.getCellName()))
                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, cellImage.getOrgName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, cellImage.getCellName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, cellImage.getCellVersion())
                .build();
        Cell cell = new Cell(objectMeta, cellSpec);
        String targetPath =
                OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + cellImage.getCellName() + YAML;
        try {
            writeToFile(toYaml(cell), targetPath);
        } catch (IOException e) {
            throw new BallerinaException(e.getMessage() + " " + targetPath);
        }
    }

    private AutoScalingSpec generateAutoScaling(AutoScaling autoScaling) {
        HorizontalPodAutoscalerSpecBuilder autoScaleSpecBuilder = new HorizontalPodAutoscalerSpecBuilder()
                .withMaxReplicas((int) autoScaling.getPolicy().getMaxReplicas())
                .withMinReplicas((int) autoScaling.getPolicy().getMinReplicas());

        // Generating scale policy metrics config
        for (AutoScalingResourceMetric metric : autoScaling.getPolicy().getMetrics()) {
            autoScaleSpecBuilder.addToMetrics(new MetricSpecBuilder()
                    .withType(CelleryConstants.AUTO_SCALING_METRIC_RESOURCE)
                    .withNewResource()
                    .withName(metric.getName())
                    .withTargetAverageUtilization(metric.getValue())
                    .endResource()
                    .build());
        }
        AutoScalingSpec autoScalingSpec = new AutoScalingSpec();
        autoScalingSpec.setOverridable(autoScaling.isOverridable());
        autoScalingSpec.setPolicy(autoScaleSpecBuilder.build());
        return autoScalingSpec;

    }

    /**
     * Generate a Cell Reference that can be used by other cells.
     */
    private void generateCellReference() {
        JSONObject json = new JSONObject();
        cellImage.getComponentNameToComponentMap().forEach((componentName, component) -> {
            component.getApis().forEach(api -> {
                String url = DEFAULT_GATEWAY_PROTOCOL + "://" + INSTANCE_NAME_PLACEHOLDER + "--gateway" +
                        "-service:" + DEFAULT_GATEWAY_PORT + "/" + api.getContext();
                json.put(api.getContext() + "_api_url", url.replaceAll("(?<!http:)//", "/"));
            });
            component.getTcpList().forEach(tcp -> json.put(componentName + "_tcp_port", tcp.getPort()));
            component.getGrpcList().forEach(grpc -> json.put(componentName + "_grpc_port", grpc.getPort()));
        });
        String targetFileNameWithPath =
                OUTPUT_DIRECTORY + File.separator + "ref" + File.separator + REFERENCE_FILE_NAME;
        try {
            writeToFile(json.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            throw new BallerinaException("Error occurred while generating reference file " + targetFileNameWithPath);
        }
    }
}
