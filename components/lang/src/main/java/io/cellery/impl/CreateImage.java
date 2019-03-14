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

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import io.cellery.CelleryConstants;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.AutoScaling;
import io.cellery.models.AutoScalingPolicy;
import io.cellery.models.AutoScalingResourceMetric;
import io.cellery.models.AutoScalingSpec;
import io.cellery.models.Cell;
import io.cellery.models.CellCache;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.ComponentHolder;
import io.cellery.models.GRPC;
import io.cellery.models.GatewaySpec;
import io.cellery.models.GatewayTemplate;
import io.cellery.models.ServiceTemplate;
import io.cellery.models.ServiceTemplateSpec;
import io.cellery.models.TCP;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.ContainerPort;
import io.fabric8.kubernetes.api.model.ContainerPortBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.HorizontalPodAutoscalerSpecBuilder;
import io.fabric8.kubernetes.api.model.MetricSpecBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.Context;
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

import java.io.File;
import java.io.IOException;
import java.io.PrintStream;
import java.io.StringWriter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Function;

import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_NAME;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_ORG;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PORT;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PROTOCOL;
import static io.cellery.CelleryConstants.PROTOCOL_GRPC;
import static io.cellery.CelleryConstants.PROTOCOL_HTTP;
import static io.cellery.CelleryConstants.PROTOCOL_TCP;
import static io.cellery.CelleryConstants.TARGET;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.copyResourceToTarget;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.processParameters;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * Native function cellery:createImage.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createImage",
        args = {@Argument(name = "cellImage", type = TypeKind.OBJECT),
                @Argument(name = "orgName", type = TypeKind.STRING),
                @Argument(name = "imageName", type = TypeKind.STRING),
                @Argument(name = "imageVersion", type = TypeKind.STRING)},
        returnType = {@ReturnType(type = TypeKind.STRING)},
        isPublic = true
)
public class CreateImage extends BlockingNativeCallableUnit {

    private static final MustacheFactory mustacheFactory = new DefaultMustacheFactory();
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir") + File.separator + TARGET;

    private ComponentHolder componentHolder;
    private PrintStream out = System.out;

    public void execute(Context ctx) {
        componentHolder = new ComponentHolder();
        String cellVersion = ctx.getStringArgument(2);
        String orgName = ctx.getStringArgument(0);
        String cellName = getValidName(ctx.getStringArgument(1));
        CellCache cellCache = CellCache.getInstance();
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        processComponents((BMap) refArgument.getMap().get("components"));
        processAPIs((BMap) refArgument.getMap().get("apis"));
        processTCP((BMap) refArgument.getMap().get("tcp"));
        processGRPC((BMap) refArgument.getMap().get("grpc"));
        Set<Component> components = new HashSet<>(componentHolder.getComponentNameToComponentMap().values());
        cellCache.setCellNameToComponentMap(cellName, components);
        generateCell(orgName, cellName, cellVersion);
        generateCellReference(cellName, cellVersion);
        ctx.setReturnValues(new BBoolean(true));
    }

    private void processComponents(BMap<?, ?> components) {
        components.getMap().forEach((key, value) -> {
            LinkedHashMap componentValues = ((BMap) value).getMap();
            Component component = new Component();
            // Mandatory fields
            component.setName(((BString) componentValues.get("name")).stringValue());
            component.setService(getValidName(component.getName()));
            component.setReplicas((int) ((BInteger) componentValues.get("replicas")).intValue());
            component.setSource(((BMap<?, ?>) componentValues.get("source")).getMap().get("image").toString());

            // Optional fields
            if (componentValues.containsKey("ingresses")) {
                processIngressPort(((BMap<?, ?>) componentValues.get("ingresses")).getMap(), component);
            }
            if (componentValues.containsKey("parameters")) {
                processParameters(component, ((BMap<?, ?>) componentValues.get("parameters")).getMap());
            }
            if (componentValues.containsKey("labels")) {
                ((BMap<?, ?>) componentValues.get("labels")).getMap().forEach((labelKey, labelValue) ->
                        component.addLabel(labelKey.toString(), labelValue.toString()));
            }
            if (componentValues.containsKey("autoscaling")) {
                processAutoScalePolicy(((BMap<?, ?>) componentValues.get("autoscaling")).getMap(), component);
            }
            componentHolder.addComponent(component);
        });
    }


