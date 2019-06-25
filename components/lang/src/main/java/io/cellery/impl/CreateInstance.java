/*
 *   Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

import com.esotericsoftware.yamlbeans.YamlReader;
import com.google.gson.Gson;
import io.cellery.models.Cell;
import io.cellery.models.CellImage;
import io.cellery.models.Component;
import io.cellery.models.Dependency;
import io.cellery.models.GatewaySpec;
import io.cellery.models.OIDC;
import io.cellery.models.ServiceTemplate;
import io.fabric8.kubernetes.api.model.Probe;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import io.fabric8.kubernetes.client.internal.SerializationUtils;
import org.apache.commons.codec.binary.Base64;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;

import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES;
import static io.cellery.CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR;
import static io.cellery.CelleryConstants.ENV_VARS;
import static io.cellery.CelleryConstants.INGRESSES;
import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.INSTANCE_NAME_PLACEHOLDER;
import static io.cellery.CelleryConstants.PROBES;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.processEnvVars;
import static io.cellery.CelleryUtils.processProbes;
import static io.cellery.CelleryUtils.processWebIngress;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;
import static org.apache.commons.lang3.StringUtils.removePattern;

/**
 * Native function cellery:createInstance.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createInstance",
        args = {@Argument(name = "cellImage", type = TypeKind.RECORD),
                @Argument(name = "iName", type = TypeKind.RECORD),
                @Argument(name = "instances", type = TypeKind.MAP)
        },
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class CreateInstance extends BlockingNativeCallableUnit {
    private static final Logger log = LoggerFactory.getLogger(CreateInstance.class);
    private CellImage cellImage = new CellImage();

    public void execute(Context ctx) {
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        String cellName = ((BString) nameStruct.get("name")).stringValue();
        String instanceName = ((BString) nameStruct.get(INSTANCE_NAME)).stringValue();
        String destinationPath = System.getenv(CELLERY_IMAGE_DIR_ENV_VAR) + File.separator +
                "artifacts" + File.separator + "cellery";
        Map dependencyInfo = ((BMap) ctx.getNullableRefArgument(2)).getMap();
        String cellYAMLPath = destinationPath + File.separator + cellName + YAML;
        Cell cell = getInstance(cellYAMLPath);
        updateDependencyAnnotations(cell, dependencyInfo);
        try {
            processComponents((BMap) refArgument.getMap().get("components"));
            cell.getSpec().getServicesTemplates().forEach(serviceTemplate -> {
                String componentName = serviceTemplate.getMetadata().getName();
                Component updatedComponent = cellImage.getComponentNameToComponentMap().get(componentName);
                //Replace env values defined in the YAML.
                updateEnvVar(instanceName, serviceTemplate, updatedComponent, dependencyInfo);
                // Update Gateway Config
                updateGatewayConfig(instanceName, destinationPath, cell, updatedComponent);
                // Update liveness and readiness probe
                updateProbes(serviceTemplate, updatedComponent);
            });
            writeToFile(removeTags(toYaml(cell)), cellYAMLPath);
        } catch (IOException | BallerinaException e) {
            String error = "Unable to persist updated cell yaml " + destinationPath;
            log.error(error, e);
            ctx.setReturnValues(BLangVMErrors.createError(ctx, error + ". " + e.getMessage()));
        }
    }

    /**
     * Update Gateway Config.
     *
     * @param instanceName     instance Name
     * @param destinationPath  destination file path
     * @param cell             Cell object
     * @param updatedComponent Updated component object
     */
    private void updateGatewayConfig(String instanceName, String destinationPath, Cell cell,
                                     Component updatedComponent) {
        GatewaySpec gatewaySpec = cell.getSpec().getGatewayTemplate().getSpec();
        updatedComponent.getWebList().forEach(web -> {
            // Create TLS secret yaml and set the name
            if (StringUtils.isNoneEmpty(web.getTlsKey())) {
                Map<String, String> tlsMap = new HashMap<>();
                tlsMap.put("tls.key",
                        Base64.encodeBase64String(web.getTlsKey().getBytes(StandardCharsets.UTF_8)));
                tlsMap.put("tls.crt",
                        Base64.encodeBase64String(web.getTlsCert().getBytes(StandardCharsets.UTF_8)));
                String tlsSecretName = instanceName + "--tls-secret";
                createSecret(tlsSecretName, tlsMap, destinationPath + File.separator + tlsSecretName + ".yaml");
                gatewaySpec.setTlsSecret(tlsSecretName);
            }
            // Set OIDC values
            if (web.getOidc() != null) {
                OIDC oidc = web.getOidc();
                if (StringUtils.isBlank(oidc.getClientSecret()) && StringUtils.isBlank(oidc.getDcrPassword())) {
                    printWarning("OIDC client secret and DCR password are empty.");
                }
                gatewaySpec.setOidc(oidc);
            }
            gatewaySpec.setHost(web.getVhost());
            gatewaySpec.addHttpAPI(Collections.singletonList(web.getHttpAPI()));
        });
    }

    /**
     * Update Probe configurations.
     *
     * @param serviceTemplate  service Template object
     * @param updatedComponent updated component to process env var
     */
    private void updateProbes(ServiceTemplate serviceTemplate, Component updatedComponent) {
        Probe livenessProbe = updatedComponent.getLivenessProbe();
        // Override values with updated values
        if (livenessProbe != null) {
            Probe probe = serviceTemplate.getSpec().getContainer().getLivenessProbe();
            probe.setInitialDelaySeconds(livenessProbe.getInitialDelaySeconds());
            probe.setFailureThreshold(livenessProbe.getFailureThreshold());
            probe.setPeriodSeconds(livenessProbe.getPeriodSeconds());
            probe.setSuccessThreshold(livenessProbe.getSuccessThreshold());
            probe.setTimeoutSeconds(livenessProbe.getTimeoutSeconds());
            serviceTemplate.getSpec().getContainer().setLivenessProbe(probe);
        }

        Probe readinessProbe = updatedComponent.getReadinessProbe();
        // Override values with updated values
        if (readinessProbe != null) {
            Probe probe = serviceTemplate.getSpec().getContainer().getReadinessProbe();
            probe.setInitialDelaySeconds(readinessProbe.getInitialDelaySeconds());
            probe.setFailureThreshold(readinessProbe.getFailureThreshold());
            probe.setPeriodSeconds(readinessProbe.getPeriodSeconds());
            probe.setSuccessThreshold(readinessProbe.getSuccessThreshold());
            probe.setTimeoutSeconds(readinessProbe.getTimeoutSeconds());
            serviceTemplate.getSpec().getContainer().setLivenessProbe(probe);
        }
    }

    /**
     * Update Environment variables.
     *
     * @param instanceName     Instance Name
     * @param serviceTemplate  service Template object
     * @param updatedComponent updated component to process env var
     */
    private void updateEnvVar(String instanceName, ServiceTemplate serviceTemplate, Component updatedComponent,
                              Map<?, ?> dependencyInfo) {
        Map<String, String> updatedParams = updatedComponent.getEnvVars();
        // Override values with updated values
        serviceTemplate.getSpec().getContainer().getEnv().forEach(envVar -> {
            if (updatedParams.containsKey(envVar.getName()) && !updatedParams.get(envVar.getName()).isEmpty()) {
                envVar.setValue(updatedParams.get(envVar.getName()));
            }
        });

        // Validate and replace dependency instance names
        serviceTemplate.getSpec().getContainer().getEnv().forEach(envVar -> {
            String value = envVar.getValue();
            if (value.isEmpty()) {
                printWarning("Value is empty for environment variable \"" + envVar.getName() + "\"");
                return;
            }
            envVar.setValue(value.replace(INSTANCE_NAME_PLACEHOLDER, instanceName));
            dependencyInfo.forEach((alias, info) -> {
                String aliasPlaceHolder = "{{" + alias + "}}";
                String depInstanceName = ((BString) ((BMap) info).getMap().get(INSTANCE_NAME)).stringValue();
                if (value.contains(aliasPlaceHolder)) {
                    envVar.setValue(value.replace(aliasPlaceHolder, depInstanceName));
                }
            });
        });
    }

    /**
     * Update the dependencies annotation with dependent instance names.
     *
     * @param cell           Cell definition
     * @param dependencyInfo dependency alias information map
     */
    private void updateDependencyAnnotations(Cell cell, Map dependencyInfo) {
        Gson gson = new Gson();
        Dependency[] dependencies = gson.fromJson(cell.getMetadata().getAnnotations()
                .get(ANNOTATION_CELL_IMAGE_DEPENDENCIES), Dependency[].class);
        Arrays.stream(dependencies).forEach(dependency -> {
            dependency.setInstance(((BString) ((BMap) dependencyInfo.get(dependency.getAlias())).getMap().get(
                    INSTANCE_NAME)).stringValue());
            dependency.setAlias(null);
        });
        cell.getMetadata().getAnnotations().put(ANNOTATION_CELL_IMAGE_DEPENDENCIES, gson.toJson(dependencies));
    }

    /**
     * Read the yaml and create a Cell object.
     *
     * @param destinationPath YAML path
     * @return Constructed Cell object
     */
    private Cell getInstance(String destinationPath) {
        Cell cell;
        try (InputStreamReader fileReader = new InputStreamReader(new FileInputStream(destinationPath),
                StandardCharsets.UTF_8)) {
            YamlReader reader = new YamlReader(fileReader);
            cell = reader.read(Cell.class);
        } catch (IOException e) {
            throw new BallerinaException("Unable to read Cell image file " + destinationPath + ". \nDid you " +
                    "pull/build the cell image ?");
        }
        if (cell == null) {
            throw new BallerinaException("Unable to extract Cell image from YAML " + destinationPath);
        }
        return cell;
    }

    /**
     * Add update components to CellImage.
     *
     * @param components component map
     */
    private void processComponents(BMap<?, ?> components) {
        components.getMap().forEach((key, componentValue) -> {
            Component component = new Component();
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            // Set mandatory fields.
            component.setName(((BString) attributeMap.get("name")).stringValue());

            //Process modifiable fields
            if (attributeMap.containsKey(PROBES)) {
                processProbes(((BMap<?, ?>) attributeMap.get(PROBES)).getMap(), component);
            }
            if (attributeMap.containsKey(INGRESSES)) {
                processIngress(((BMap<?, ?>) attributeMap.get(INGRESSES)).getMap(), component);
            }
            if (attributeMap.containsKey(ENV_VARS)) {
                processEnvVars(((BMap<?, ?>) attributeMap.get(ENV_VARS)).getMap(), component);
            }
            cellImage.addComponent(component);
        });
    }

    /**
     * Process and update Ingress of a component.
     *
     * @param ingressMap ingress attribute map
     * @param component  component to be updatedÒ
     */
    private void processIngress(LinkedHashMap<?, ?> ingressMap, Component component) {
        ingressMap.forEach((key, ingressValues) -> {
            BMap ingressValueMap = ((BMap) ingressValues);
            LinkedHashMap attributeMap = ingressValueMap.getMap();
            if ("WebIngress".equals(ingressValueMap.getType().getName())) {
                processWebIngress(component, attributeMap);
            }
        });
    }

    /**
     * Removes yaml tags.
     *
     * @param string generate yaml
     * @return yaml without tags
     */
    private String removeTags(String string) {
        //a tag is a sequence of characters starting with ! and ending with whitespace
        return removePattern(string, " ![^\\s]*");
    }

    /**
     * Create a secret file.
     *
     * @param instanceName    Cell Instance Name
     * @param data            secret data
     * @param destinationPath path to save the secret yaml
     */
    private void createSecret(String instanceName, Map<String, String> data, String destinationPath) {
        Secret secret = new SecretBuilder()
                .withNewMetadata()
                .withName(instanceName)
                .endMetadata()
                .withData(data)
                .build();
        try {
            writeToFile(SerializationUtils.dumpWithoutRuntimeStateAsYaml(secret), destinationPath);
        } catch (IOException e) {
            throw new BallerinaException("Error while generating secrets for instance " + instanceName);
        }
    }
}
