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
import io.cellery.models.API;
import io.cellery.models.APIDefinition;
import io.cellery.models.Cell;
import io.cellery.models.CellCache;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.ComponentHolder;
import io.cellery.models.Egress;
import io.cellery.models.GatewaySpec;
import io.cellery.models.GatewayTemplate;
import io.cellery.models.ServiceTemplate;
import io.cellery.models.ServiceTemplateSpec;
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
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;

import static org.apache.commons.lang3.StringUtils.isEmpty;
import static org.apache.commons.lang3.StringUtils.removePattern;

/**
 * Native function cellery/build.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createImage",
        args = {@Argument(name = "cellImage", type = TypeKind.OBJECT)},
        returnType = {@ReturnType(type = TypeKind.STRING)},
        isPublic = true
)
public class CreateImage extends BlockingNativeCallableUnit {

    private int gatewayPort = 80;
    private ComponentHolder componentHolder;
    private PrintStream out = System.out;

    public void execute(Context ctx) {
        componentHolder = new ComponentHolder();
        String cellName = (((BMap) ctx.getNullableRefArgument(0)).getMap()).get("name").toString();
        CellCache cellCache = CellCache.getInstance();
        processComponents(
                ((BValueArray) ((BMap) ctx.getNullableRefArgument(0)).getMap().get("components")).getValues());
        processAPIs(((BValueArray) ((BMap) ctx.getNullableRefArgument(0)).getMap().get("apis")).getValues());
        Set<Component> components = new HashSet<>(componentHolder.getComponentNameToComponentMap().values());
        cellCache.setCellNameToComponentMap(cellName, components);
        processCellEgress(((BValueArray) ((BMap) ctx.getNullableRefArgument(0)).getMap().get("egresses")).getValues());
        generateCell(cellName);
        ctx.setReturnValues(new BBoolean(true));
    }

    private void processCellEgress(BRefType<?>[] egresses) {
        for (BRefType egressDefinition : egresses) {
            if (egressDefinition == null) {
                continue;
            }
            Egress egress = new Egress();
            for (Map.Entry<?, ?> entry : ((BMap<?, ?>) egressDefinition).getMap().entrySet()) {
                Object key = entry.getKey();
                BValue value = (BValue) entry.getValue();
                switch (key.toString()) {
                    case "targetCell":
                        egress.setCellName(value.toString());
                        break;
                    case "envVar":
                        egress.setEnvVar(value.toString());
                        break;
                    default:
                        break;
                }
            }
            componentHolder.getComponentNameToComponentMap().values().forEach(component -> component.addEgress(egress));
        }
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
                    case "egresses":
                        processEgress(((BValueArray) value).getValues(), component);
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

    private void processEgress(BRefType<?>[] egresses, Component component) {
        for (BRefType egressDefinition : egresses) {
            if (egressDefinition == null) {
                continue;
            }
            Egress egress = new Egress();
            for (Map.Entry<?, ?> entry : ((BMap<?, ?>) egressDefinition).getMap().entrySet()) {
                Object key = entry.getKey();
                BValue value = (BValue) entry.getValue();
                switch (key.toString()) {
                    case "targetComponent":
                        egress.setTargetComponent(value.toString());
                        break;
                    case "envVar":
                        egress.setEnvVar(value.toString());
                        break;
                    default:
                        break;
                }
            }
            component.addEgress(egress);
        }
    }

    /**
     * Extract the ingress port.
     *
     * @param ingressMap list of ingresses defined
     * @param component  current component
     */
    private void processIngressPort(LinkedHashMap<?, ?> ingressMap, Component component) {
        ingressMap.forEach((name, entry) -> ((BMap<?, ?>) entry).getMap().forEach((key, value) -> {
            switch (key.toString()) {
                case "port":
                    component.addPorts(Integer.parseInt(value.toString()), gatewayPort);
                    break;
                default:
                    break;
            }
        }));
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
            spec.setApis(component.getApis());
            ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();
            templateSpec.setReplicas(component.getReplicas());
            component.getEgresses().forEach((egress) -> {
                String serviceName;
                if (isEmpty(egress.getCellName())) {
                    //Inter cell mapping
                    Component targetComponent =
                            componentHolder.getComponentNameToComponentMap().get(egress.getTargetComponent());
                    if (targetComponent == null) {
                        throw new BallerinaException("Invalid component egress. " +
                                "Components " + component.getName() + " and " + egress.getTargetComponent() + " are " +
                                "not defined in the same cell.");
                    }
                    // Generate service name => cellName--serviceName-service
                    serviceName = getValidName(name) + "--" +
                            targetComponent.getService()
                            + "-service";
                } else {
                    serviceName = getValidName(egress.getCellName()) + "--gateway-service";
                }
                component.addEnv(egress.getEnvVar(), serviceName);

            });
            List<EnvVar> envVarList = new ArrayList<>();
            component.getEnvVars().forEach((key, value) -> {
                if (StringUtils.isEmpty(value)) {
                    out.println("Warning: Value is empty for environment variable \"" + key + "\"");
                }
                envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
            });
            //TODO:Fix service port
            templateSpec.setServicePort(gatewayPort);
            List<ContainerPort> ports = new ArrayList<>();
            component.getContainerPortToServicePortMap().forEach((containerPort, servicePort) ->
                    ports.add(new ContainerPortBuilder().withContainerPort(containerPort).build()));
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
        String targetPath = System.getProperty("user.dir") + File.separator + "target" + File.separator
                + "cellery" + File.separator + name + ".yaml";
        try {
            writeToFile(toYaml(cell), targetPath);
        } catch (IOException e) {
            throw new BallerinaException(e.getMessage() + " " + targetPath);
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


}


