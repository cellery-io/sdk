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
package io.cellery;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.Composite;
import io.cellery.models.Destination;
import io.cellery.models.OIDC;
import io.cellery.models.Port;
import io.cellery.models.Test;
import io.cellery.models.Web;
import io.cellery.models.internal.ImageComponent;
import io.cellery.models.internal.VolumeInfo;
import io.cellery.util.KubernetesClient;
import io.fabric8.kubernetes.api.model.ConfigMap;
import io.fabric8.kubernetes.api.model.ConfigMapBuilder;
import io.fabric8.kubernetes.api.model.HTTPGetActionBuilder;
import io.fabric8.kubernetes.api.model.HTTPHeader;
import io.fabric8.kubernetes.api.model.HTTPHeaderBuilder;
import io.fabric8.kubernetes.api.model.LabelSelector;
import io.fabric8.kubernetes.api.model.LabelSelectorRequirement;
import io.fabric8.kubernetes.api.model.LabelSelectorRequirementBuilder;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaim;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaimBuilder;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaimSpecBuilder;
import io.fabric8.kubernetes.api.model.Probe;
import io.fabric8.kubernetes.api.model.ProbeBuilder;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.QuantityBuilder;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import io.fabric8.kubernetes.client.utils.Serialization;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.bouncycastle.util.Strings;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.function.Consumer;
import java.util.stream.Collectors;
import java.util.stream.IntStream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

import static io.cellery.CelleryConstants.BTYPE_STRING;
import static io.cellery.CelleryConstants.DEFAULT_PARAMETER_VALUE;

/**
 * Cellery Utility methods.
 */
public class CelleryUtils {
    private static final String LOWER_CASE_ALPHA_NUMERIC_STRING = "0123456789abcdefghijklmnopqrstuvwxyz";
    private static SecureRandom random = new SecureRandom();

    /**
     * Returns valid kubernetes name.
     *
     * @param name actual value
     * @return valid name
     */
    public static String getValidName(String name) {
        return name.toLowerCase(Locale.getDefault()).replaceAll("\\P{Alnum}", "-");
    }


    /**
     * Process Web Ingress and add to component.
     *
     * @param component    Component
     * @param attributeMap WebIngress properties
     */
    public static void processWebIngress(ImageComponent component, MapValue attributeMap) {
        Web webIngress = new Web();
        MapValue gatewayConfig = attributeMap.getMapValue("gatewayConfig");
        API httpAPI = getApi(component, attributeMap);
        httpAPI.setGlobal(false);
        httpAPI.setPort(CelleryConstants.DEFAULT_GATEWAY_PORT);
        httpAPI.setContext(gatewayConfig.getStringValue("context"));
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
        webIngress.setHttpAPI(httpAPI);
        webIngress.setVhost(gatewayConfig.getStringValue("vhost"));
        if (gatewayConfig.containsKey("tls")) {
            // TLS enabled
            MapValue tlsConfig = gatewayConfig.getMapValue("tls");
            webIngress.setTlsKey(tlsConfig.getStringValue("key"));
            webIngress.setTlsCert(tlsConfig.getStringValue("cert"));
            if (StringUtils.isBlank(webIngress.getTlsKey())) {
                printWarning("TLS Key value is empty in component " + component.getName());
            }
            if (StringUtils.isBlank(webIngress.getTlsCert())) {
                printWarning("TLS Cert value is empty in component " + component.getName());
            }
        }
        if (gatewayConfig.containsKey("oidc")) {
            // OIDC enabled
            webIngress.setOidc(processOidc(gatewayConfig.getMapValue("oidc")));
        }
        component.setWeb(webIngress);
    }

    /**
     * Process API info and returns a API.
     *
     * @param component    component object
     * @param attributeMap API attribute map
     * @return API object
     */
    public static API getApi(ImageComponent component, MapValue attributeMap) {
        API httpAPI = new API();
        if (attributeMap.containsKey("apiVersion")) {
            httpAPI.setVersion(attributeMap.getStringValue("apiVersion"));
        }
        int containerPort = Math.toIntExact(attributeMap.getIntValue("port"));
        // Validate the container port is same for all the ingresses.
        if (component.getContainerPort() > 0 && containerPort != component.getContainerPort()) {
            throw new BallerinaException("Invalid container port" + containerPort + ". Multiple container ports are " +
                    "not supported.");
        }
        component.setContainerPort(containerPort);
        return httpAPI;
    }

