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
import org.ballerinalang.jvm.BallerinaValues;
import org.ballerinalang.jvm.types.BArrayType;
import org.ballerinalang.jvm.types.BPackage;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.jvm.values.MapValueImpl;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.Properties;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryConstants.CELLERY;
import static io.cellery.CelleryConstants.CELLERY_PKG_NAME;
import static io.cellery.CelleryConstants.CELLERY_PKG_ORG;
import static io.cellery.CelleryConstants.CELLERY_PKG_VERSION;
import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.IS_ROOT;
import static io.cellery.CelleryConstants.KIND;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.VERSION;
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
import static io.cellery.CelleryUtils.processEnvVars;
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
    private static Image image = new Image();
    private static String instanceName;
    private static MapValue<?, ?> dependencyInfo;
    private static MapValue<String, Object> bmap;
    private static AtomicLong runCount;
    private static Map<String, Node<Meta>> dependencyTreeInstances;
    private static Map<String, Node<Meta>> dependencyTreeTable;
    private static boolean shareDependencies;

    public static ArrayValue createInstanceExternal(MapValue<?, ?> image, MapValue<?, ?> iName,
                                                    MapValue<?, ?> userDependencyLinks,
                                                    boolean startDependencies, boolean shareDependencies)
            throws BallerinaCelleryException {
        dependencyTreeInstances = new HashMap<>();
        dependencyTreeTable = new HashMap<>();
        ArrayValue bValueArray = new ArrayValue(new BArrayType(BallerinaValues.createRecordValue(new BPackage(
                        CELLERY_PKG_ORG, CELLERY_PKG_NAME, CELLERY_PKG_VERSION),
                CelleryConstants.INSTANCE_STATE_DEFINITION).getType()));
        runCount = new AtomicLong(0L);
        String instanceArg;
        CreateInstance.shareDependencies = shareDependencies;
        String cellName = iName.getStringValue("name");
        boolean isRoot = iName.getBooleanValue("isRoot");
        instanceName = iName.getStringValue(CelleryConstants.INSTANCE_NAME);
        if (instanceName.isEmpty()) {
            instanceName = generateRandomInstanceName(iName.getStringValue("name"), iName.getStringValue("ver"));
        }
        String cellImageDir = System.getenv(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR);
        if (cellImageDir == null) {
            String balConfPath = Paths.get(System.getProperty("user.dir"),
                    CelleryConstants.DEBUG_BALLERINA_CONF).toString();
            printDebug(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR +
                    " environment variable is not set. Reading value from " + balConfPath);
            try (InputStream inputStream = new FileInputStream(balConfPath)) {
                Properties properties = new Properties();
                properties.load(inputStream);
                cellImageDir = properties.getProperty(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR).replaceAll("\"",
                        "");
            } catch (IOException e) {
                throw new BallerinaCelleryException("unable to read Ballerina conf " + balConfPath);
            }
        }
        String destinationPath = cellImageDir + File.separator +
                "artifacts" + File.separator + CELLERY;

        String cellYAMLPath = destinationPath + File.separator + cellName + CelleryConstants.YAML;
        Composite composite;
        if (CelleryConstants.CELL.equals(image.getStringValue("kind"))) {
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
                validateMainInstance(instanceName, dependencyTree.getRoot().getData().getKind());
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
                Meta rootMeta = dependencyTree.getRoot().getData();
                MapValue<String, Object> rootCellInfo = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                        CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.IMAGE_NAME_DEFINITION);
                rootCellInfo.put(ORG, rootMeta.getOrg());
                rootCellInfo.put(NAME, rootMeta.getName());
                rootCellInfo.put(VERSION, rootMeta.getVer());
                rootCellInfo.put(INSTANCE_NAME, rootMeta.getInstanceName());
                rootCellInfo.put(IS_ROOT, true);
                bmap = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                        CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.INSTANCE_STATE_DEFINITION);
                bmap.put("iName", rootCellInfo);
                bmap.put("isRunning", isInstanceRunning(rootMeta.getInstanceName(), rootMeta.getKind()));
                bValueArray.add(runCount.getAndIncrement(), bmap);
                dependencyInfo = generateDependencyInfo();
                printInfo("Instances to be Used");
                displayDependentCellTable();
                printInfo("Dependency Tree to be Used\n");
                printDependencyTree();

                if (startDependencies) {
                    dependencyInfo.forEach((alias, info) -> {
                        final MapValue<?, ?> infoMap = (MapValue<?, ?>) info;
                        String depInstanceName = infoMap.getStringValue(CelleryConstants.INSTANCE_NAME);
                        String depKind = infoMap.getStringValue(CelleryConstants.KIND);
                        bmap = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                                CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.INSTANCE_STATE_DEFINITION);
                        bmap.put("iName", infoMap);
                        bmap.put("isRunning", isInstanceRunning(depInstanceName, depKind));
                        bmap.put("alias", alias);
                        bValueArray.add(runCount.getAndIncrement(), bmap);
                    });
                    // Start the dependency tree
                    startDependencyTree(dependencyTree.getRoot());
                } else {
                    dependencyInfo.forEach((alias, info) -> {
                        bmap = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                                CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.INSTANCE_STATE_DEFINITION);
                        bmap.put("iName", info);
                        bmap.put("isRunning", true);
                        bmap.put("alias", alias);
                        bValueArray.add(runCount.getAndIncrement(), bmap);
                    });
                }
            } else {
                PrintStream out = System.out;
                out.println("starting instance " + instanceName + " ");
                // If not root instance, simply start the instance itself
                dependencyInfo = userDependencyLinks;
            }

        } catch (IOException e) {
            throw new BallerinaCelleryException("Failed to create instance. " + e.getMessage());
        }
        updateDependencyAnnotations(composite, dependencyInfo);
        try {
            processComponents(image.getMapValue(CelleryConstants.COMPONENTS));
            composite.getSpec().getComponents().forEach(component -> {
                String componentName = component.getMetadata().getName();
                ImageComponent updatedComponent =
                        CreateInstance.image.getComponentNameToComponentMap().get(componentName);
                //Replace env values defined in the YAML.
                updateEnvVar(instanceName, component, updatedComponent, dependencyInfo);
                // Update Gateway Config
                if (composite instanceof Cell) {
                    try {
                        updateGatewayConfig(instanceName, destinationPath, (Cell) composite, updatedComponent);
                    } catch (BallerinaCelleryException e) {
                        printWarning("Unable to update gateway config " + e.getMessage());
                        return;
                    }
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
            return bValueArray;
        } catch (IOException e) {
            String error = "Unable to persist updated composite yaml " + destinationPath;
            log.error(error, e);
            throw new BallerinaCelleryException(error + ". " + e.getMessage());
        } catch (Exception e) {
            String error = "Unable to apply composite yaml " + destinationPath;
            log.error(error, e);
            throw new BallerinaCelleryException(error + ". " + e.getMessage());
        }
    }

    /**
     * Generate a random instance name.
     *
     * @param name cell name
     * @param ver  cell version
     * @return random instance name generated
     */
    private static String generateRandomInstanceName(String name, String ver) {
        String instanceName = name + "-" + ver + "-" + randomString(4);
        return instanceName.replace(".", "");
    }

    /**
     * Update Gateway Config.
     *
     * @param instanceName     instance Name
     * @param destinationPath  destination file path
     * @param cell             Cell object
     * @param updatedComponent Updated component object
     */
    private static void updateGatewayConfig(String instanceName, String destinationPath, Cell cell,
                                            ImageComponent updatedComponent) throws BallerinaCelleryException {
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
    private static void updateProbes(Component component, ImageComponent updatedComponent) {
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
    private static void updateResources(Component component, ImageComponent updatedComponent) {
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
    private static void updateVolumeInstanceName(Component component, String instanceName) {
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
    private static void updateEnvVar(String instanceName, Component component, ImageComponent updatedComponent,
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
                    String depInstanceName = ((MapValue<?, ?>) info).getStringValue(CelleryConstants.INSTANCE_NAME);
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
    private static void updateDependencyAnnotations(Composite composite, MapValue<?, ?> dependencyInfo) {
        Gson gson = new Gson();
        Dependency[] dependencies = gson.fromJson(composite.getMetadata().getAnnotations()
                .get(CelleryConstants.ANNOTATION_CELL_IMAGE_DEPENDENCIES), Dependency[].class);
        Arrays.stream(dependencies).forEach(dependency -> {
            dependency.setInstance(dependencyInfo.getMapValue(dependency.getAlias())
                    .getStringValue(CelleryConstants.INSTANCE_NAME));
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
    private static void processComponents(MapValue<?, ?> components) {
        components.forEach((key, componentValue) -> {
            ImageComponent component = new ImageComponent();
            MapValue<?, ?> attributeMap = (MapValue<?, ?>) componentValue;
            // Set mandatory fields.
            component.setName(attributeMap.getStringValue("name"));

            //Process modifiable fields
            if (attributeMap.containsKey(CelleryConstants.PROBES)) {
                processProbes(attributeMap.getMapValue(CelleryConstants.PROBES), component);
            }
            if (attributeMap.containsKey(CelleryConstants.INGRESSES)) {
                processIngress(attributeMap.getMapValue(CelleryConstants.INGRESSES), component);
            }
            if (attributeMap.containsKey(CelleryConstants.ENV_VARS)) {
                processEnvVars(attributeMap.getMapValue(CelleryConstants.ENV_VARS), component);
            }
            if (attributeMap.containsKey(CelleryConstants.POD_RESOURCES)) {
                processResources(attributeMap.getMapValue(CelleryConstants.POD_RESOURCES), component);
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
    private static void processIngress(MapValue<?, ?> ingressMap, ImageComponent component) {
        ingressMap.forEach((key, ingressValues) -> {
            MapValue<?, ?> ingressValueMap = ((MapValue<?, ?>) ingressValues);
            if ("WebIngress".equals(ingressValueMap.getType().getName())) {
                processWebIngress(component, ingressValueMap);
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
    private static void createSecret(String instanceName, Map<String, String> data, String destinationPath)
            throws BallerinaCelleryException {
        Secret secret = new SecretBuilder()
                .withNewMetadata()
                .withName(instanceName)
                .endMetadata()
                .withData(data)
                .build();
        try {
            writeToFile(toYaml(secret), destinationPath);
        } catch (IOException e) {
            throw new BallerinaCelleryException("Error while generating secrets for instance " + instanceName);
        }
    }

    /**
     * Generate dependency tree.
     *
     * @param metadataJsonPath path to the metadata.json of the cell
     * @throws IOException if dependency tree generation fails
     */
    private static void generateDependencyTree(String metadataJsonPath) throws IOException {
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
    private static void buildDependencyTree(Node<Meta> node) {
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
    private static void displayDependentCellTable() {
        ArrayList<Node<Meta>> tree = new ArrayList<>();
        if (shareDependencies) {
            // Multiple instances with same cell image would have only one entry in the table
            for (Node<Meta> node : dependencyTree.getTree()) {
                Meta meta = node.getData();
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
        out.println
                ("------------------------------------------------------------------------------------------------" +
                        "------------------------");
        out.printf("%-30s %-35s %-25s %-15S %-15S", "INSTANCE NAME", "CELL IMAGE", "USED INSTANCE", "KIND", "SHARED");
        out.println();
        out.println
                ("------------------------------------------------------------------------------------------------" +
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
        out.println
                ("------------------------------------------------------------------------------------------------" +
                        "------------------------");
    }

    /**
     * Start a cell instance.
     *
     * @param org              organization
     * @param name             cell name
     * @param version          cell version
     * @param cellInstanceName cell instance name
     * @throws IOException if cell start fails
     */
    private static void startInstance(String org, String name, String version, String cellInstanceName, String
            dependentCells, boolean shareDependencies, Map<String, String> environmentVariables) throws IOException {
        Path imageDir = Paths.get(System.getProperty("user.home"), ".cellery", "repo", org, name, version,
                name + ".zip");
        if (!fileExists(imageDir.toString())) {
            pullImage(org, name, version);
        }
        Path tempDir = Paths.get(System.getProperty("user.home"), ".cellery", "tmp");
        Path tempBalFileDir = Files.createTempDirectory(Paths.get(tempDir.toString()), "cellery-cell-image");
        unzip(imageDir.toString(), tempBalFileDir.toString());
        String tempBalFile = getFilesByExtension(tempBalFileDir + File.separator + "src", "bal").
                get(0).toString();
        String ballerinaMain = "public function main(string action, cellery:ImageName iName, map<cellery:ImageName> " +
                "instances, boolean startDependencies, boolean shareDependencies) returns error? {\n" +
                "\tif(action == \"build\") {\n" +
                "\t\treturn <@untainted> build(<@untainted> iName);\n" +
                "\t} else if (action == \"run\") {\n" +
                "\t\t cellery:InstanceState[]|error? result = run(<@untainted> iName, instances, startDependencies, " +
                "shareDependencies);\n" +
                "\t\tif (result is error?) {\n" +
                "\t\t\treturn <@untainted> result;\n" +
                "\t\t}\n" +
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

    private static void createTempDirForDependency(String tempBalFile, String workingDir, String instanceName)
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
    private static void startDependencyTree(Node<Meta> node) throws IOException {
        Meta meta = node.getData();
        for (Node<Meta> childNode : node.getChildren()) {
            startDependencyTree(childNode);
        }
        // Once all dependent cells are started or there are no dependent cells, start the current instance
        String cellMetaInstanceName = meta.getInstanceName();
        // Start the cell instance if not already running
        // This will not start root instance
        if (!dependencyTree.getRoot().equals(node) && !node.getData().isRunning()) {
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
            node.getData().setRunning(true);
        }
    }

    /**
     * Assign names to dependent cells of a cell instance.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private static void finalizeDependencyTree(Node<Meta> node, MapValue<?, ?> dependencyLinks) throws
            BallerinaCelleryException {
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
            String instanceName = dependencyLinks.getMapValue(node.getData().getAlias())
                    .getStringValue(CelleryConstants.INSTANCE_NAME);
            String kind = node.getData().getKind();
            node.getData().setInstanceName(instanceName);
            if (isInstanceRunning(instanceName, kind)) {
                if (!getInstanceImageName(instanceName, node.getData().getKind()).equals(node.getData().getOrg() +
                        File.separator + node.getData().getName() + ":" + node.getData().getVer())) {
                    String errMsg = "Cell dependency validation failed. There already exists a cell instance with " +
                            "the instance name " + instanceName + "and cell image name " +
                            getInstanceImageName(instanceName, node.getData().getKind());
                    throw new BallerinaCelleryException(errMsg);
                }
                node.getData().setRunning(true);
            }
        } else {
            // Assign a random instance name if user has not defined an instance name
            node.getData().setInstanceName(generateRandomInstanceName(node.getData().getName(), node.getData().
                    getVer()));
        }
        if (shareDependencies) {
            if (dependencyTreeInstances.containsKey(cellImage)) {
                // If there already is a cell image that means this is a shared instance
                dependencyTreeInstances.get(cellImage).getData().setShared(true);
                node.getData().setShared(true);
                node.getData().setInstanceName(dependencyTreeInstances.get(cellImage).getData().getInstanceName());
            } else {
                dependencyTreeInstances.put(cellImage, node);
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
    private static void assignEnvironmentVariables(Node<Meta> node) {
        for (Map.Entry<String, String> environmentVariable : System.getenv().entrySet()) {
            if (environmentVariable.getKey().startsWith(CelleryConstants.CELLERY_ENV_VARIABLE + node.getData()
                    .getInstanceName() +
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
    private static void validateEnvironmentVariables() throws BallerinaCelleryException {
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
                            throw new BallerinaCelleryException(errMsg);
                        }
                    }
                }
                if (!validInstance) {
                    String errMsg = "Invalid environment variable, the instances of the environment variables should "
                            + "be provided as a dependency link, instance " + instanceName + " of the environment " +
                            "variable " + key + " not found";
                    throw new BallerinaCelleryException(errMsg);
                }
            }
        }
    }

    /**
     * Validate whether dependency aliases of dependent cells are correct.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private static void validateDependencyLinksAliasNames(Map<?, ?> dependencyLinks) throws BallerinaCelleryException {
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
            throw new BallerinaCelleryException(errMsg);
        }
    }

    /**
     * Validate whether all immediate dependencies of root have been defined when running without starting
     * dependencies.
     *
     * @param dependencyLinks links to the dependent cells
     */
    private static void validateRootDependencyLinks(Map<?, ?> dependencyLinks) throws BallerinaCelleryException {
        ArrayList<String> missingAliases = new ArrayList<>();
        ArrayList<String> missingInstances = new ArrayList<>();
        for (Map.Entry<String, Meta> dependentCell : dependencyTree.getRoot().getData().
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
            throw new BallerinaCelleryException(errMsg);
        }
        if (missingInstances.size() > 0) {
            String errMsg = "Cell dependency validation failed. All immediate dependencies of root instance " +
                    instanceName + " should be available when running instance without starting dependencies. " +
                    "Missing instances: " + String.join(", ", missingInstances);
            throw new BallerinaCelleryException(errMsg);
        }
    }

    /**
     * Validate if there is another cell instance running with the same instance name.
     *
     * @param instanceName cell instance name
     */
    private static void validateMainInstance(String instanceName, String kind) throws BallerinaCelleryException {
        if (isInstanceRunning(instanceName, kind)) {
            String errMsg = "instance to be created should not be present in the runtime, instance " + instanceName +
                    " is already available in the runtime";
            throw new BallerinaCelleryException(errMsg);
        }
    }

    /**
     * Pull cell image.
     *
     * @param org     cell organization
     * @param name    cell name
     * @param version cell version
     */
    private static void pullImage(String org, String name, String version) {
        Map<String, String> environment = new HashMap<>();
        String image = CelleryConstants.CENTRAL_REGISTRY_HOST + File.separator + org + File.separator
                + name + ":" + version;
        printInfo("Pulling image " + image);
        CelleryUtils.executeShellCommand(null, CelleryUtils::printInfoWithCarriageReturn,
                CelleryUtils::printInfoWithCarriageReturn, environment, "cellery", "pull", "--silent", image);
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
    private static void printDependencyTreeNode(StringBuilder buffer, String prefix, String childrenPrefix, Node<Meta>
            node) {
        buffer.append(prefix);
        if (!node.getData().getAlias().isEmpty()) {
            buffer.append(node.getData().getAlias()).append(":").append(node.getData().getInstanceName());
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
    private static void printDependencyTree() {
        PrintStream out = System.out;
        StringBuilder buffer = new StringBuilder(50);
        if (dependencyTree.getRoot().getChildren().size() > 0) {
            printDependencyTreeNode(buffer, "", "", dependencyTree.getRoot());
            out.println(buffer.toString());
        } else {
            out.println("No Dependencies");
        }
    }

    /**
     * Generate cell dependency information of the cell instance.
     *
     * @return map of dependency info
     */
    private static MapValue<?, ?> generateDependencyInfo() {
        MapValue<String, Object> dependencyInfoMap = new MapValueImpl<>();
        for (Map.Entry<String, Meta> dependentCell : dependencyTree.getRoot().getData().getDependencies().entrySet()) {
            MapValue<String, String> dependentCellMap = new MapValueImpl<>();
            dependentCellMap.put(ORG, dependentCell.getValue().getOrg());
            dependentCellMap.put(NAME, dependentCell.getValue().getName());
            dependentCellMap.put(VERSION, dependentCell.getValue().getVer());
            dependentCellMap.put(INSTANCE_NAME, dependentCell.getValue().getInstanceName());
            dependentCellMap.put(KIND, dependentCell.getValue().getKind());
            dependencyInfoMap.put(dependentCell.getKey(), dependentCellMap);
        }
        return dependencyInfoMap;
    }
}