    /**
     * Extract the ingress port.
     *
     * @param ingressMap list of ingresses defined
     * @param component  current component
     */
    private void processIngressPort(LinkedHashMap<?, ?> ingressMap, Component component) {
        AtomicInteger containerPort = new AtomicInteger(0);
        AtomicInteger servicePort = new AtomicInteger(0);
        ingressMap.forEach((name, entry) -> ((BMap<?, ?>) entry).getMap().forEach((key, value) -> {
            if ("port".equals(key.toString())) {
                containerPort.set((int) ((BInteger) value).intValue());
            } else if ("targetPort".equals(key.toString())) {
                servicePort.set((int) ((BInteger) value).intValue());
            }
        }));
        if (servicePort.get() == 0) {
            //HTTP ingress
            component.addPorts(containerPort.get(), DEFAULT_GATEWAY_PORT);
        } else {
            //TCP ingress
            component.addPorts(containerPort.get(), containerPort.get());
        }
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

    private void processAPIs(BMap<?, ?> apiMap) {
        apiMap.getMap().forEach((key, value) -> {
            LinkedHashMap apiValues = ((BMap) value).getMap();
            API api = new API();
            api.setGlobal(((BBoolean) apiValues.get("global")).booleanValue());
            String componentName = ((BString) apiValues.get("targetComponent")).stringValue();
            LinkedHashMap<?, ?> ingress = ((BMap<?, ?>) apiValues.get("ingress")).getMap();
            api.setContext(((BString) ingress.get("context")).stringValue());
            List<APIDefinition> apiDefinitions = new ArrayList<>();
            BValueArray ingressDefs = ((BValueArray) ingress.get("definitions"));
            for (int i = 0; i < ingressDefs.size(); i++) {
                APIDefinition apiDefinition = new APIDefinition();
                LinkedHashMap definitions = ((BMap) ingressDefs.getBValue(i)).getMap();
                apiDefinition.setPath(((BString) definitions.get("path")).stringValue());
                apiDefinition.setMethod(((BString) definitions.get("method")).stringValue());
                apiDefinitions.add(apiDefinition);
            }
            api.setDefinitions(apiDefinitions);
            Component component = componentHolder.getComponent(componentName);
            component.setProtocol(PROTOCOL_HTTP);
            api.setBackend(component.getService());
            component.addApi(api);
        });
    }

    private void processTCP(BMap<?, ?> tcpMap) {
        tcpMap.getMap().forEach((key, value) -> {
            LinkedHashMap tcpValues = ((BMap) value).getMap();
            TCP tcp = new TCP();
            String componentName = ((BString) tcpValues.get("targetComponent")).stringValue();
            LinkedHashMap<?, ?> ingress = ((BMap<?, ?>) tcpValues.get("ingress")).getMap();
            tcp.setPort((int) ((BInteger) ingress.get("port")).intValue());
            tcp.setBackendPort((int) ((BInteger) ingress.get("targetPort")).intValue());
            Component component = componentHolder.getComponent(componentName);
            component.setProtocol(PROTOCOL_TCP);
            tcp.setBackendHost(component.getService());
            component.addTCP(tcp);
        });
    }

    private void processGRPC(BMap<?, ?> grpcMap) {
        grpcMap.getMap().forEach((key, value) -> {
            LinkedHashMap grpcValues = ((BMap) value).getMap();
            GRPC grpc = new GRPC();
            String componentName = ((BString) grpcValues.get("targetComponent")).stringValue();
            LinkedHashMap<?, ?> ingress = ((BMap<?, ?>) grpcValues.get("ingress")).getMap();
            grpc.setPort((int) ((BInteger) ingress.get("port")).intValue());
            grpc.setBackendPort((int) ((BInteger) ingress.get("targetPort")).intValue());
            String protoFile = ((BString) ingress.get("protoFile")).stringValue();
            if (!protoFile.isEmpty()) {
                copyResourceToTarget(protoFile);
            }
            Component component = componentHolder.getComponent(componentName);
            component.setProtocol(PROTOCOL_GRPC);
            grpc.setBackendHost(component.getService());
            component.addGRPC(grpc);
        });
    }


    private void generateCell(String orgName, String name, String version) {
        List<Component> components =
                new ArrayList<>(componentHolder.getComponentNameToComponentMap().values());
        GatewaySpec spec = new GatewaySpec();
        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();

        for (Component component : components) {
            spec.addHttpAPI(component.getApis());
            spec.addTCP(component.getTcpList());
            spec.addGRPC(component.getGrpcList());
            ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();
            templateSpec.setReplicas(component.getReplicas());
            templateSpec.setProtocol(component.getProtocol());
            List<EnvVar> envVarList = new ArrayList<>();
            component.getEnvVars().forEach((key, value) -> {
                if (StringUtils.isEmpty(value)) {
                    out.println("Warning: Value is empty for environment variable \"" + key + "\"");
                }
                envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
            });

            //TODO:Fix service port
            List<ContainerPort> ports = new ArrayList<>();
            component.getContainerPortToServicePortMap().forEach((containerPort, servicePort) -> {
                ports.add(new ContainerPortBuilder().withContainerPort(containerPort).build());
                templateSpec.setServicePort(servicePort);
            });
            templateSpec.setContainer(new ContainerBuilder()
                    .withImage(component.getSource())
                    .withPorts(ports)
                    .withEnv(envVarList)
                    .build());

            AutoScaling autoScaling = component.getAutoScaling();
            if (autoScaling != null) {
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
                templateSpec.setAutoscaling(autoScalingSpec);
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
        gatewayTemplate.setSpec(spec);

        CellSpec cellSpec = new CellSpec();
        cellSpec.setGatewayTemplate(gatewayTemplate);
        cellSpec.setServicesTemplates(serviceTemplateList);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(name))
                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, orgName)
                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, name)
                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, version)
                .build();
        Cell cell = new Cell(objectMeta, cellSpec);

