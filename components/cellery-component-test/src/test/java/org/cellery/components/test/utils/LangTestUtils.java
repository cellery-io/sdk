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
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import io.cellery.CelleryConstants;
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
import java.nio.file.StandardOpenOption;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

import static org.cellery.components.test.utils.CelleryTestConstants.ARTIFACTS;
import static org.cellery.components.test.utils.CelleryTestConstants.BAL;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_REPO_PATH;
import static org.cellery.components.test.utils.CelleryTestConstants.JSON;
import static org.cellery.components.test.utils.CelleryTestConstants.METADATA;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

/**
 * Language test utils.
 */
public class LangTestUtils {

    private static final Log log = LogFactory.getLog(LangTestUtils.class);
    private static final String JAVA_OPTS = "JAVA_OPTS";
    private static final Path DISTRIBUTION_PATH = Paths.get(FilenameUtils.separatorsToSystem(
            System.getProperty("ballerina.pack")));
    private static final String COMMAND =
            System.getProperty("os.name").toLowerCase(Locale.getDefault()).contains("win")
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
    public static int compileCellBuildFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo,
                                               Map<String, String> envVar) throws InterruptedException, IOException {

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
    public static int compileCellRunFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo,
                                             Map<String, String> envVar, Map<String, CellImageInfo> instanceData,
                                             String tmpDir)
            throws InterruptedException, IOException {
        envVar.put("CELLERY_IMAGE_DIR", tmpDir);
        return compileBallerinaFunction(RUN, sourceDirectory, fileName, cellImageInfo, instanceData, envVar);
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
    public static int compileCellRunFunction(Path sourceDirectory, String fileName, CellImageInfo cellImageInfo,
                                             Map<String, CellImageInfo> instanceData, String tmpDir)
            throws InterruptedException, IOException {

        return compileCellRunFunction(sourceDirectory, fileName, cellImageInfo, new HashMap<>(),
                instanceData, tmpDir);
    }

    private static int compileBallerinaFunction(String action, Path sourceDirectory, String fileName,
                                                CellImageInfo cellImageInfo, Map<String, CellImageInfo> cellInstances
            , Map<String, String> envVar) throws IOException, InterruptedException {

        Path ballerinaInternalLog = Paths.get(sourceDirectory.toAbsolutePath().toString(), "ballerina" +
                "-internal.log");
        if (ballerinaInternalLog.toFile().exists()) {
            log.warn("Deleting already existing ballerina-internal.log file.");
            FileUtils.deleteQuietly(ballerinaInternalLog.toFile());
        }

        Gson cellImgDataJSON = new GsonBuilder().create();
        String imgData = cellImgDataJSON.toJson(cellImageInfo);

        Gson dependencyJSON = new GsonBuilder().create();
        String instanceData = dependencyJSON.toJson(cellInstances);

        String balExecutable = createExecutableBalFiles(sourceDirectory, fileName, action);

        ProcessBuilder pb;
        if (action.equals(BUILD)) {
            pb = new ProcessBuilder(BALLERINA_COMMAND, RUN,
                    balExecutable, action, imgData, "{}");
        } else {
            pb = new ProcessBuilder(BALLERINA_COMMAND, RUN,
                    balExecutable, action, imgData, instanceData, "false", "false");
        }
        log.info(COMPILING + sourceDirectory.resolve(fileName).normalize());
        log.debug(EXECUTING_COMMAND + pb.command());
        pb.directory(sourceDirectory.toFile());
        Map<String, String> environment = pb.environment();
        addJavaAgents(environment);
        environment.putAll(envVar);

        Process process = pb.start();
        int exitCode = process.waitFor();

        boolean isIgnoreErr = false;
        if (exitCode > 0) {
            InputStream error = process.getErrorStream();
            StringBuilder errStr = new StringBuilder();
            int c;
            while ((c = error.read()) != -1) {
                errStr.append((char) c);
            }
            if (errStr.toString().contains("i/o timeout") ||
                    errStr.toString().contains("exited with exit code 127")) {
                exitCode = 0;
                isIgnoreErr = true;
            }
        }
        log.info(EXIT_CODE + exitCode);
        logOutput(process.getInputStream());
        if (!isIgnoreErr) {
            logOutput(process.getErrorStream());
        }

        Files.deleteIfExists(sourceDirectory.resolve(balExecutable));

        // log ballerina-internal.log content
        if (!isIgnoreErr && Files.exists(ballerinaInternalLog)) {
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

    public static Map<String, CellImageInfo> getDependencyInfo(Path source) throws IOException {

        String metadataJsonPath = source.toAbsolutePath().toString() + File.separator + TARGET
                + File.separator + CELLERY + File.separator + METADATA;
        Map<String, CellImageInfo> dependencyMap = new HashMap<>();
        try (InputStream input = new FileInputStream(metadataJsonPath)) {
            try (InputStreamReader inputStreamReader = new InputStreamReader(input)) {
                JsonElement parsedJson = new JsonParser().parse(inputStreamReader);
                JsonObject componentsJsonObject = parsedJson.getAsJsonObject().getAsJsonObject("components");
                for (Map.Entry<String, JsonElement> componentEntry : componentsJsonObject.entrySet()) {
                    final JsonObject dependencies = componentEntry.getValue().getAsJsonObject()
                            .getAsJsonObject("dependencies");
                    // Add component and composite dependencies
                    populateDependencyMap(dependencyMap, dependencies.getAsJsonObject("cells"));
                    populateDependencyMap(dependencyMap, dependencies.getAsJsonObject("composites"));
                }
            }
        }
        return dependencyMap;
    }

    private static void populateDependencyMap(Map<String, CellImageInfo> dependencyMap,
                                              JsonObject compositeDependencies) {
        for (Map.Entry<String, JsonElement> e : compositeDependencies.entrySet()) {
            JsonObject dependency = e.getValue().getAsJsonObject();
            String key = e.getKey();
            String org = dependency.getAsJsonPrimitive(CelleryConstants.ORG).getAsString();
            String name = dependency.getAsJsonPrimitive(CelleryConstants.NAME).getAsString();
            String ver = dependency.getAsJsonPrimitive(CelleryConstants.VERSION).getAsString();
            CellImageInfo cellImageInfo = new CellImageInfo(org, name, ver, "");
            dependencyMap.put(key, cellImageInfo);
        }
    }

    public static String createTempImageDir(Path sourceDir, String imageName) throws IOException {

        Path tmpDirPath = Files.createTempDirectory("cellery-sample");
        tmpDirPath.toFile().deleteOnExit();
        File source = new File(sourceDir.toString() + File.separator + TARGET + File.separator + CELLERY +
                File.separator + imageName + YAML);
        File sourceMeta = new File(sourceDir.toString() + File.separator + TARGET + File.separator + CELLERY +
                File.separator + imageName + "_meta" + JSON);
        File cellDir =
                new File(tmpDirPath.toString() + File.separator + ARTIFACTS + File.separator + CELLERY);
        cellDir.deleteOnExit();
        boolean folderCreated = cellDir.mkdirs();
        if (folderCreated) {
            File dest = new File(cellDir.toPath().toString() + File.separator + imageName + YAML);
            dest.deleteOnExit();
            File destMeta =
                    new File(cellDir.toPath().toString() + File.separator + imageName + "_meta" + JSON);
            dest.deleteOnExit();
            Files.copy(source.toPath(), dest.toPath());
            Files.copy(sourceMeta.toPath(), destMeta.toPath());
            return tmpDirPath.toString();
        } else {
            throw new IOException();
        }
    }

    private static void moveRefJsonToCelleryHome(Path sourceDirectory, CellImageInfo cellImageInfo) {
        Path targetPath = sourceDirectory.resolve("target");
        File destDir =
                new File(CELLERY_REPO_PATH + File.separator + cellImageInfo.getOrg() + File.separator +
                        cellImageInfo.getName() + File.separator + cellImageInfo.getVer());
        if (destDir.mkdirs()) {
            log.info("Created directory " + destDir);
        }
        ZipUtil.pack(new File(targetPath.toString()),
                new File(destDir.toPath() + File.separator + cellImageInfo.getName() + ".zip"),
                name -> "artifacts/" + name);
    }

    private static String createExecutableBalFiles(Path sourcePath, String fileName, String action) throws IOException {
        String executableBalName = fileName.replace(BAL, "") + "_" + action + BAL;
        Path targetDir = sourcePath.resolve(TARGET);
        if (!Files.exists(targetDir)) {
            Files.createDirectory(targetDir);
        }
        Path executableBalPath = targetDir.resolve(executableBalName);
        Files.copy(sourcePath.resolve(fileName), executableBalPath);
        String balMain;
        if (action.equals(BUILD)) {
            balMain = "\npublic function main(string action, cellery:ImageName iName, " +
                    "map<cellery:ImageName> " +
                    "instances) returns error? {\n" +
                    "\treturn build(iName);\n" +
                    "}\n";
        } else if (action.equals(RUN)) {
            balMain = "\npublic function main(string action, cellery:ImageName iName, " +
                    "map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) " +
                    "returns error? {\n" +
                    "\tcellery:InstanceState[]|error? result = " +
                    "run(iName, instances, startDependencies, shareDependencies);\n" +
                    "\tif (result is error?) {\n" +
                    "\t\treturn result;\n" +
                    "\t}" +
                    "}\n";
        } else {
            throw new IllegalArgumentException("Cell action is not supported");
        }
        Files.write(executableBalPath, balMain.getBytes(), StandardOpenOption.APPEND);
        return executableBalPath.toAbsolutePath().toString();
    }
}