    /**
     * Extract the Readiness Probe & Liveness Probe.
     *
     * @param probes    Scale policy to be processed
     * @param component current component
     */
    public static void processProbes(MapValue<?, ?> probes, ImageComponent component) {
        if (probes.containsKey(CelleryConstants.LIVENESS)) {
            MapValue livenessConf = probes.getMapValue(CelleryConstants.LIVENESS);
            component.setLivenessProbe(getProbe(livenessConf));
        }
        if (probes.containsKey(CelleryConstants.READINESS)) {
            MapValue readinessConf = probes.getMapValue(CelleryConstants.READINESS);
            component.setReadinessProbe(getProbe(readinessConf));
        }
    }

    /**
     * Extract the Resource limits and requests.
     *
     * @param resources Resource to be processed
     * @param component current component
     */
    public static void processResources(MapValue<?, ?> resources, ImageComponent component) {
        ResourceRequirements resourceRequirements = new ResourceRequirements();
        if (resources.containsKey(CelleryConstants.LIMITS)) {
            MapValue limitsConf = resources.getMapValue(CelleryConstants.LIMITS);
            resourceRequirements.setLimits(getResourceQuantityMap(limitsConf));
        }
        if (resources.containsKey(CelleryConstants.REQUESTS)) {
            MapValue requestConf = resources.getMapValue(CelleryConstants.REQUESTS);
            resourceRequirements.setRequests(getResourceQuantityMap(requestConf));
        }
        component.setResources(resourceRequirements);
    }

    public static void processVolumes(MapValue<?, ?> volumes, ImageComponent component) {
        volumes.forEach((key, volume) -> {
            MapValue volumeAttributes = (MapValue) volume;
            VolumeInfo volumeInfo = new VolumeInfo();
            volumeInfo.setPath(volumeAttributes.getStringValue("path"));
            volumeInfo.setReadOnly(volumeAttributes.getBooleanValue("readOnly"));
            MapValue k8sVolume = volumeAttributes.getMapValue("volume");
            switch (k8sVolume.getType().getName()) {
                case "K8sNonSharedPersistence":
                    volumeInfo.setShared(false);
                    PersistentVolumeClaim persistentVolumeClaim = getVolumeClaim(k8sVolume);
                    volumeInfo.setVolume(persistentVolumeClaim);
                    break;
                case "K8sSharedPersistence":
                    volumeInfo.setShared(true);
                    PersistentVolumeClaim pvShared =
                            new PersistentVolumeClaimBuilder().withNewMetadata()
                                    .withName(k8sVolume.getStringValue(CelleryConstants.NAME))
                                    .endMetadata()
                                    .build();
                    volumeInfo.setVolume(pvShared);
                    break;
                case "NonSharedConfiguration":
                    volumeInfo.setShared(false);
                    ConfigMap configMap = getConfigMap(k8sVolume);
                    volumeInfo.setVolume(configMap);
                    break;
                case "SharedConfiguration":
                    volumeInfo.setShared(true);
                    ConfigMap configMapShred =
                            new ConfigMapBuilder().withNewMetadata()
                                    .withName((k8sVolume.getStringValue(CelleryConstants.NAME)))
                                    .endMetadata()
                                    .build();
                    volumeInfo.setVolume(configMapShred);
                    break;
                case "SharedSecret":
                    volumeInfo.setShared(true);
                    Secret secretShared =
                            new SecretBuilder().withNewMetadata()
                                    .withName((k8sVolume.getStringValue(CelleryConstants.NAME)))
                                    .endMetadata()
                                    .build();
                    volumeInfo.setVolume(secretShared);
                    break;
                case "NonSharedSecret":
                    volumeInfo.setShared(false);
                    Secret secret = getSecret(k8sVolume);
                    volumeInfo.setVolume(secret);
                    break;
                default:
                    break;
            }
            component.addVolumeInfo(volumeInfo);
        });
    }

    public static Secret getSecret(MapValue k8sVolume) {
        return new SecretBuilder().withNewMetadata()
                .withName((k8sVolume.getStringValue(CelleryConstants.NAME)))
                .endMetadata()
                .withStringData(getDataMap((k8sVolume.getMapValue("data"))))
                .build();
    }

    public static ConfigMap getConfigMap(MapValue k8sVolume) {
        return new ConfigMapBuilder().withNewMetadata()
                .withName((k8sVolume.getStringValue(CelleryConstants.NAME)))
                .endMetadata()
                .withData(getDataMap((k8sVolume.getMapValue("data"))))
                .build();
    }

