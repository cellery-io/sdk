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
 */

package org.cellery.components.test.utils;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import org.apache.commons.io.FileUtils;
import org.apache.commons.io.FilenameUtils;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.cellery.components.test.models.CellImageInfo;
import org.zeroturnaround.zip.ZipUtil;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Properties;

import static org.cellery.components.test.utils.CelleryTestConstants.ARTIFACTS;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_REPO_PATH;
import static org.cellery.components.test.utils.CelleryTestConstants.DEPENDENCIES_PROPERTIES;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.TMP;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

/**
 * Language test utils.
 */
public class LangTestUtils {

    private static final Log log = LogFactory.getLog(LangTestUtils.class);
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
     * Compile and Executes the Build function of the Cell file with env variables.
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @param cellImageInfo   Information of the cell
     * @param envVar          environment variables required to build the cell
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileCellBuildFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo
            , Map<String, String> envVar) throws InterruptedException, IOException {

        return compileBallerinaFunction(BUILD, sourceDirectory, fileName, cellImageInfo, new HashMap<>(), envVar);
    }

    /**
     * Compile and Executes the build function of the Cell file.
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @param cellImageInfo   Information of the cell
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileCellBuildFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo)
            throws InterruptedException, IOException {

        return compileCellBuildFunction(sourceDirectory, fileName, cellImageInfo, new HashMap<>());
    }

    /**
     * Compile and Executes the run function of the Cell file with env variables.
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @param cellImageInfo   Information of the cell
     * @param envVar          environment variables required to run the cell
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileCellRunFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo
            , Map<String, String> envVar) throws InterruptedException, IOException {

        Map<String, CellImageInfo> instancesData = getDependancyInfo(sourceDirectory);
        String tmpDir = createTempImageDir(sourceDirectory, cellImageInfo.getName());
        envVar.put("CELLERY_IMAGE_DIR", tmpDir);
        return compileBallerinaFunction(RUN, sourceDirectory, fileName, cellImageInfo, instancesData, envVar);
    }

    /**
     * Compile and Executes the run function of the Cell file.
     *
     * @param sourceDirectory Ballerina source directory
     * @param fileName        Ballerina source file name
     * @param cellImageInfo   Information of the cell
     * @return Exit code
     * @throws InterruptedException if an error occurs while compiling
     * @throws IOException          if an error occurs while writing file
     */
    public static int compileCellRunFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo)
            throws InterruptedException, IOException {

        return compileCellRunFunction(sourceDirectory, fileName, cellImageInfo, new HashMap<>());
    }

    private static int compileBallerinaFunction(String action, Path sourceDirectory, String fileName
            , CellImageInfo cellImageInfo, Map<String, CellImageInfo> cellInstances
            , Map<String, String> envVar) throws IOException, InterruptedException {

        Path ballerinaInternalLog = Paths.get(sourceDirectory.toAbsolutePath().toString(), "ballerina-internal.log");
        if (ballerinaInternalLog.toFile().exists()) {
            log.warn("Deleting already existing ballerina-internal.log file.");
            FileUtils.deleteQuietly(ballerinaInternalLog.toFile());
        }

        Gson cellImgDataJSON = new GsonBuilder().create();
        String imgData = cellImgDataJSON.toJson(cellImageInfo);

        Gson dependencyJSON = new GsonBuilder().create();
        String instanceData = dependencyJSON.toJson(cellInstances);

        ProcessBuilder pb;
        if (action.equals(BUILD)) {
            pb = new ProcessBuilder(BALLERINA_COMMAND, RUN,
                    fileName + ':' + action, imgData);
        } else {
            pb = new ProcessBuilder(BALLERINA_COMMAND, RUN,
                    fileName + ':' + action, imgData, instanceData);
        }
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

        if (action.equals(BUILD) && exitCode == 0) {
            moveRefJsonToCelleryHome(sourceDirectory, cellImageInfo);
        }

        return exitCode;
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

    private static Map<String, CellImageInfo> getDependancyInfo(Path source) {

        String txtPath =
                source.toAbsolutePath().toString() + File.separator + TARGET + File.separator + TMP +
                        File.separator + DEPENDENCIES_PROPERTIES;
        Map<String, CellImageInfo> dependencyMap = new HashMap<>();
        try (InputStream input =
                     new FileInputStream(txtPath)) {
            Properties dependencyProperties = new Properties();
            dependencyProperties.load(input);

            for (Map.Entry<Object, Object> e : dependencyProperties.entrySet()) {
                String key = e.getKey().toString();
                String org = e.getValue().toString().split("/")[0];
                String name = e.getValue().toString().split("/")[1].split(":")[0];
                String ver = e.getValue().toString().split("/")[1].split(":")[1];
                CellImageInfo cell = new CellImageInfo(org, name, ver);
                dependencyMap.put(key, cell);
            }
        } catch (IOException ex) {
            log.debug("No dependencies available for the cell file");
        }
        return dependencyMap;
    }

    private static String createTempImageDir(Path sourceDir, String imageName) throws IOException {

        Path tmpDirPath = Files.createTempDirectory("cellery-sample");
        tmpDirPath.toFile().deleteOnExit();
        File source = new File(sourceDir.toString() + File.separator + TARGET + File.separator + CELLERY +
                File.separator + imageName + YAML);
        File cellDir = new File(tmpDirPath.toString() + File.separator + ARTIFACTS + File.separator + CELLERY);
        cellDir.deleteOnExit();
        boolean folderCreated = cellDir.mkdirs();
        if (folderCreated) {
            File dest = new File(cellDir.toPath().toString() + File.separator + imageName + YAML);
            dest.deleteOnExit();
            Files.copy(source.toPath(), dest.toPath());
            return tmpDirPath.toString();
        } else {
            throw new IOException();
        }
    }

    private static void moveRefJsonToCelleryHome(Path sourceDirectory, CellImageInfo cellImageInfo) throws IOException {
        Path targetPath = sourceDirectory.resolve("target");
        File destDir =
                new File(CELLERY_REPO_PATH + File.separator + cellImageInfo.getOrg() + File.separator +
                        cellImageInfo.getName() + File.separator + cellImageInfo.getVer());
        destDir.mkdirs();
        ZipUtil.pack(new File(targetPath.toString()),
                new File(destDir.toPath() + File.separator + cellImageInfo.getName() + ".zip"),
                name -> "artifacts/" + name);
    }
}