        String targetPath = OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + name + YAML;
        try {
            writeToFile(toYaml(cell), targetPath);
        } catch (IOException e) {
            throw new BallerinaException(e.getMessage() + " " + targetPath);
        }
    }

    /**
     * Generate a Cell Reference that can be used by other cells.
     *
     * @param name The name of the Cell
     */
    private void generateCellReference(String name, String cellVersion) {
        String ballerinaPathName = name.replace("-", "_").toLowerCase(Locale.ENGLISH);

        // Generating the context
        Map<String, Object> context = new HashMap<>();
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_NAME, name);
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_VERSION, cellVersion);
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_GATEWAY_PORT, DEFAULT_GATEWAY_PORT);
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_GATEWAY_PROTOCOL, DEFAULT_GATEWAY_PROTOCOL);
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_COMPONENTS,
                componentHolder.getComponentNameToComponentMap().values());
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_HANDLE_API_NAME,
                (Function<Object, Object>) string -> convertToTitleCase((String) string, "/|-"));
        context.put(CelleryConstants.CELL_REFERENCE_TEMPLATE_CONTEXT_HANDLE_TYPE_NAME,
                (Function<Object, Object>) string -> convertToTitleCase((String) string, "-"));

        // Writing the template to file
        String targetFileNameWithPath = OUTPUT_DIRECTORY + File.separator + "bal" + File.separator + ballerinaPathName
                + File.separator + ballerinaPathName + "_reference.bal";
        writeMustacheTemplateToFile(context, targetFileNameWithPath);
    }

    /**
     * Write a mustache template to a file.
     *
     * @param mustacheContext        The mustache context to be passed to the template
     * @param targetFileNameWithPath The name of the target file name with path
     */
    private void writeMustacheTemplateToFile(Object mustacheContext, String targetFileNameWithPath) {
        Mustache mustache = mustacheFactory.compile(CelleryConstants.CELL_REFERENCE_TEMPLATE_FILE);

        // Writing the template to the file
        try {
            StringWriter writer = new StringWriter();
            mustache.execute(writer, mustacheContext).flush();
            writeToFile(writer.toString(), targetFileNameWithPath);
        } catch (IOException e) {
            throw new BallerinaException(e.getMessage() + " " + targetFileNameWithPath);
        }
    }


    /**
     * Convert a string to title case.
     *
     * @param text      The text to be converted to title case
     * @param separator The word separator
     * @return The title case text
     */
    private String convertToTitleCase(String text, String separator) {
        String[] wordsArray = text.split(separator);
        StringBuilder stringBuilder = new StringBuilder();
        for (String word : wordsArray) {
            if (word.length() >= 1) {
                stringBuilder.append(word.substring(0, 1).toUpperCase(Locale.ENGLISH));
            }
            if (word.length() >= 2) {
                stringBuilder.append(word.substring(1).toLowerCase(Locale.ENGLISH));
            }
        }
        return stringBuilder.toString();
    }

}