    public static PersistentVolumeClaim getVolumeClaim(MapValue k8sVolume) {
        PersistentVolumeClaimSpecBuilder specBuilder = new PersistentVolumeClaimSpecBuilder();
        Map<String, Quantity> requests = new HashMap<>();
        requests.put("storage", new QuantityBuilder()
                .withAmount(k8sVolume.getStringValue("request"))
                .build());
        specBuilder.withNewResources().withRequests(requests).endResources();
        if (k8sVolume.containsKey("storageClass")) {
            specBuilder.withStorageClassName(k8sVolume.getStringValue("storageClass"));
        }
        if (k8sVolume.containsKey("mode")) {
            specBuilder.withVolumeMode((k8sVolume.getStringValue("mode")));
        }
        if (k8sVolume.containsKey("accessMode")) {
            final ArrayValue accessModes = k8sVolume.getArrayValue("accessMode");
            List<String> accessModeList = new ArrayList<>();
            IntStream.range(0, accessModes.size()).forEach(accessModeIndex ->
                    accessModeList.add(accessModes.getRefValue(accessModeIndex).toString()));
            specBuilder.withAccessModes(accessModeList);
        }
        if (k8sVolume.containsKey("lookup")) {
            final MapValue lookupAttributes = k8sVolume.getMapValue("lookup");
            LabelSelector labelSelector = new LabelSelector();
            if (lookupAttributes.containsKey("labels")) {
                final MapValue labels = lookupAttributes.getMapValue("labels");
                labelSelector.setMatchLabels(getDataMap(labels));
            }
            if (lookupAttributes.containsKey("expressions")) {
                final ArrayValue expressions = lookupAttributes.getArrayValue("expressions");
                List<LabelSelectorRequirement> labelSelectorRequirements = new ArrayList<>();
                specBuilder.withNewSelector().withMatchExpressions();
                IntStream.range(0, expressions.size()).forEach(index -> {
                    MapValue expressionAttributes = ((MapValue) expressions.getRefValue(index));
                    final ArrayValue valueList = expressionAttributes.getArrayValue("values");
                    LabelSelectorRequirement labelSelectorRequirement = new LabelSelectorRequirementBuilder()
                            .withKey(expressionAttributes.getStringValue("key"))
                            .withOperator((expressionAttributes.getStringValue("operator")))
                            .withValues(Arrays.copyOfRange(valueList.getStringArray(), 0, valueList.size()))
                            .build();
                    labelSelectorRequirements.add(labelSelectorRequirement);
                });
                labelSelector.setMatchExpressions(labelSelectorRequirements);
            }
            specBuilder.withSelector(labelSelector);
        }
        return new PersistentVolumeClaimBuilder().withNewMetadata()
                .withName((k8sVolume.getStringValue(CelleryConstants.NAME)))
                .endMetadata().withSpec(specBuilder.build()).build();
    }

    /**
     * Get Resource Quantity Map.
     *
     * @param conf map of configurations
     * @return ResourceQuantityMap
     */
    private static Map<String, Quantity> getResourceQuantityMap(MapValue<String, ?> conf) {
        return conf.entrySet()
                .stream()
                .collect(Collectors.toMap(Map.Entry::getKey,
                        e -> new Quantity(e.getValue().toString()))
                );
    }

    /**
     * Get Data Quantity Map.
     *
     * @param conf map of configurations
     * @return ResourceQuantityMap
     */
    private static Map<String, String> getDataMap(MapValue<String, ?> conf) {
        return conf.entrySet()
                .stream()
                .collect(Collectors.toMap(Map.Entry::getKey,
                        e -> e.getValue().toString())
                );
    }

