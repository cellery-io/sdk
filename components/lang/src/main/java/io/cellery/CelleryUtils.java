/*
 * Copyright (c) 2018, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
 */
package io.cellery;

import com.esotericsoftware.yamlbeans.YamlWriter;
import io.cellery.models.API;
import io.cellery.models.Component;
import io.cellery.models.OIDC;
import io.cellery.models.Web;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.model.values.BInteger;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.io.PrintStream;
import java.io.StringWriter;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Set;
import java.util.stream.IntStream;

import static io.cellery.CelleryConstants.DEFAULT_PARAMETER_VALUE;
import static io.cellery.CelleryConstants.RESOURCES;
import static io.cellery.CelleryConstants.TARGET;

/**
 * Cellery Utility methods.
 */
public class CelleryUtils {

    /**
     * Returns swagger file as a String.
     *
     * @param path     swagger file path
     * @param encoding string encoding
     * @return swagger file as a String
     * @throws IOException if unable to read file
     */
    public static String readSwaggerFile(String path, Charset encoding) throws IOException {
        byte[] encoded = Files.readAllBytes(Paths.get(path));
        return new String(encoded, encoding);
    }

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
    public static void processWebIngress(Component component, LinkedHashMap attributeMap) {
        Web webIngress = new Web();
        LinkedHashMap gatewayConfig = ((BMap) attributeMap.get("gatewayConfig")).getMap();
        API httpAPI = getApi(component, attributeMap);
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
            if (StringUtils.isBlank(webIngress.getTlsKey())) {
                printWarning("TLS Key value is empty in component " + component.getName());
            }
            if (StringUtils.isBlank(webIngress.getTlsCert())) {
                printWarning("TLS Cert value is empty in component " + component.getName());
            }
        }
        if (gatewayConfig.containsKey("oidc")) {
            // OIDC enabled
            webIngress.setOidc(processOidc(((BMap) gatewayConfig.get("oidc")).getMap()));
        }
        component.addWeb(webIngress);
    }

    /**
     * Process API info and returns a API.
     *
     * @param component    component object
     * @param attributeMap API attribute map
     * @return API object
     */
    public static API getApi(Component component, LinkedHashMap attributeMap) {
        API httpAPI = new API();
        int containerPort = (int) ((BInteger) attributeMap.get("port")).intValue();
        // Validate the container port is same for all the ingresses.
        if (component.getContainerPort() > 0 && containerPort != component.getContainerPort()) {
            throw new BallerinaException("Invalid container port" + containerPort + ". Multiple container ports are " +
                    "not supported.");
        }
        component.setContainerPort(containerPort);
        return httpAPI;
    }

    /**
     * Process envVars and add to component.
     *
     * @param envVars   Map of EnvVars
     * @param component targetComponent
     */
    public static void processEnvVars(LinkedHashMap<?, ?> envVars, Component component) {
        envVars.forEach((k, v) -> {
            if (((BMap) v).getMap().get("value").toString().isEmpty()) {
                //value is empty for envVar
                component.addEnv(k.toString(), DEFAULT_PARAMETER_VALUE);
            } else {
                component.addEnv(k.toString(), ((BMap) v).getMap().get("value").toString());
            }
        });
    }

    /**
     * Process OIDCConfig.
     *
     * @param oidcConfig OIDC configuration
     */
    private static OIDC processOidc(LinkedHashMap oidcConfig) {
        OIDC oidc = new OIDC();
        oidc.setProviderUrl(((BString) oidcConfig.get("providerUrl")).stringValue());
        oidc.setRedirectUrl(((BString) oidcConfig.get("redirectUrl")).stringValue());
        oidc.setBaseUrl(((BString) oidcConfig.get("baseUrl")).stringValue());
        oidc.setClientId(((BString) oidcConfig.get("clientId")).stringValue());
        BValueArray nonSecurePaths = ((BValueArray) oidcConfig.get("nonSecurePaths"));
        Set<String> nonSecurePathList = new HashSet<>();
        IntStream.range(0, (int) nonSecurePaths.size()).forEach(nonSecurePathIndex ->
                nonSecurePathList.add(nonSecurePaths.getString(nonSecurePathIndex)));
        oidc.setNonSecurePaths(nonSecurePathList);

        BValueArray securePaths = ((BValueArray) oidcConfig.get("securePaths"));
        Set<String> securePathList = new HashSet<>();
        IntStream.range(0, (int) securePaths.size()).forEach(securePathIndex ->
                securePathList.add(securePaths.getString(securePathIndex)));
        oidc.setSecurePaths(securePathList);

        if (((BValue) oidcConfig.get("clientSecret")).getType().getName().equals("string")) {
            // Not using DCR
            oidc.setClientSecret(((BString) oidcConfig.get("clientSecret")).stringValue());
        } else {
            // Using DCR
            LinkedHashMap dcrConfig = ((BMap) oidcConfig.get("clientSecret")).getMap();
            oidc.setDcrUser(((BString) dcrConfig.get("dcrUser")).stringValue());
            oidc.setDcrPassword(((BString) dcrConfig.get("dcrPassword")).stringValue());
            if (dcrConfig.containsKey("dcrUrl")) {
                // DCR url is optional
                oidc.setDcrUrl(((BString) oidcConfig.get("dcrUrl")).stringValue());
            }
        }
        if (oidcConfig.containsKey("subjectClaim")) {
            //optional field
            oidc.setSubjectClaim(((BString) oidcConfig.get("subjectClaim")).stringValue());
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
     * Generates Yaml from a object.
     *
     * @param object Object
     * @param <T>    Any Object type
     * @return Yaml as a string.
     */
    public static <T> String toYaml(T object) {
        try (StringWriter stringWriter = new StringWriter()) {
            YamlWriter writer = new YamlWriter(stringWriter);
            writer.write(object);
            writer.getConfig().writeConfig.setWriteRootTags(false); //replaces only root tag
            writer.close(); //don't add this to finally, because the text will not be flushed
            return stringWriter.toString();
        } catch (IOException e) {
            throw new BallerinaException("Error occurred while generating yaml definition." + e.getMessage());
        }
    }

    /**
     * Print a Warring message.
     *
     * @param message warning message
     */
    public static void printWarning(String message) {
        PrintStream out = System.out;
        out.println("Warning: " + message);
    }

    /**
     * Copy file target/resources directory.
     *
     * @param sourcePath source file/directory path
     */
    public static void copyResourceToTarget(String sourcePath) {
        File src = new File(sourcePath);
        String targetPath = TARGET + File.separator + RESOURCES + File.separator + src.getName();
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
}
