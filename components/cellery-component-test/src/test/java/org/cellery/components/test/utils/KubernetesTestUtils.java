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
package org.cellery.components.test.utils;

import org.apache.commons.io.FileUtils;
import org.apache.commons.io.FilenameUtils;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

/**
 * Kubernetes test utils.
 */
public class KubernetesTestUtils {
    private static final Log log = LogFactory.getLog(KubernetesTestUtils.class);
    private static final String JAVA_OPTS = "JAVA_OPTS";
    private static final Path DISTRIBUTION_PATH = Paths.get(FilenameUtils.separatorsToSystem(
            System.getProperty("ballerina.pack")));
    private static final String COMMAND = System.getProperty("os.name").toLowerCase(Locale.getDefault()).contains("win")
            ? "ballerina.bat" : "ballerina";
    private static final String BALLERINA_COMMAND = DISTRIBUTION_PATH.resolve(COMMAND).toString();
    private static final String BUILD = "build";
    private static final String RUN = "run";
    private static final String EXECUTING_COMMAND = "Executing command: ";
    private static final String COMPILING = "Compiling: ";
    private static final String EXIT_CODE = "Exit code: ";

    private static void logOutput(InputStream inputStream) throws IOException {
        try (
                BufferedReader br = new BufferedReader(new InputStreamReader(inputStream))
        ) {
            br.lines().forEach(log::info);
        }
    }

    /**
     * Compile a ballerina file in a given directory
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileBallerinaFile(Path sourceDirectory, String fileName, String orgName, String imageName,
                                           String imageVersion, Map<String, String> envVar) throws InterruptedException,
            IOException {
        Path ballerinaInternalLog = Paths.get(sourceDirectory.toAbsolutePath().toString(), "ballerina-internal.log");
        if (ballerinaInternalLog.toFile().exists()) {
            log.warn("Deleting already existing ballerina-internal.log file.");
            FileUtils.deleteQuietly(ballerinaInternalLog.toFile());
        }

        ProcessBuilder pb = new ProcessBuilder(BALLERINA_COMMAND, RUN,
                fileName + ':' + BUILD, orgName, imageName, imageVersion);
        log.info(COMPILING + sourceDirectory.resolve(fileName).normalize());
        log.debug(EXECUTING_COMMAND + pb.command());
        pb.directory(sourceDirectory.toFile());
        Map<String, String> environment = pb.environment();
        addJavaAgents(environment);
        environment.putAll(envVar);

        Process process = pb.start();
        int exitCode = process.waitFor();
        log.info(EXIT_CODE + exitCode);
        logOutput(process.getInputStream());
        logOutput(process.getErrorStream());

        // log ballerina-internal.log content
        if (Files.exists(ballerinaInternalLog)) {
            log.error("ballerina-internal.log file found. content: ");
            log.error(FileUtils.readFileToString(ballerinaInternalLog.toFile()));
        }
        return exitCode;
    }

    /**
     * Compile a ballerina file in a given directory
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileBallerinaFile(Path sourceDirectory, String fileName, String orgName, String imageName,
                                           String imageVersion) throws InterruptedException, IOException {
        return compileBallerinaFile(sourceDirectory, fileName, orgName, imageName, imageVersion, new HashMap<>());
    }


    private static synchronized void addJavaAgents(Map<String, String> envProperties) {
        String javaOpts = "";
        if (envProperties.containsKey(JAVA_OPTS)) {
            javaOpts = envProperties.get(JAVA_OPTS);
        }
        if (javaOpts.contains("jacoco.agent")) {
            return;
        }
        javaOpts = getJacocoAgentArgs() + javaOpts;
        envProperties.put(JAVA_OPTS, javaOpts);
    }

    private static String getJacocoAgentArgs() {
        String jacocoArgLine = System.getProperty("jacoco.agent.argLine");
        if (jacocoArgLine == null || jacocoArgLine.isEmpty()) {
            log.warn("Running integration test without jacoco test coverage");
            return "";
        }
        return jacocoArgLine + " ";
    }
}