    /**
     * Create ProbeBuilder with given Liveness/Readiness Probe config.
     *
     * @param probeConf probeConfig map
     * @return ProbeBuilder
     */
    private static Probe getProbe(MapValue probeConf) {
        ProbeBuilder probeBuilder = new ProbeBuilder();
        final MapValue probeKindConf = probeConf.getMapValue(CelleryConstants.KIND);
        String probeKind = probeKindConf.getType().getName();
        if ("TcpSocket".equals(probeKind)) {
            probeBuilder.withNewTcpSocket()
                    .withNewPort(Math.toIntExact(probeKindConf.getIntValue("port")))
                    .endTcpSocket();
        } else if ("HttpGet".equals(probeKind)) {
            List<HTTPHeader> headers = new ArrayList<>();
            if (probeKindConf.containsKey("httpHeaders")) {
                probeKindConf.getMapValue("httpHeaders").forEach((key, value) -> {
                    HTTPHeader header = new HTTPHeaderBuilder()
                            .withName(key.toString())
                            .withValue(value.toString())
                            .build();
                    headers.add(header);
                });
            }
            probeBuilder.withHttpGet(new HTTPGetActionBuilder()
                    .withNewPort(Math.toIntExact(probeKindConf.getIntValue("port")))
                    .withPath(probeKindConf.getStringValue("path"))
                    .withHttpHeaders(headers)
                    .build()
            );
        } else {
            final ArrayValue commandList = probeKindConf.getArrayValue("commands");
            String[] commands = Arrays.copyOfRange(commandList.getStringArray(), 0, commandList.size());
            probeBuilder.withNewExec().addToCommand(commands).endExec();
        }
        return probeBuilder
                .withInitialDelaySeconds(Math.toIntExact(probeConf.getIntValue("initialDelaySeconds")))
                .withPeriodSeconds(Math.toIntExact(probeConf.getIntValue("periodSeconds")))
                .withFailureThreshold(Math.toIntExact(probeConf.getIntValue("failureThreshold")))
                .withTimeoutSeconds(Math.toIntExact(probeConf.getIntValue("timeoutSeconds")))
                .withSuccessThreshold(Math.toIntExact(probeConf.getIntValue("successThreshold"))).build();
    }


    /**
     * Process envVars and add to component.
     *
     * @param envVars   Map of EnvVars
     * @param component targetComponent
     */
    public static void processEnvVars(MapValue<?, ?> envVars, ImageComponent component) {
        envVars.forEach((k, v) -> {
            final String value = ((MapValue) v).get("value").toString();
            if (StringUtils.isEmpty(value)) {
                //value is empty for envVar
                component.addEnv(k.toString(), DEFAULT_PARAMETER_VALUE);
            } else {
                component.addEnv(k.toString(), value);
            }
        });
    }

    /**
     * Process envVars and add to test.
     *
     * @param envVars Map of EnvVars
     * @param test    targetComponent
     */
    public static void processEnvVars(MapValue<?, ?> envVars, Test test) {
        envVars.forEach((k, v) -> {
            if (StringUtils.isEmpty(((MapValue) v).getStringValue("value"))) {
                //value is empty for envVar
                test.addEnv(k.toString(), DEFAULT_PARAMETER_VALUE);
            } else {
                test.addEnv(k.toString(), ((MapValue) v).getStringValue("value"));
            }
        });
    }

    /**
     * Process OIDCConfig.
     *
     * @param oidcConfig OIDC configuration
     */
    private static OIDC processOidc(MapValue oidcConfig) {
        OIDC oidc = new OIDC();
        oidc.setProviderUrl(oidcConfig.getStringValue("providerUrl"));
        oidc.setRedirectUrl(oidcConfig.getStringValue("redirectUrl"));
        oidc.setBaseUrl(oidcConfig.getStringValue("baseUrl"));
        oidc.setClientId(oidcConfig.getStringValue("clientId"));
        ArrayValue nonSecurePaths = oidcConfig.getArrayValue("nonSecurePaths");
        Set<String> nonSecurePathList = new HashSet<>();
        IntStream.range(0, nonSecurePaths.size()).forEach(nonSecurePathIndex ->
                nonSecurePathList.add(nonSecurePaths.getString(nonSecurePathIndex)));
        oidc.setNonSecurePaths(nonSecurePathList);

        ArrayValue securePaths = oidcConfig.getArrayValue("securePaths");
        Set<String> securePathList = new HashSet<>();
        IntStream.range(0, securePaths.size()).forEach(securePathIndex ->
                securePathList.add(securePaths.getString(securePathIndex)));
        oidc.setSecurePaths(securePathList);

        if (BTYPE_STRING.equals(oidcConfig.getMapValue("clientSecret").getType().getName())) {
            // Not using DCR
            oidc.setClientSecret(oidcConfig.getStringValue("clientSecret"));
        } else {
            // Using DCR
            MapValue dcrConfig = oidcConfig.getMapValue("clientSecret");
            oidc.setDcrUser(dcrConfig.getStringValue("dcrUser"));
            oidc.setDcrPassword(dcrConfig.getStringValue("dcrPassword"));
            if (dcrConfig.containsKey("dcrUrl")) {
                // DCR url is optional
                oidc.setDcrUrl(oidcConfig.getStringValue("dcrUrl"));
            }
        }
        if (oidcConfig.containsKey("subjectClaim")) {
            //optional field
            oidc.setSubjectClaim(oidcConfig.getStringValue("subjectClaim"));
        }
        return oidc;
    }

