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

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.gson.Gson;
import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.exception.BallerinaCelleryException;
import io.cellery.models.Cell;
import io.cellery.models.Component;
import io.cellery.models.Composite;
import io.cellery.models.GatewaySpec;
import io.cellery.models.Meta;
import io.cellery.models.Node;
import io.cellery.models.OIDC;
import io.cellery.models.Tree;
import io.cellery.models.Web;
import io.cellery.models.internal.Dependency;
import io.cellery.models.internal.Image;
import io.cellery.models.internal.ImageComponent;
import io.cellery.util.KubernetesClient;
import io.fabric8.kubernetes.api.model.Probe;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import org.apache.commons.codec.binary.Base64;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.connector.api.BLangConnectorSPIUtil;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.model.types.BArrayType;
import org.ballerinalang.model.types.BMapType;
import org.ballerinalang.model.types.BTypes;
import org.ballerinalang.model.values.BBoolean;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Properties;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryUtils.appendToFile;
import static io.cellery.CelleryUtils.fileExists;
import static io.cellery.CelleryUtils.getBallerinaExecutablePath;
import static io.cellery.CelleryUtils.getDependentInstanceName;
import static io.cellery.CelleryUtils.getFilesByExtension;
import static io.cellery.CelleryUtils.getInstanceImageName;
import static io.cellery.CelleryUtils.isInstanceRunning;
import static io.cellery.CelleryUtils.printDebug;
import static io.cellery.CelleryUtils.printInfo;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.processProbes;
import static io.cellery.CelleryUtils.processResources;
import static io.cellery.CelleryUtils.processWebIngress;
import static io.cellery.CelleryUtils.randomString;
import static io.cellery.CelleryUtils.removePrefix;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.unzip;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * Native function cellery:createInstance.
 */
public class CreateInstance {
    private static final Logger log = LoggerFactory.getLogger(CreateInstance.class);
    private static Tree dependencyTree = new Tree();
    private Image image = new Image();
    private String instanceName;
    private Map dependencyInfo = new LinkedHashMap();
    private BValueArray bValueArray;
    private BMap<String, BValue> bmap;
    private AtomicLong runCount;
    private Map<String, Node<Meta>> uniqueInstances;
    private Map<String, Node<Meta>> queuedInstances;
    private Map dependencyTreeTable;
    private boolean shareDependencies;
    private boolean isRoot;

