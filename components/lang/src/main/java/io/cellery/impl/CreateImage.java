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

import com.esotericsoftware.yamlbeans.YamlWriter;
import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import io.cellery.CelleryConstants;
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.Cell;
import io.cellery.models.CellCache;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.ComponentHolder;
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
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BBoolean;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BRefType;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.io.PrintStream;
import java.io.StringWriter;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Function;

import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PORT;
import static io.cellery.CelleryConstants.DEFAULT_GATEWAY_PROTOCOL;
import static io.cellery.CelleryConstants.TARGET;
import static org.apache.commons.lang3.StringUtils.removePattern;

/**
 * Native function cellery:createImage.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createImage",
        args = {@Argument(name = "cellImage", type = TypeKind.OBJECT),
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
        String cellVersion = ctx.getNullableStringArgument(1);
        String cellName = getValidName(ctx.getStringArgument(0));
        CellCache cellCache = CellCache.getInstance();
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        processComponents(((BValueArray) refArgument.getMap().get("components")).getValues());
        processAPIs(((BValueArray) refArgument.getMap().get("apis")).getValues());
        processTCP(((BValueArray) refArgument.getMap().get("tcp")).getValues());
        Set<Component> components = new HashSet<>(componentHolder.getComponentNameToComponentMap().values());
        cellCache.setCellNameToComponentMap(cellName, components);
        generateCell(cellName);
        generateCellReference(cellName, cellVersion);
        ctx.setReturnValues(new BBoolean(true));
    }

    private void processComponents(BRefType<?>[] components) {
        for (BRefType componentDefinition : components) {
            if (componentDefinition == null) {
                continue;
            }
            Component component = new Component();
            ((BMap<?, ?>) componentDefinition).getMap().forEach((key, value) -> {
                switch (key.toString()) {
                    case "name":
                        component.setName(value.toString());
                        component.setService(getValidName(value.toString()));
                        break;
                    case "replicas":
                        component.setReplicas(Integer.parseInt(value.toString()));
                        break;
                    case "isStub":
                        component.setIsStub(Boolean.parseBoolean(value.toString()));
                        break;
                    case "source":
                        component.setSource(Collections.
                                singletonList(((BMap) value).getMap()).get(0).values().toArray()[0].toString());
                        break;
                    case "ingresses":
                        processIngressPort(((BMap<?, ?>) value).getMap(), component);
                        break;
                    case "env":
                        ((BMap<?, ?>) value).getMap().forEach((envKey, envValue) ->
                                component.addEnv(envKey.toString(), envValue.toString()));
                        break;
                    case "parameters":
                        ((BMap<?, ?>) value).getMap().forEach((k, v) -> {
                            if (!((BMap) v).getMap().get("value").toString().isEmpty()) {
                                component.addEnv(k.toString(), ((BMap) v).getMap().get("value").toString());
                            }
                            //TODO:Handle secrets
                        });
                        break;
                    case "labels":
                        ((BMap<?, ?>) value).getMap().forEach((labelKey, labelValue) ->
                                component.addLabel(labelKey.toString(), labelValue.toString()));
                        break;

                    default:
                        break;
                }
            });
            componentHolder.addComponent(component);
        }
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
                containerPort.set(Integer.parseInt(value.toString()));
            } else if ("targetPort".equals(key.toString())) {
                servicePort.set(Integer.parseInt(value.toString()));
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

    private void processAPIs(BRefType<?>[] apiMap) {
        for (BRefType apiDefinition : apiMap) {
            if (apiDefinition == null) {
                continue;
            }
            API api = new API();
            String componentName = null;
            for (Map.Entry<?, ?> entry : ((BMap<?, ?>) apiDefinition).getMap().entrySet()) {
                Object key = entry.getKey();
                BValue value = (BValue) entry.getValue();
                switch (key.toString()) {
                    case "global":
                        api.setGlobal(Boolean.parseBoolean(value.toString()));
                        break;
                    case "targetComponent":
                        componentName = value.toString();
                        break;
                    case "ingress":
                        ((BMap<?, ?>) value).getMap().forEach((contextKey, contextValue) -> {
                            switch (contextKey.toString()) {
                                case "basePath":
                                    api.setContext(contextValue.toString());
                                    break;
                                case "definitions":
                                    api.setDefinitions(processDefinitions(((BValueArray) contextValue).getValues()));
                                    break;
                                default:
                                    break;
                            }
                        });
                        break;
                    default:
                        break;

                }
            }
            if (componentName != null) {
                componentHolder.addAPI(componentName, api);
            } else {
                throw new BallerinaException("Undefined target component.");
            }
        }
    }

    private void processTCP(BRefType<?>[] tcpMap) {
        for (BRefType tcpDefinition : tcpMap) {
            if (tcpDefinition == null) {
                continue;
            }
            TCP tcp = new TCP();
            String componentName = null;
            for (Map.Entry<?, ?> entry : ((BMap<?, ?>) tcpDefinition).getMap().entrySet()) {
                Object key = entry.getKey();
                BValue value = (BValue) entry.getValue();
                switch (key.toString()) {
                    case "targetComponent":
                        componentName = value.toString();
                        break;
                    case "ingress":
                        ((BMap<?, ?>) value).getMap().forEach((contextKey, contextValue) -> {
                            switch (contextKey.toString()) {
                                case "port":
                                    tcp.setPort(Integer.parseInt(contextValue.toString()));
                                    break;
                                case "targetPort":
                                    tcp.setBackendPort(Integer.parseInt(contextValue.toString()));
                                    break;
                                default:
                                    break;
                            }
                        });
                        break;
                    default:
                        break;

                }
            }
            if (componentName != null) {
                componentHolder.addTCP(componentName, tcp);
            } else {
                throw new BallerinaException("Undefined target component.");
            }
        }
    }

    private List<APIDefinition> processDefinitions(BRefType<?>[] definitions) {
        List<APIDefinition> apiDefinitions = new ArrayList<>();
        for (BRefType definition : definitions) {
            if (definition == null) {
                continue;
            }
            APIDefinition apiDefinition = new APIDefinition();
            ((BMap<?, ?>) definition).getMap().forEach((key, value) -> {
                switch (key.toString()) {
                    case "path":
                        apiDefinition.setPath(value.toString());
                        break;
                    case "method":
                        apiDefinition.setMethod(value.toString());
                        break;
                    default:
                        break;

                }
            });
            apiDefinitions.add(apiDefinition);
        }
        return apiDefinitions;
    }

    /**
     * Returns valid kubernetes name.
     *
     * @param name actual value
     * @return valid name
     */
    private String getValidName(String name) {
        return name.toLowerCase(Locale.getDefault()).replace("_", "-").replace(".", "-");
    }

    private void generateCell(String name) {
        List<Component> components =
                new ArrayList<>(componentHolder.getComponentNameToComponentMap().values());
        GatewaySpec spec = new GatewaySpec();
        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
        for (Component component : components) {
            if (component.getIsStub()) {
                continue;
            }
            spec.setHttp(component.getApis());
            spec.setTcp(component.getTcpList());
            ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();
            templateSpec.setReplicas(component.getReplicas());
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
        Cell cell = new Cell(new ObjectMetaBuilder().withName(getValidName(name)).build(), cellSpec);
        String targetPath = OUTPUT_DIRECTORY + File.separator + "cellery" + File.separator + name + ".yaml";
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
     * Write content to a File. Create the required directories if they don't not exists.
     *
     * @param context    context of the file
     * @param targetPath target file path
     * @throws IOException If an error occurs when writing to a file
     */
    private void writeToFile(String context, String targetPath) throws IOException {
        File newFile = new File(targetPath);
        // delete if file exists
        if (newFile.exists()) {
            Files.delete(Paths.get(newFile.getPath()));
        }
        //create required directories
        if (newFile.getParentFile().mkdirs()) {
            Files.write(Paths.get(targetPath), context.getBytes(StandardCharsets.UTF_8));
            return;
        }
        Files.write(Paths.get(targetPath), context.getBytes(StandardCharsets.UTF_8));
    }

    /**
     * Generates Yaml from a object.
     *
     * @param object Object
     * @param <T>    Any Object type
     * @return Yaml as a string.
     */
    private <T> String toYaml(T object) {
        try (StringWriter stringWriter = new StringWriter()) {
            YamlWriter writer = new YamlWriter(stringWriter);
            writer.write(object);
            writer.getConfig().writeConfig.setWriteRootTags(false); //replaces only root tag
            writer.close(); //don't add this to finally, because the text will not be flushed
            return removeTags(stringWriter.toString());
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private String removeTags(String string) {
        //a tag is a sequence of characters starting with ! and ending with whitespace
        return removePattern(string, " ![^\\s]*");
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