    /**
     * Write content to a File. Create the required directories if they don't not exists.
     *
     * @param context    context of the file
     * @param targetPath target file path
     * @throws IOException If an error occurs when writing to a file
     */
    public static void writeToFile(String context, String targetPath) throws IOException {
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
     * Append content to a file.
     *
     * @param content    content to be added
     * @param targetPath path to the file
     */
    public static void appendToFile(String content, String targetPath) {
        try (BufferedWriter bufferedWriter = new BufferedWriter(new OutputStreamWriter(new FileOutputStream(targetPath,
                true), StandardCharsets.UTF_8))) {
            bufferedWriter.newLine();
            bufferedWriter.write(content);
        } catch (FileNotFoundException e) {
            throw new BallerinaException("Error getting file: " + targetPath + " " + e.getMessage());
        } catch (IOException e) {
            throw new BallerinaException("Error appending to file: " + targetPath + " " + e.getMessage());
        }
    }

    /**
     * Generates Yaml from a object.
     *
     * @param object Object
     * @param <T>    Any Object type
     * @return Yaml as a string.
     */
    public static <T> String toYaml(T object) {
        return Serialization.asYaml(object);
    }

    /**
     * Print a Warning message.
     *
     * @param message warning message
     */
    public static void printWarning(String message) {
        PrintStream out = System.out;
        out.println("Warning: " + message);
    }

    /**
     * Print a Info message.
     *
     * @param message info message
     */
    public static void printInfo(String message) {
        PrintStream out = System.out;
        out.println("Info: " + message);
    }

    public static void printInfoWithCarriageReturn(String message) {
        PrintStream out = System.out;
        out.print(message + "\r");
    }

    /**
     * Print a Debug message.
     *
     * @param message debug message
     */
    public static void printDebug(String message) {
        if ("true".equalsIgnoreCase(System.getenv("DEBUG_MODE"))) {
            PrintStream out = System.out;
            out.println("Debug: " + message);
        }
    }

    /**
     * Copy file target/resources directory.
     *
     * @param sourcePath source file/directory path
     */
    public static void copyResourceToTarget(String sourcePath) {
        File src = new File(sourcePath);
        String targetPath =
                CelleryConstants.TARGET + File.separator + CelleryConstants.RESOURCES + File.separator + src.getName();
        File dst = new File(targetPath);
        // if source is file
        try {
            if (Files.isRegularFile(Paths.get(sourcePath))) {
                if (Files.isDirectory(dst.toPath())) {
                    // if destination is directory
                    FileUtils.copyFileToDirectory(src, dst);
                } else {
                    // if destination is file
                    FileUtils.copyFile(src, dst);
                }
            } else if (Files.isDirectory(Paths.get(sourcePath))) {
                FileUtils.copyDirectory(src, dst);
            }

        } catch (IOException e) {
            throw new BallerinaException("Error occurred while copying resource file " + sourcePath +
                    ". " + e.getMessage());
        }
    }

    /**
     * Executes a shell command.
     *
     * @param command          command to execute
     * @param workingDirectory working directory
     * @param stdout           stdout of the command
     * @param stderr           stderr of the command
     * @return stdout/stderr
     */
    public static String executeShellCommand(String command, Path workingDirectory, Writer stdout, Writer stderr) {
        StringBuilder stdOut = new StringBuilder();
        StringBuilder stdErr = new StringBuilder();
        ProcessBuilder processBuilder = new ProcessBuilder("/bin/bash", "-c", command);

        ExecutorService executor = Executors.newFixedThreadPool(2);
        int exitCode;
        try {
            if (workingDirectory != null) {
                File workDirectory = workingDirectory.toFile();
                if (workDirectory.exists()) {
                    processBuilder.directory(workDirectory);
                }
            }
            Process process = processBuilder.start();

            StreamGobbler outputStreamGobbler = new StreamGobbler(process.getInputStream(), msg -> {
                stdOut.append(msg);
                stdout.writeMessage(msg);
            });
            StreamGobbler errorStreamGobbler = new StreamGobbler(process.getErrorStream(), msg -> {
                stdErr.append(msg);
                stderr.writeMessage(msg);
            });

            executor.execute(outputStreamGobbler);
            executor.execute(errorStreamGobbler);

            exitCode = process.waitFor();
            if (exitCode > 0) {
                throw new BallerinaException("Command " + command + " exited with exit code " + exitCode +
                        " message: " + stdErr.toString());
            }

        } catch (IOException e) {
            throw new BallerinaException(
                    "Error occurred while executing the command '" + command + "', " + "from directory '"
                            + Objects.requireNonNull(workingDirectory).toString(), e);
        } catch (InterruptedException e) {
            throw new BallerinaException(
                    "InterruptedException occurred while executing the command '" + command + "', " + "from directory '"
                            + Objects.requireNonNull(workingDirectory).toString(), e);
        } finally {
            executor.shutdownNow();
        }

        if (stdOut.toString().isEmpty()) {
            return stdErr.toString();
        }
        return stdOut.toString();
    }

    /**
     * Executes a shell command.
     *
     * @param command          command to execute
     * @param workingDirectory working directory
     * @param stdout           stdout of the command
     * @param stderr           stderr of the command
     * @return stdout/stderr
     */
    public static String executeShellCommand(Path workingDirectory, Writer stdout, Writer stderr,
                                             Map<String, String> environment, String... command) {
        StringBuilder stdOutBuilder = new StringBuilder();
        StringBuilder stdErrBuilder = new StringBuilder();

        // Set environment variables
        ProcessBuilder processBuilder = new ProcessBuilder(command);
        Map<String, String> processEnvironment = processBuilder.environment();
        for (Map.Entry<String, String> environmentEntry : environment.entrySet()) {
            processEnvironment.put(environmentEntry.getKey(), environmentEntry.getValue());
        }

        ExecutorService executor = Executors.newFixedThreadPool(2);
        int exitCode;
        try {
            if (workingDirectory != null) {
                File workDirectory = workingDirectory.toFile();
                if (workDirectory.exists()) {
                    processBuilder.directory(workDirectory);
                }
            }
            Process process = processBuilder.start();

            StreamGobbler outputStreamGobbler = new StreamGobbler(process.getInputStream(), msg -> {
                stdOutBuilder.append(msg);
                stdout.writeMessage(msg);
            });
            StreamGobbler errorStreamGobbler = new StreamGobbler(process.getErrorStream(), msg -> {
                stdErrBuilder.append(msg);
                stderr.writeMessage(msg);
            });

            executor.execute(outputStreamGobbler);
            executor.execute(errorStreamGobbler);

            exitCode = process.waitFor();
            if (exitCode > 0) {
                throw new BallerinaException("Command " + String.join(" ", command) +
                        " exited with exit code " + exitCode + " message: " + stdErrBuilder.toString());
            }

        } catch (IOException e) {
            throw new BallerinaException(
                    "Error occurred while executing the command '" + String.join(" ", command) + "', " +
                            "from directory '" + Objects.requireNonNull(workingDirectory).toString(), e);
        } catch (InterruptedException e) {
            throw new BallerinaException(
                    "InterruptedException occurred while executing the command '" + String.join(" ", command) +
                            "', " + "from directory '" + Objects.requireNonNull(workingDirectory).toString(), e);
        } finally {
            executor.shutdownNow();
        }

        if (stdOutBuilder.toString().isEmpty()) {
            return stdErrBuilder.toString();
        }
        return stdOutBuilder.toString();
    }

    /**
     * Read the yaml and create a Cell object.
     *
     * @param destinationPath YAML path
     * @return Constructed Cell object
     */
    public static Cell readCellYaml(String destinationPath) {
        Cell cell;
        try (FileInputStream fileInputStream = new FileInputStream(destinationPath)) {
            cell = Serialization.unmarshal(fileInputStream, Cell.class);
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
     * Read the yaml and create a Composite object.
     *
     * @param destinationPath YAML path
     * @return Constructed Composite object
     */
    public static Composite readCompositeYaml(String destinationPath) {
        Composite composite;
        try (FileInputStream fileInputStream = new FileInputStream(destinationPath)) {
            composite = Serialization.unmarshal(fileInputStream, Composite.class);
        } catch (IOException e) {
            throw new BallerinaException("Unable to read Cell image file " + destinationPath + ". \nDid you " +
                    "pull/build the cell image ?");
        }
        if (composite == null) {
            throw new BallerinaException("Unable to extract Cell image from YAML " + destinationPath);
        }
        return composite;
    }

    /**
     * Check if a cell instance is running.
     *
     * @param instance name of the instance
     * @return whether cell instance is running or note
     */
    public static boolean isInstanceRunning(String instance, String kind) {
        if (kind.equals(CelleryConstants.CELL)) {
            return !(KubernetesClient.getCells(instance).contains("not found"));
        } else {
            return !(KubernetesClient.getComposites(instance).contains("not found"));
        }
    }

    /**
     * Get the fully qualified image name of an instance.
     *
     * @param instance instance name
     * @param kind     instance kind
     * @return image name
     */
    public static String getInstanceImageName(String instance, String kind) {
        String image;
        if (kind.equals(CelleryConstants.CELL)) {
            image = KubernetesClient.getCells(instance);
        } else {
            image = KubernetesClient.getComposites(instance);
        }

        JsonObject imageJson = new Gson().fromJson(image, JsonObject.class);
        JsonObject cellAnnotations = imageJson.get("metadata").getAsJsonObject().get("annotations").
                getAsJsonObject();
        return cellAnnotations.get(CelleryConstants.MESH_CELLERY_IO + "/cell-image-org").getAsString() + File.separator
                + cellAnnotations.get(CelleryConstants.MESH_CELLERY_IO + "/cell-image-name").getAsString()
                + ":" + cellAnnotations.get(CelleryConstants.MESH_CELLERY_IO + "/cell-image-version").getAsString();
    }

    /**
     * Get dependent instance name using its image name.
     *
     * @param parentInstance   parent instance name
     * @param dependentOrg     dependent instance org
     * @param dependentName    dependent instance name
     * @param dependentVersion dependent instance version
     * @param dependentKind    dependent instance kind
     * @return dependent instance name
     */
    public static String getDependentInstanceName(String parentInstance, String dependentOrg, String dependentName,
                                                  String dependentVersion, String dependentKind) {
        String instanceName = "";
        String cellImage;
        if (dependentKind.equals(CelleryConstants.CELL)) {
            cellImage = KubernetesClient.getCells(parentInstance);
        } else {
            cellImage = KubernetesClient.getComposites(parentInstance);
        }
        JsonObject imageJson = new Gson().fromJson(cellImage, JsonObject.class);
        String cellDependenciesJson = imageJson.get("metadata").getAsJsonObject().get("annotations").
                getAsJsonObject().get(CelleryConstants.MESH_CELLERY_IO + "/cell-dependencies").getAsString();

        JsonArray cellDependencies = new JsonParser().parse(cellDependenciesJson).getAsJsonArray();

        for (JsonElement cellDependency : cellDependencies) {
            if (cellDependency.getAsJsonObject().get("org").getAsString().equals(dependentOrg) &&
                    cellDependency.getAsJsonObject().get("name").getAsString().equals(dependentName) &&
                    cellDependency.getAsJsonObject().get("version").getAsString().equals(dependentVersion)) {
                instanceName = cellDependency.getAsJsonObject().get("instance").getAsString();
            }
        }
        return instanceName;
    }

    /**
     * Extract a zip file to a given location.
     *
     * @param zipFilePath path to the zip file
     * @param destDir     location which the zip file will be extracted to
     */
    public static void unzip(String zipFilePath, String destDir) {
        PrintStream out = System.out;
        File dir = new File(destDir);
        // create output directory if it doesn't exist
        if (!dir.exists()) {
            boolean dirCreated = dir.mkdirs();
            if (!dirCreated && !dir.exists()) {
                out.println("Failed to create directory " + dir);
                throw new BallerinaException("Failed to create directory " + dir);
            }
        }
        //buffer for read and write data to file
        byte[] buffer = new byte[1024];
        try (FileInputStream fileInputStream = new FileInputStream(zipFilePath)) {
            if (!zipFilePath.isEmpty()) {
                try (ZipInputStream zipInputStream = new ZipInputStream(fileInputStream)) {
                    ZipEntry zipEntry = zipInputStream.getNextEntry();
                    while (zipEntry != null) {
                        String fileName = zipEntry.getName();
                        File newFile = new File(destDir + File.separator + fileName);
                        //create directories for sub directories in zip
                        File subDir = new File(newFile.getParent());
                        boolean dirCreated = subDir.mkdirs();
                        if (!dirCreated && !subDir.exists()) {
                            out.println("Failed to create sub directory " + subDir);
                            throw new BallerinaException("Failed to create sub directory " + subDir);
                        }
                        try (FileOutputStream fileOutputStream = new FileOutputStream(newFile)) {
                            int len;
                            while ((len = zipInputStream.read(buffer)) > 0) {
                                fileOutputStream.write(buffer, 0, len);
                            }
                            fileOutputStream.close();
                            zipInputStream.closeEntry();
                            zipEntry = zipInputStream.getNextEntry();
                        }
                    }
                    zipInputStream.closeEntry();
                    zipInputStream.close();
                    fileInputStream.close();
                }
            }
        } catch (IOException e) {
            throw new BallerinaException("Error while extracting file " + zipFilePath);
        }
    }

    /**
     * Generate a random string.
     *
     * @param len length of the random string to be generated
     * @return random string
     */
    public static String randomString(int len) {
        StringBuilder sb = new StringBuilder(len);
        for (int i = 0; i < len; i++) {
            sb.append(LOWER_CASE_ALPHA_NUMERIC_STRING.charAt(random.nextInt(LOWER_CASE_ALPHA_NUMERIC_STRING.length())));
        }
        return sb.toString();
    }

    /**
     * Get a list of files in a directory for a given extension.
     *
     * @param directory name of the directory
     * @param extension name of the extension
     * @return list of files
     */
    public static List<File> getFilesByExtension(String directory, String extension) {
        File dir = new File(directory);
        String[] extensions = new String[]{extension};
        return new ArrayList<>(FileUtils.listFiles(dir, extensions, true));
    }

    /**
     * Check if file exists.
     *
     * @param path Path to the file
     * @return whether file exists or not
     */
    public static boolean fileExists(String path) {
        File tmpDir = new File(path);
        return tmpDir.exists();
    }

    /**
     * Remove prefix from a string.
     *
     * @param s      string
     * @param prefix prefix
     * @return prefix removed string
     */
    public static String removePrefix(String s, String prefix) {
        return StringUtils.removeStart(s, prefix);
    }

    /**
     * Get ballerina installation directory.
     *
     * @return ballerina installation directory
     */
    private static String ballerinaInstallationDirectory() {
        String ballerinaHome = "";
        String osName = Strings.toLowerCase(System.getProperty("os.name"));
        if (osName.contains("mac")) {
            return CelleryConstants.BALLERINA_INSTALLATION_PATH_MAC;
        }
        if (osName.contains("nix") || osName.contains("nux") || osName.contains("aix")) {
            return CelleryConstants.BALLERINA_INSTALLATION_PATH_UBUNTU;
        }
        return ballerinaHome;
    }

    /**
     * Get ballerina executable path.
     *
     * @return ballerina executable path
     */
    public static String getBallerinaExecutablePath() {
        String exePath;
        Map<String, String> environment = new HashMap<>();
        String balVersionCmdOutput = CelleryUtils.executeShellCommand(null, msg -> {
                }, msg -> {
                }, environment,
                "ballerina", "version");
        if (balVersionCmdOutput.contains("Ballerina")) {
            String ballerinaVersion = balVersionCmdOutput.split(" ")[1];
            if (ballerinaVersion.equals(CelleryConstants.BALLERINA_VERSION)) {
                return "";
            }
        }
        exePath = ballerinaInstallationDirectory() + CelleryConstants.BALLERINA_EXECUTABLE_PATH;
        File dir = new File(exePath);
        if (!dir.exists()) {
            throw new BallerinaException("Ballerina executable path not found");
        }
        return exePath;
    }

    /**
     * Interface to print shell command output.
     */
    public interface Writer {

        /**
         * Called when a newline should be printed.
         *
         * @param msg message to write
         */
        void writeMessage(String msg);
    }

    /**
     * StreamGobbler to handle process builder output.
     */
    private static class StreamGobbler implements Runnable {
        private InputStream inputStream;
        private Consumer<String> consumer;

        StreamGobbler(InputStream inputStream, Consumer<String> consumer) {
            this.inputStream = inputStream;
            this.consumer = consumer;
        }

        @Override
        public void run() {
            new BufferedReader(new InputStreamReader(inputStream, StandardCharsets.UTF_8)).lines()
                    .forEach(consumer);
        }
    }
}