    public static void createInstance(MapValue image, MapValue iName, MapValue instances, boolean startDependencies,
                                      boolean shareDependencies) throws BallerinaCelleryException {
        uniqueInstances = new HashMap();
        queuedInstances = new HashMap();
        dependencyTreeTable = new HashMap<String, Node>();
        BArrayType bArrayType =
                new BArrayType(ctx.getProgramFile().getPackageInfo(CelleryConstants.CELLERY_PACKAGE).getTypeDefInfo(
                        CelleryConstants.INSTANCE_STATE_DEFINITION).typeInfo.getType());
        bValueArray = new BValueArray(bArrayType);
        runCount = new AtomicLong(0L);
        String instanceArg;
        boolean startDependencies = ctx.getBooleanArgument(0);
        shareDependencies = ctx.getBooleanArgument(1);
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        String cellName = ((BString) nameStruct.get("name")).stringValue();
        isRoot = ((BBoolean) nameStruct.get("isRoot")).booleanValue();
        instanceName = ((BString) nameStruct.get(CelleryConstants.INSTANCE_NAME)).stringValue();
        if (instanceName.isEmpty()) {
            instanceName = generateRandomInstanceName(((BString) nameStruct.get("name")).stringValue(),
                    ((BString) nameStruct.get("ver")).stringValue());
        }
        String cellImageDir = System.getenv(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR);
        if (cellImageDir == null) {
            try (InputStream inputStream = new FileInputStream(CelleryConstants.DEBUG_BALLERINA_CONF)) {
                Properties properties = new Properties();
                properties.load(inputStream);
                cellImageDir = properties.getProperty(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR).replaceAll("\"", "");
            } catch (IOException e) {
                ctx.setReturnValues(BLangVMErrors.createError(ctx, e.getMessage()));
                return;
            }
        }
        String destinationPath = cellImageDir + File.separator +
                "artifacts" + File.separator + "io/cellery";
        Map userDependencyLinks = ((BMap) ctx.getNullableRefArgument(2)).getMap();

        String cellYAMLPath = destinationPath + File.separator + cellName + CelleryConstants.YAML;
        Composite composite;
        if (CelleryConstants.CELL.equals(((BString) (refArgument.getMap().get("kind"))).stringValue())) {
            instanceArg = "cells.mesh.cellery.io/" + instanceName;
            composite = CelleryUtils.readCellYaml(cellYAMLPath);
        } else {
            instanceArg = "composites.mesh.cellery.io/" + instanceName;
            composite = CelleryUtils.readCompositeYaml(cellYAMLPath);
        }
        try {
            if (isRoot) {
                printInfo("Main Instance: " + instanceName);
                generateDependencyTree(destinationPath + File.separator + "metadata.json");
                // Validate main instance
                validateMainInstance(instanceName, ((Meta) dependencyTree.getRoot().getData()).getKind());
                printInfo("Validating dependencies");
                // Validate dependencies provided by user
                validateDependencyLinksAliasNames(userDependencyLinks);
                // Assign user defined instance names to dependent cells and create the finalized dependency tree
                finalizeDependencyTree(dependencyTree.getRoot(), userDependencyLinks);
                if (!startDependencies) {
                    // If not starting dependent instances, i.e all the dependent instances should be running
                    // Validate immediate dependency links
                    validateRootDependencyLinks(userDependencyLinks);
                }
                validateEnvironmentVariables();
                // Assign environment variables to dependent instances
                assignEnvironmentVariables(dependencyTree.getRoot());
                Meta rootMeta = (Meta) dependencyTree.getRoot().getData();
                BMap<String, BValue> rootCellInfo = new BMap<>(new BArrayType(BTypes.typeString));
                rootCellInfo.put("org", new BString(rootMeta.getOrg()));
                rootCellInfo.put("name", new BString(rootMeta.getName()));
                rootCellInfo.put("ver", new BString(rootMeta.getVer()));
                rootCellInfo.put("instanceName", new BString(rootMeta.getInstanceName()));
                rootCellInfo.put("isRoot", new BBoolean(true));
                bmap = BLangConnectorSPIUtil.createBStruct(ctx,
                        CelleryConstants.CELLERY_PACKAGE,
                        CelleryConstants.INSTANCE_STATE_DEFINITION,
                        rootCellInfo, isInstanceRunning(rootMeta.getInstanceName(), rootMeta.getKind()));

                bValueArray.add(runCount.getAndIncrement(), bmap);
                dependencyInfo = generateDependencyInfo();
                printInfo("Instances to be Used");
                displayDependentCellTable();
                printInfo("Dependency Tree to be Used\n");
                printDependencyTree();
                Map<String, String> testEnvVars = new HashMap<>();

                testEnvVars.put(CelleryConstants.IMAGE_NAME_ENV_VAR, rootCellInfo.toString()
                        .replace("org", "\"org\"")
                        .replace("name", "\"name\"")
                        .replace("ver", "\"ver\"")
                        .replace("instanceName", "\"instanceName\"")
                        .replace("isRoot", "\"isRoot\""));
                testEnvVars.put(CelleryConstants.DEPENDENCY_LINKS_ENV_VAR, dependencyInfo.toString()
                        .replace("org", "\"org\"")
                        .replace("name", "\"name\"")
                        .replace("ver", "\"ver\"")
                        .replace("instanceName", "\"instanceName\"")
                        .replace("isRoot", "\"isRoot\""));
                setEnvVarsForTests(testEnvVars);

                if (startDependencies) {
                    dependencyInfo.forEach((alias, info) -> {
                        String depInstanceName =
                                ((BString) ((BMap) info).getMap().get(CelleryConstants.INSTANCE_NAME)).stringValue();
                        String depKind = ((BString) ((BMap) info).getMap().get(CelleryConstants.KIND)).stringValue();
                        bmap = BLangConnectorSPIUtil.createBStruct(ctx,
                                CelleryConstants.CELLERY_PACKAGE,
                                CelleryConstants.INSTANCE_STATE_DEFINITION,
                                info, isInstanceRunning(depInstanceName, depKind), alias);
                        bValueArray.add(runCount.getAndIncrement(), bmap);
                    });
                    // Start the dependency tree
                    startDependencyTree(dependencyTree.getRoot());
                } else {
                    dependencyInfo.forEach((alias, info) -> {
                        bmap = BLangConnectorSPIUtil.createBStruct(ctx,
                                CelleryConstants.CELLERY_PACKAGE,
                                CelleryConstants.INSTANCE_STATE_DEFINITION,
                                info, true, alias);
                        bValueArray.add(runCount.getAndIncrement(), bmap);
                    });
                }
            } else {
                PrintStream out = System.out;
                out.println("starting instance " + instanceName + " ");
                // If not root instance, simply start the instance itself
                dependencyInfo = ((BMap) ctx.getNullableRefArgument(2)).getMap();
            }

        } catch (IOException | BallerinaException e) {
            ctx.setReturnValues(BLangVMErrors.createError(ctx, "Failed to create instance. " +
                    e.getMessage()));
            return;
        }
        updateDependencyAnnotations(composite, dependencyInfo);
        try {
            processComponents((BMap) refArgument.getMap().get(CelleryConstants.COMPONENTS));
            composite.getSpec().getComponents().forEach(component -> {
                String componentName = component.getMetadata().getName();
                ImageComponent updatedComponent = this.image.getComponentNameToComponentMap().get(componentName);
                //Replace env values defined in the YAML.
                updateEnvVar(instanceName, component, updatedComponent, dependencyInfo);
                // Update Gateway Config
                if (composite instanceof Cell) {
                    updateGatewayConfig(instanceName, destinationPath, (Cell) composite, updatedComponent);
                }
                // Update liveness and readiness probe
                updateProbes(component, updatedComponent);
                // Update resource limit and requests
                updateResources(component, updatedComponent);
                //update volume Instance name
                updateVolumeInstanceName(component, instanceName);
            });
            // Update cell yaml with instance name
            composite.getMetadata().setName(instanceName);
            writeToFile(toYaml(composite), cellYAMLPath);
            // Apply yaml file of the instance
            KubernetesClient.apply(cellYAMLPath);
            KubernetesClient.waitFor("Ready", 30 * 60, instanceArg, "default");
            ctx.setReturnValues(bValueArray);
        } catch (IOException | BallerinaException e) {
            String error = "Unable to persist updated composite yaml " + destinationPath;
            log.error(error, e);
            ctx.setReturnValues(BLangVMErrors.createError(ctx, error + ". " + e.getMessage()));
        } catch (Exception e) {
            String error = "Unable to apply composite yaml " + destinationPath;
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
                                     ImageComponent updatedComponent) {
        GatewaySpec gatewaySpec = cell.getSpec().getGateway().getSpec();
        Web web = updatedComponent.getWeb();
        if (web != null) {
            // Create TLS secret yaml and set the name
            if (StringUtils.isNoneEmpty(web.getTlsKey())) {
                Map<String, String> tlsMap = new HashMap<>();
                tlsMap.put("tls.key",
                        Base64.encodeBase64String(web.getTlsKey().getBytes(StandardCharsets.UTF_8)));
                tlsMap.put("tls.crt",
                        Base64.encodeBase64String(web.getTlsCert().getBytes(StandardCharsets.UTF_8)));
                String tlsSecretName = instanceName + "--tls-secret";
                createSecret(tlsSecretName, tlsMap, destinationPath + File.separator + tlsSecretName + ".yaml");
                gatewaySpec.getIngress().getExtensions().getClusterIngress().getTls().setSecret(tlsSecretName);
            }
            // Set OIDC values
            if (web.getOidc() != null) {
                OIDC oidc = web.getOidc();
                if (StringUtils.isBlank(oidc.getClientSecret()) && StringUtils.isBlank(oidc.getDcrPassword())) {
                    printWarning("OIDC client secret and DCR password are empty.");
                }
                gatewaySpec.getIngress().getExtensions().setOidc(oidc);
            }
        }
        if (web != null) {
            gatewaySpec.getIngress().getExtensions().getClusterIngress().setHost(web.getVhost());
        }

    }

    /**
     * Update Probe configurations.
     *
     * @param component        service Template object
     * @param updatedComponent updated component to process env var
     */
    private void updateProbes(Component component, ImageComponent updatedComponent) {
        Probe livenessProbe = updatedComponent.getLivenessProbe();
        // Override values with updated values
        component.getSpec().getTemplate().getContainers().forEach(container -> {
            if (container.getName().equals(updatedComponent.getName())) {
                if (livenessProbe != null) {
                    Probe probe = container.getLivenessProbe();
                    probe.setInitialDelaySeconds(livenessProbe.getInitialDelaySeconds());
                    probe.setFailureThreshold(livenessProbe.getFailureThreshold());
                    probe.setPeriodSeconds(livenessProbe.getPeriodSeconds());
                    probe.setSuccessThreshold(livenessProbe.getSuccessThreshold());
                    probe.setTimeoutSeconds(livenessProbe.getTimeoutSeconds());
                    container.setLivenessProbe(probe);
                }

                Probe readinessProbe = updatedComponent.getReadinessProbe();
                // Override values with updated values
                if (readinessProbe != null) {
                    Probe probe = container.getReadinessProbe();
                    probe.setInitialDelaySeconds(readinessProbe.getInitialDelaySeconds());
                    probe.setFailureThreshold(readinessProbe.getFailureThreshold());
                    probe.setPeriodSeconds(readinessProbe.getPeriodSeconds());
                    probe.setSuccessThreshold(readinessProbe.getSuccessThreshold());
                    probe.setTimeoutSeconds(readinessProbe.getTimeoutSeconds());
                    container.setReadinessProbe(probe);
                }
            }
        });

    }

    /**
     * Update Resource configurations.
     *
     * @param component        component object from YAML
     * @param updatedComponent updated component to process env var
     */
    private void updateResources(Component component, ImageComponent updatedComponent) {
        ResourceRequirements resourceRequirement = updatedComponent.getResources();
        component.getSpec().getTemplate().getContainers().forEach(container -> {
            if (container.getName().equals(updatedComponent.getName())) {
                container.setResources(resourceRequirement);
            }
        });

    }

    /**
     * Update Volume configurations.
     *
     * @param instanceName cell instance name
     * @param component    component object from YAML
     */
    private void updateVolumeInstanceName(Component component, String instanceName) {
        component.getSpec().getConfigurations().forEach(configMap -> {
            String name = configMap.getMetadata().getName();
            configMap.getMetadata().setName(name.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName));
        });
        component.getSpec().getVolumeClaims().forEach(volumeClaim -> {
            String name = volumeClaim.getName();
            volumeClaim.setName(name.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName));
            volumeClaim.getTemplate().getMetadata().setName(name.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER,
                    instanceName));
        });
        component.getSpec().getSecrets().forEach(secret -> {
            String name = secret.getMetadata().getName();
            secret.getMetadata().setName(name.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName));
        });
        component.getSpec().getTemplate().getVolumes().forEach(volume -> {
            String updatedName = volume.getName().replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName);
            volume.setName(updatedName);
            if (volume.getConfigMap() != null) {
                volume.getConfigMap().setName(updatedName);
            } else if (volume.getSecret() != null) {
                volume.getSecret().setSecretName(updatedName);
            } else {
                volume.getPersistentVolumeClaim().setClaimName(updatedName);
            }
        });

    }

    /**
     * Update Environment variables.
     *
     * @param instanceName     Instance Name
     * @param component        component object from YAML
     * @param updatedComponent updated component to process env var
     */
    private void updateEnvVar(String instanceName, Component component, ImageComponent updatedComponent,
                              Map<?, ?> dependencyInfo) {
        printDebug("[" + instanceName + "] cell instance [" + updatedComponent.getName() + "] component environment " +
                "variables:");
        Map<String, String> updatedParams = updatedComponent.getEnvVars();
        // Override values with updated values
        component.getSpec().getTemplate().getContainers().forEach(container -> {
            container.getVolumeMounts().forEach(volumeMount -> {
                String name = volumeMount.getName();
                volumeMount.setName(name.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName));
            });
            container.getEnv().forEach(envVar -> {
                if (updatedParams.containsKey(envVar.getName()) && !updatedParams.get(envVar.getName()).isEmpty()) {
                    envVar.setValue(updatedParams.get(envVar.getName()));
                }
                String value = envVar.getValue();
                if (value.isEmpty()) {
                    printWarning("Value is empty for environment variable \"" + envVar.getName() + "\"");
                    return;
                }
                // Validate and replace dependency instance names
                envVar.setValue(value.replace(CelleryConstants.INSTANCE_NAME_PLACEHOLDER, instanceName));
                dependencyInfo.forEach((alias, info) -> {
                    String aliasPlaceHolder = "{{" + alias + "}}";
                    String depInstanceName =
                            ((BString) ((BMap) info).getMap().get(CelleryConstants.INSTANCE_NAME)).stringValue();
                    if (value.contains(aliasPlaceHolder)) {
                        envVar.setValue(value.replace(aliasPlaceHolder, depInstanceName));
                    }
                });
                printDebug("\t" + envVar.getName() + "=" + envVar.getValue());
            });
        });
    }

    /**
     * Update the dependencies annotation with dependent instance names.
     *
     * @param composite      Cell definition
     * @param dependencyInfo dependency alias information map
     */
    private void updateDependencyAnnotations(Composite composite, Map dependencyInfo) {
        Gson gson = new Gson();
        Dependency[] dependencies = gson.fromJson(composite.getMetadata().getAnnotations()
                .get(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES), Dependency[].class);
        Arrays.stream(dependencies).forEach(dependency -> {
            dependency.setInstance(((BString) ((BMap) dependencyInfo.get(dependency.getAlias())).getMap().get(
                    CelleryConstants.INSTANCE_NAME)).stringValue());
            dependency.setAlias(null);
        });
        composite.getMetadata().getAnnotations().put(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES,
                gson.toJson(dependencies));
    }

    /**
     * Add update components to Image.
     *
     * @param components component map
     */
    private void processComponents(BMap<?, ?> components) {
        components.getMap().forEach((key, componentValue) -> {
            ImageComponent component = new ImageComponent();
            LinkedHashMap attributeMap = ((BMap) componentValue).getMap();
            // Set mandatory fields.
            component.setName(((BString) attributeMap.get("name")).stringValue());

            //Process modifiable fields
            if (attributeMap.containsKey(CelleryConstants.PROBES)) {
                processProbes(((BMap<?, ?>) attributeMap.get(CelleryConstants.PROBES)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.INGRESSES)) {
                processIngress(((BMap<?, ?>) attributeMap.get(CelleryConstants.INGRESSES)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.ENV_VARS)) {
                processEnvVars(((BMap<?, ?>) attributeMap.get(CelleryConstants.ENV_VARS)).getMap(), component);
            }
            if (attributeMap.containsKey(CelleryConstants.POD_RESOURCES)) {
                processResources(((BMap<?, ?>) attributeMap.get(CelleryConstants.POD_RESOURCES)).getMap(), component);
            }
            image.addComponent(component);
        });
    }

    /**
     * Process and update Ingress of a component.
     *
     * @param ingressMap ingress attribute map
     * @param component  component to be updatedÒ
     */
    private void processIngress(LinkedHashMap<?, ?> ingressMap, ImageComponent component) {
        ingressMap.forEach((key, ingressValues) -> {
            BMap ingressValueMap = ((BMap) ingressValues);
            LinkedHashMap attributeMap = ingressValueMap.getMap();
            if ("WebIngress".equals(ingressValueMap.getType().getName())) {
                processWebIngress(component, attributeMap);
            }
        });
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
            writeToFile(toYaml(secret), destinationPath);
        } catch (IOException e) {
            throw new BallerinaException("Error while generating secrets for instance " + instanceName);
        }
    }

    /**
     * Generate cell dependency information of the cell instance.
     *
     * @return map of dependency info
     */
    private Map generateDependencyInfo() {
        BMap<String, BValue> dependencyInfoMap = new BMap<>(new BMapType(new BArrayType(BTypes.typeString)));
        for (Map.Entry<String, Meta> dependentCell :
                ((Meta) dependencyTree.getRoot().getData()).getDependencies().entrySet()) {
            BMap<String, BValue> dependentCellMap = new BMap<>(new BArrayType(BTypes.typeString));
            dependentCellMap.put("org", new BString(dependentCell.getValue().getOrg()));
            dependentCellMap.put("name", new BString(dependentCell.getValue().getName()));
            dependentCellMap.put("ver", new BString(dependentCell.getValue().getVer()));
            dependentCellMap.put("instanceName", new BString(dependentCell.getValue().getInstanceName()));
            dependentCellMap.put("kind", new BString(dependentCell.getValue().getKind()));
            dependencyInfoMap.put(dependentCell.getKey(), dependentCellMap);
        }
        return ((BMap) dependencyInfoMap).getMap();
    }

    /**
     * Generate dependency tree.
     *
     * @param metadataJsonPath path to the metadata.json of the cell
     * @throws IOException if dependency tree generation fails
     */
    private void generateDependencyTree(String metadataJsonPath) throws IOException {
        //read json file data to String
        byte[] jsonData = Files.readAllBytes(Paths.get(metadataJsonPath));
        //create ObjectMapper instance
        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
        //convert json string to object
        Meta rootMeta = objectMapper.readValue(jsonData, Meta.class);
        Node<Meta> rootNode = new Node<>(rootMeta);
        rootMeta.setInstanceName(instanceName);
        // Set as root node
        dependencyTree.setRoot(rootNode);
        // Generating dependency tree
        buildDependencyTree(rootNode);
    }

    /**
     * Build dependency tree.
     *
     * @param node node that will be added to the tree
     */
    private void buildDependencyTree(Node<Meta> node) {
        Meta cell = node.getData();
        Map<String, Meta> dependentCells = cell.getDependencies();
        if (dependentCells.size() > 0) {
            for (Map.Entry<String, Meta> dependentCell : dependentCells.entrySet()) {
                dependentCell.getValue().setAlias(dependentCell.getKey());
                Node<Meta> childNode = node.addChild(dependentCell.getValue());
                buildDependencyTree(childNode);
            }
        }
        dependencyTree.addNode(node);
    }

    /**
     * Display cell dependency information table.
     */
    private void displayDependentCellTable() {
        ArrayList<Node> tree = new ArrayList<>();
        if (shareDependencies) {
            // Multiple instances with same cell image would have only one entry in the table
            for (Node node : dependencyTree.getTree()) {
                Meta meta = (Meta) node.getData();
                String cellImageName = meta.getOrg() + "/" + meta.getName() + ":" + meta.getVer();
                if (!dependencyTreeTable.containsKey(cellImageName)) {
                    dependencyTreeTable.put(cellImageName, node);
                    tree.add(node);
                }
            }
        } else {
            tree = dependencyTree.getTree();
        }
        PrintStream out = System.out;
        out.println("------------------------------------------------------------------------------------------------" +
                "------------------------");
        out.printf("%-30s %-35s %-25s %-15S %-15S", "INSTANCE NAME", "CELL IMAGE", "USED INSTANCE", "KIND", "SHARED");
        out.println();
        out.println("------------------------------------------------------------------------------------------------" +
                "------------------------");
        for (Node<Meta> node : tree) {
            Meta meta = node.getData();
            String image = meta.getOrg() + File.separator + meta.getName() + ":" + meta.getVer();
            String availability = "To be Created";
            if ((node.getData()).isRunning()) {
                availability = "Available in Runtime";
            }
            String shared = "-";
            if (node.getData().isShared()) {
                shared = "Shared";
            }
            out.format("%-30s %-35s %-25s %-15s %-15s", meta.getInstanceName(), image, availability,
                    meta.getKind(), shared);
            out.println();
        }
        out.println("------------------------------------------------------------------------------------------------" +
                "------------------------");
    }

    /**
     * Start a cell instance.
     *
     * @param org              organization
     * @param name             cell name
     * @param version          cell version
     * @param cellInstanceName cell instance name
     * @throws Exception if cell start fails
     */
    private void startInstance(String org, String name, String version, String cellInstanceName, String dependentCells,
                               boolean shareDependencies, Map<String, String> environmentVariables) throws IOException {
        Path imageDir = Paths.get(System.getProperty("user.home"), ".cellery", "repo", org, name, version,
                name + ".zip");
        if (!fileExists(imageDir.toString())) {
            pullImage(CelleryConstants.CENTRAL_REGISTRY_HOST, org, name, version);
        }
        Path tempDir = Paths.get(System.getProperty("user.home"), ".cellery", "tmp");
        Path tempBalFileDir = Files.createTempDirectory(Paths.get(tempDir.toString()), "cellery-cell-image");
        unzip(imageDir.toString(), tempBalFileDir.toString());
        String tempBalFile = getFilesByExtension(tempBalFileDir + File.separator + "src", "bal").
                get(0).toString();
        String ballerinaMain = "public function main(string action, cellery:ImageName iName, map<cellery:ImageName> " +
                "instances, boolean startDependencies, boolean shareDependencies) returns error? {\n" +
                "\tcellery:InstanceState[]|error? result = run(iName, instances, startDependencies, " +
                "shareDependencies);\n" +
                "\tif (result is error?) {\n" +
                "\t\treturn result;\n" +
                "\t}\n" +
                "}";
        appendToFile(ballerinaMain, tempBalFile);
        // Create a cell image json object
        JSONObject image = new JSONObject();
        image.put("org", org);
        image.put("name", name);
        image.put("ver", version);
        image.put("instanceName", cellInstanceName);
        Map<String, String> environment = new HashMap<>();
        environment.put(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR, tempBalFileDir.toString());
        String shareDependenciesFlag = "false";
        if (shareDependencies) {
            shareDependenciesFlag = "true";
        }
        for (Map.Entry<String, String> environmentVariable : environmentVariables.entrySet()) {
            environment.put(environmentVariable.getKey(), environmentVariable.getValue());
        }
        Path workingDir = Paths.get(System.getProperty("user.dir"));
        String exePath = getBallerinaExecutablePath();
        if (Files.exists(workingDir.resolve(CelleryConstants.BALLERINA_TOML))) {
            createTempDirForDependency(tempBalFile, System.getProperty("user.dir"), cellInstanceName);
            CelleryUtils.executeShellCommand(null, CelleryUtils::printInfo, CelleryUtils::printInfo,
                    environment, exePath + "ballerina", "run", cellInstanceName, "run", image.toString(),
                    dependentCells, "false", shareDependenciesFlag);
        } else {
            CelleryUtils.executeShellCommand(null, CelleryUtils::printInfo, CelleryUtils::printInfo,
                    environment, exePath + "ballerina", "run", tempBalFile, "run", image.toString(),
                    dependentCells, "false", shareDependenciesFlag);
        }
    }

    private void createTempDirForDependency(String tempBalFile, String workingDir, String instanceName)
            throws IOException {
        Path sourceFilePath = Paths.get(tempBalFile);
        Path module = Files.createDirectory(Paths.get(workingDir).resolve(instanceName));
        Files.copy(sourceFilePath, module.resolve(sourceFilePath.getFileName()), StandardCopyOption.REPLACE_EXISTING);
    }

    /**
     * Start the dependency tree.
     *
     * @param node node which will be used to initiate the starting of the dependency tree
     * @throws IOException if dependency tree start fails
     */
    private void startDependencyTree(Node<Meta> node) throws IOException {
        Meta meta = node.getData();
        for (Node<Meta> childNode : node.getChildren()) {
            startDependencyTree(childNode);
        }
        // Once all dependent cells are started or there are no dependent cells, start the current instance
        String cellMetaInstanceName = meta.getInstanceName();
        String cellImageName = meta.getOrg() + "/" + meta.getName() + ":" + meta.getVer();
        // Start the cell instance if not already running
        // This will not start root instance
        if (!dependencyTree.getRoot().equals(node) && !node.getData().isRunning()) {
            if (!queuedInstances.containsKey(cellImageName)) {
                JSONObject dependentCellsMap = new JSONObject();
                for (Map.Entry<String, Meta> dependentCell : meta.getDependencies().entrySet()) {
                    // Create a dependent cell image json object
                    JSONObject dependentCellImage = new JSONObject();
                    dependentCellImage.put("org", dependentCell.getValue().getOrg());
                    dependentCellImage.put("name", dependentCell.getValue().getName());
                    dependentCellImage.put("ver", dependentCell.getValue().getVer());
                    dependentCellImage.put("instanceName", dependentCell.getValue().getInstanceName());
                    dependentCellsMap.put(dependentCell.getKey(), dependentCellImage);
                }
                startInstance(meta.getOrg(), meta.getName(), meta.getVer(), cellMetaInstanceName,
                        dependentCellsMap.toString(), shareDependencies, meta.getEnvironmentVariables());
                queuedInstances.put(cellImageName, node);
            }
        }
    }

    /**
     * Assign names to dependent cells of a cell instance.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private void finalizeDependencyTree(Node<Meta> node, Map<?, ?> dependencyLinks) {
        String cellImage = node.getData().getOrg() + File.separator + node.getData().getName() + ":" +
                node.getData().getVer();
        // Instance names should be assigned from top to bottom
        if (node.equals(dependencyTree.getRoot())) {
            node.getData().setInstanceName(instanceName);
        } else if (node.getParent().getData().isRunning()) {
            // Even if the user has not given link to this instance, if the link of parent instance is given and it
            // is running that means the child instance is also running. Therefore get the child instance name using
            // the parent instance.
            node.getData().setInstanceName(getDependentInstanceName(node.getParent().getData().getInstanceName(),
                    node.getData().getOrg(), node.getData().getName(), node.getData().getVer(), node.getData().
                            getKind()));
            if (isInstanceRunning(node.getData().getInstanceName(), node.getData().getKind())) {
                node.getData().setRunning(true);
            }
        } else if (dependencyLinks.containsKey(node.getData().getAlias())) {
            // Check if there is a dependency Alias equal to the key of dependency cell
            // If there exists an alias assign its instance name as the dependent cell instance name
            String instanceName = ((BString) (((BMap) (dependencyLinks.get(node.getData().getAlias()))).getMap().
                    get(CelleryConstants.INSTANCE_NAME))).stringValue();
            String kind = node.getData().getKind();
            node.getData().setInstanceName(instanceName);
            if (isInstanceRunning(instanceName, kind)) {
                if (!getInstanceImageName(instanceName, node.getData().getKind()).equals(node.getData().getOrg() +
                        File.separator + node.getData().getName() + ":" + node.getData().getVer())) {
                    String errMsg = "Cell dependency validation failed. There already exists a cell instance with " +
                            "the instance name " + instanceName + "and cell image name " +
                            getInstanceImageName(instanceName, node.getData().getKind());
                    throw new BallerinaException(errMsg);
                }
                node.getData().setRunning(true);
            }
        } else {
            // Assign a random instance name if user has not defined an instance name
            node.getData().setInstanceName(generateRandomInstanceName(node.getData().getName(), node.getData().
                    getVer()));
        }
        if (shareDependencies) {
            if (uniqueInstances.containsKey(cellImage)) {
                // If there already is a cell image that means this is a shared instance
                uniqueInstances.get(cellImage).getData().setShared(true);
                node.getData().setShared(true);
                node.getData().setInstanceName(uniqueInstances.get(cellImage).getData().getInstanceName());
            } else {
                uniqueInstances.put(cellImage, node);
            }
        }
        for (Node<Meta> childNode : node.getChildren()) {
            finalizeDependencyTree(childNode, dependencyLinks);
        }
    }

    /**
     * Assign environment variables to dependent instances.
     *
     * @param node tree node
     */
    private void assignEnvironmentVariables(Node<Meta> node) {
        for (Map.Entry<String, String> environmentVariable : System.getenv().entrySet()) {
            if (environmentVariable.getKey().startsWith(CelleryConstants.CELLERY_ENV_VARIABLE + node.getData().getInstanceName() +
                    ".")) {
                node.getData().getEnvironmentVariables().put(removePrefix(environmentVariable.getKey(),
                        CelleryConstants.CELLERY_ENV_VARIABLE + node.getData().getInstanceName() + "."),
                        environmentVariable.
                                getValue());
            }
        }
        for (Node<Meta> childNode : node.getChildren()) {
            assignEnvironmentVariables(childNode);
        }
    }

    /**
     * Validate environment variables of dependent instances.
     */
    private void validateEnvironmentVariables() {
        for (Map.Entry<String, String> environmentVariable : System.getenv().entrySet()) {
            if (environmentVariable.getKey().startsWith(CelleryConstants.CELLERY_ENV_VARIABLE)) {
                boolean validInstance = false;
                String instanceName = removePrefix(environmentVariable.getKey().split("\\.")[0],
                        CelleryConstants.CELLERY_ENV_VARIABLE);
                // Environment variable key itself could contain "."
                String key = removePrefix(environmentVariable.getKey(), environmentVariable.getKey().
                        split("\\.")[0] + ".");
                for (Node<Meta> node : dependencyTree.getTree()) {
                    if (node.getData().getInstanceName().equals(instanceName)) {
                        validInstance = true;
                        if (node.getData().isRunning()) {
                            String errMsg = "Invalid environment variable, the instance of the environment should be " +
                                    "an instance to be created, instance " + node.getData().getInstanceName() + " is " +
                                    "already available in the runtime";
                            throw new BallerinaException(errMsg);
                        }
                    }
                }
                if (!validInstance) {
                    String errMsg = "Invalid environment variable, the instances of the environment variables should " +
                            "be provided as a dependency link, instance " + instanceName + " of the environment " +
                            "variable " + key + " not found";
                    throw new BallerinaException(errMsg);
                }
            }
        }
    }

    /**
     * Generate a random instance name.
     *
     * @param name cell name
     * @param ver  cell version
     * @return random instance name generated
     */
    private String generateRandomInstanceName(String name, String ver) {
        String instanceName = name + ver + randomString(4);
        return instanceName.replace(".", "");
    }

    private void setEnvVarsForTests(Map<String, String> envVars) {
        Map<String, String> env = System.getenv();
        try {
            Field field = env.getClass().getDeclaredField("m");
            field.setAccessible(true);
            ((Map<String, String>) field.get(env)).putAll(envVars);
            field.setAccessible(false);
        } catch (IllegalAccessException | NoSuchFieldException e) {
            throw new BallerinaException("Error occurred while creating " + CelleryConstants.BALLERINA_CONF);
        }
    }

    /**
     * Validate whether dependency aliases of dependent cells are correct.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private void validateDependencyLinksAliasNames(Map<?, ?> dependencyLinks) {
        ArrayList<String> invalidAliases = new ArrayList<>();
        dependencyLinks.forEach((alias, info) -> {
            boolean invalidAlias = true;
            // Iterate the dependency tree to check if the user given alias name exists
            for (Node<Meta> node : dependencyTree.getTree()) {
                if (node.getData().getDependencies().size() > 0 && node.getData().getDependencies().containsKey
                        (alias.toString())) {
                    invalidAlias = false;
                }
            }
            if (invalidAlias) {
                invalidAliases.add(alias.toString());
            }
        });
        if (invalidAliases.size() > 0) {
            String errMsg = "Cell dependency validation failed. Aliases " + String.join(", ", invalidAliases)
                    + " invalid";
            throw new BallerinaException(errMsg);
        }
    }

    /**
     * Validate whether all immediate dependencies of root have been defined when running without starting dependencies.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private void validateRootDependencyLinks(Map<?, ?> dependencyLinks) {
        ArrayList<String> missingAliases = new ArrayList<>();
        ArrayList<String> missingInstances = new ArrayList<>();
        for (Map.Entry<String, Meta> dependentCell : ((Meta) dependencyTree.getRoot().getData()).
                getDependencies().entrySet()) {
            if (!dependencyLinks.containsKey(dependentCell.getKey())) {
                missingAliases.add(dependentCell.getKey());
            }
            if (!dependentCell.getValue().isRunning()) {
                missingInstances.add(dependentCell.getValue().getInstanceName());
            }
        }
        if (missingAliases.size() > 0) {
            String errMsg = "Cell dependency validation failed. All links to dependent cells of instance " +
                    instanceName + " should be defined when running instance without starting dependencies. Missing " +
                    "dependency aliases: " + String.join(", ", missingAliases);
            throw new BallerinaException(errMsg);
        }
        if (missingInstances.size() > 0) {
            String errMsg = "Cell dependency validation failed. All immediate dependencies of root instance " +
                    instanceName + " should be available when running instance without starting dependencies. " +
                    "Missing instances: " + String.join(", ", missingInstances);
            throw new BallerinaException(errMsg);
        }
    }

    /**
     * Validate if there is another cell instance running with the same instance name.
     *
     * @param instanceName cell instance name
     */
    private void validateMainInstance(String instanceName, String kind) {
        if (isInstanceRunning(instanceName, kind)) {
            String errMsg = "instance to be created should not be present in the runtime, instance " + instanceName +
                    " is already available in the runtime";
            throw new BallerinaException(errMsg);
        }
    }

    /**
     * Pull cell image.
     *
     * @param registry name of the registry from which the cell is being pulled from
     * @param org      cell organization
     * @param name     cell name
     * @param version  cell version
     */
    private void pullImage(String registry, String org, String name, String version) {
        Map<String, String> environment = new HashMap<>();
        String image = registry + File.separator + org + File.separator + name + ":" + version;
        printInfo("Pulling image " + image);
        CelleryUtils.executeShellCommand(null, CelleryUtils::printInfoWithCarriageReturn,
                CelleryUtils::printInfoWithCarriageReturn, environment, "io/cellery", "pull", "--silent", image);
        printInfo("\nImage pull completed.");
    }

    /**
     * Print dependency tree node.
     *
     * @param buffer         string buffer
     * @param prefix         prefix
     * @param childrenPrefix child prefix
     * @param node           tree node
     */
    private void printDependencyTreeNode(StringBuilder buffer, String prefix, String childrenPrefix, Node<Meta> node) {
        buffer.append(prefix);
        if (!node.getData().getAlias().isEmpty()) {
            buffer.append(node.getData().getAlias() + ":" + node.getData().getInstanceName());
        } else {
            buffer.append(node.getData().getInstanceName());
        }
        buffer.append('\n');
        int index = 0;
        for (Node<Meta> child : node.getChildren()) {
            if (child.getChildren().size() > 0) {
                if (index == (node.getChildren().size() - 1)) {
                    printDependencyTreeNode(buffer, childrenPrefix + "├── ", childrenPrefix + "    ", child);
                } else {
                    printDependencyTreeNode(buffer, childrenPrefix + "├── ", childrenPrefix + "│   ", child);
                }
            } else {
                if (index == (node.getChildren().size() - 1)) {
                    printDependencyTreeNode(buffer, childrenPrefix + "└── ", childrenPrefix + "    ", child);
                } else {
                    printDependencyTreeNode(buffer, childrenPrefix + "├── ", childrenPrefix + "    ", child);
                }
            }
            index++;
        }
    }

    /**
     * Print dependency tree.
     */
    private void printDependencyTree() {
        PrintStream out = System.out;
        StringBuilder buffer = new StringBuilder(50);
        if (dependencyTree.getRoot().getChildren().size() > 0) {
            printDependencyTreeNode(buffer, "", "", dependencyTree.getRoot());
            out.println(buffer.toString());
        } else {
            out.println("No Dependencies");
        }
    }
}
