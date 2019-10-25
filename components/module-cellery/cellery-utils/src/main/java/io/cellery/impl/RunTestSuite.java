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

import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.exception.BallerinaCelleryException;
import io.cellery.models.Cell;
import io.cellery.models.CellSpec;
import io.cellery.models.Component;
import io.cellery.models.ComponentSpec;
import io.cellery.models.Test;
import io.cellery.models.internal.Image;
import io.cellery.util.KubernetesClient;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.PodSpec;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.jvm.util.exceptions.BallerinaException;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.IOException;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import static io.cellery.CelleryConstants.CELLERY;
import static io.cellery.CelleryConstants.IMAGE_SOURCE;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.printDebug;
import static io.cellery.CelleryUtils.printInfo;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.toYaml;

/**
 * Native function cellery:runTestSuite.
 */
public class RunTestSuite {
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir");
    private static final Logger log = LoggerFactory.getLogger(RunTestSuite.class);

    public static void runTestSuite(ArrayValue instanceList, MapValue testSuite) throws BallerinaCelleryException {
        int bound = instanceList.size();
        for (int index = 0; index < bound; index++) {
            MapValue instance = (MapValue) instanceList.get(index);
            MapValue iName = instance.getMapValue("iName");
            String alias = iName.getStringValue("alias");
            if (StringUtils.isEmpty(alias)) {
                //Root instance alias is empty.
                ArrayValue tests = testSuite.getArrayValue("tests");
                executeTests(tests, instance);
            }
        }
    }

    private static void executeTests(ArrayValue tests, MapValue nameStruct) throws BallerinaCelleryException {
        int bound = tests.size();
        for (int index = 0; index < bound; index++) {
            final MapValue testInfo = (MapValue) tests.get(index);
            String name = testInfo.getStringValue(CelleryConstants.NAME);
            String instanceName = nameStruct.getStringValue(CelleryConstants.INSTANCE_NAME);
            Test test = new Test();
            test.setName(name);
            MapValue sourceMap = testInfo.getMapValue(IMAGE_SOURCE);
            if ("FileSource".equals(sourceMap.getType().getName())) {
                test.setSource(sourceMap.getStringValue("filepath"));
                runInlineTest(instanceName);
            } else {
                test.setSource(sourceMap.getStringValue("image"));
                MapValue envMap = testInfo.getMapValue("envVars");
                CelleryUtils.processEnvVars(envMap, test);
                Cell testCell = generateTestCell(test, nameStruct);
                runImageBasedTest(testCell, test.getName());
            }
        }
    }

    /**
     * Poll periodically for 10 minutes till the Pod reaches Running state.
     *
     * @param podName      name of the pod
     * @param podInfo      pod info string from shell command output
     * @param instanceName test cell name
     * @throws InterruptedException thrown if error occurs in Thread.sleep
     */
    private static boolean waitForPodRunning(String podName, String podInfo, String instanceName) throws
            InterruptedException {
        int min = 10;
        for (int i = 0; i < 12 * min; i++) {
            if (podName.isEmpty()) {
                podInfo = CelleryUtils.executeShellCommand("kubectl get pods | grep " + instanceName + "--" +
                                instanceName + "-job",
                        null, CelleryUtils::printDebug, CelleryUtils::printWarning);
                podName = podInfo.substring(0, podInfo.indexOf(' '));
            }
            if (!podInfo.contains("Running") && !podInfo.contains("Error") && !podInfo.contains("Completed")) {
                Thread.sleep(5000);
                podInfo = CelleryUtils.executeShellCommand("kubectl get pods | grep "
                                + instanceName + "--" + instanceName + "-job",
                        null, CelleryUtils::printDebug, CelleryUtils::printWarning);
            } else {
                return true;
            }
        }
        return false;
    }

    /**
     * Poll periodically for 1 minute till the job reaches Complete or Failed state.
     *
     * @param jobName      name of the job
     * @param podName      name of the pod
     * @param instanceName test cell name
     * @throws InterruptedException thrown if error occurs in Thread.sleep
     */
    private static void waitForJobCompletion(String jobName, String podName, String instanceName) throws
            InterruptedException {
        printInfo("Waiting for test job to complete...");
        String jobStatus = "";
        int min = 1;
        for (int i = 0; i < 12 * min; i++) {
            jobStatus = CelleryUtils.executeShellCommand("kubectl get jobs " + jobName + " " +
                            "-o jsonpath='{.status.conditions[?(@.type==\"Complete\")].status}'\n", null,
                    CelleryUtils::printDebug, CelleryUtils::printWarning);

            if (!"True".equalsIgnoreCase(jobStatus)) {
                jobStatus = CelleryUtils.executeShellCommand("kubectl get jobs " + jobName + " " +
                                "-o jsonpath='{.status.conditions[?(@.type==\"Failed\")].status}'\n", null,
                        CelleryUtils::printDebug, CelleryUtils::printWarning);
            }
            if ("True".equalsIgnoreCase(jobStatus)) {
                break;
            }
            Thread.sleep(5000);
        }
        if (!"True".equalsIgnoreCase(jobStatus)) {
            String podInfo = CelleryUtils.executeShellCommand("kubectl get pods " + podName,
                    null, CelleryUtils::printDebug, CelleryUtils::printWarning);
            if (podInfo.contains("CrashLoopBackOff")) {
                printWarning("Pod status turned to CrashLoopBackOff.");
            } else {
                printWarning("Error getting status of job " + jobName + ". Skipping collection of logs.");
                return;
            }
        }
        printInfo("Test execution completed. Collecting logs to logs/" +
                instanceName + ".log");
        CelleryUtils.executeShellCommand(
                "kubectl logs " + podName + " " + instanceName + " > logs/" + instanceName + ".log", null,
                CelleryUtils::printDebug, CelleryUtils::printWarning);
    }

    private static String getPodName(String podInfo, String instanceName) throws InterruptedException {
        String podName;
        int min = 1;
        for (int i = 0; i < 12 * min; i++) {
            if (podInfo.length() > 0) {
                podName = podInfo.substring(0, podInfo.indexOf(' '));
                return podName;
            } else {
                Thread.sleep(5000);
                podInfo = CelleryUtils.executeShellCommand("kubectl get pods | grep "
                                + instanceName + "--" + instanceName + "-job",
                        null, CelleryUtils::printDebug, CelleryUtils::printWarning);
            }
        }
        return null;
    }

    /**
     * Deletes the test cell.
     *
     * @param instanceName test cell name
     */
    private static void deleteTestCell(String instanceName) {
        printInfo("Deleting test cell " + instanceName);
        CelleryUtils.executeShellCommand("kubectl delete cells.mesh.cellery.io " + instanceName, null,
                CelleryUtils::printDebug, CelleryUtils::printWarning);
    }

    private static void runInlineTest(String module) throws BallerinaCelleryException {
        Path workingDir = Paths.get(System.getProperty("user.dir"));
        String srcDir = Paths.get(System.getenv(CelleryConstants.CELLERY_IMAGE_DIR_ENV_VAR), "src").toString();
        List<File> sourceBalList = CelleryUtils.getFilesByExtension(srcDir, "bal");
        if (!(sourceBalList.size() > 0)) {
            throw new BallerinaCelleryException("no bal files not found in " + srcDir);
        }
        String sourceBal = sourceBalList.get(0).toString();

        if (Files.exists(workingDir.resolve(CelleryConstants.TEMP_TEST_MODULE))) {
            module = CelleryConstants.TEMP_TEST_MODULE;
        }

        if (Paths.get(sourceBal).getFileName() != null) {
            Path sourcebalFileName = Paths.get(sourceBal).getFileName();
            Path destBalFilePath = workingDir.resolve(module).resolve(sourcebalFileName);

            List<File> destBalFileList = new ArrayList<>(FileUtils.listFiles(
                    workingDir.resolve(module).toFile(), new String[]{"bal"}, false));
            if (!(destBalFileList.size() > 0)) {
                try {
                    Files.copy(Paths.get(sourceBal), destBalFilePath, StandardCopyOption.REPLACE_EXISTING);
                } catch (IOException e) {
                    throw new BallerinaException(e);
                }
            } else {
                printDebug("Found bal file: " + destBalFilePath);
            }

        } else {
            String err = "Unable to find source bal file in " + srcDir;
            printWarning(err);
            throw new BallerinaException(err);
        }

        if (Files.notExists(workingDir.resolve("Ballerina.toml"))) {
            CelleryUtils.executeShellCommand("ballerina init", workingDir, CelleryUtils::printInfo,
                    CelleryUtils::printWarning);
        }

        CelleryUtils.executeShellCommand(workingDir, CelleryUtils::printInfo, CelleryUtils::printWarning, System
                .getenv(), "ballerina", "test", module);


    }

    private static Cell generateTestCell(Test test, MapValue nameStruct) {
        Image cellImage = new Image();
        cellImage.setCellName(test.getName());
        cellImage.setTest(test);
        cellImage.setOrgName(nameStruct.getStringValue(CelleryConstants.ORG));
        cellImage.setCellVersion(nameStruct.getStringValue(CelleryConstants.VERSION));

        List<Component> componentList = new ArrayList<>();

        List<EnvVar> envVarList = new ArrayList<>();
        cellImage.getTest().getEnvVars().forEach((key, value) -> {
            if (StringUtils.isEmpty(value)) {
                printWarning("Value is empty for environment variable \"" + key + "\"");
            }
            envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
        });
        PodSpec componentTemplate = new PodSpec();
        componentTemplate.setContainers(Collections.singletonList(new ContainerBuilder()
                .withImage(cellImage.getTest().getSource()).withEnv(envVarList)
                .withName(cellImage.getTest().getName()).build()));
        componentTemplate.setRestartPolicy("Never");
        componentTemplate.setShareProcessNamespace(true);

        ComponentSpec componentSpec = new ComponentSpec();
        componentSpec.setType(CelleryConstants.SERVICE_TYPE_JOB);
        componentSpec.setTemplate(componentTemplate);

        Component component = new Component();
        component.setMetadata(new ObjectMetaBuilder()
                .withName(cellImage.getTest().getName())
                .withLabels(cellImage.getTest().getLabels())
                .build());
        component.setSpec(componentSpec);
        componentList.add(component);

        CellSpec cellSpec = new CellSpec();
        cellSpec.setComponents(componentList);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(cellImage.getCellName()))
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_ORG, cellImage.getOrgName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_NAME, cellImage.getCellName())
                .addToAnnotations(CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION, cellImage.getCellVersion())
                .build();
        return new Cell(objectMeta, cellSpec);
    }

    private static void runImageBasedTest(Cell testCell, String testName) throws BallerinaCelleryException {
        String targetPath = Paths.get(OUTPUT_DIRECTORY, CELLERY, testName + CelleryConstants.YAML).toString();
        try {

            CelleryUtils.writeToFile(toYaml(testCell), targetPath);
            printDebug("Creating test cell " + testName);
            KubernetesClient.apply(targetPath);
            printInfo("Executing test " + testName + "...");

            // Wait for job to be available
            Thread.sleep(5000);

            String jobName = testName + "--" + testName + "-job";
            String podInfo = CelleryUtils.executeShellCommand("kubectl get pods | grep " + testName + "--" +
                            testName + "-job",
                    null, CelleryUtils::printDebug, CelleryUtils::printWarning);
            String podName = getPodName(podInfo, testName);
            if (podName == null) {
                printWarning("Error while getting name of the test pod. Skipping execution of test " + testName);
                return;
            }

            printDebug("podName is: " + podName);
            printDebug("Waiting for pod " + podName + " status to be 'Running'...");

            if (!waitForPodRunning(podName, podInfo, testName)) {
                printWarning("Error getting status of pod " + podName + ". Skipping execution of test " + testName);
                deleteTestCell(testName);
                return;
            }

            CelleryUtils.executeShellCommand("kubectl logs " + podName + " " + testName + " -f", null,
                    msg -> {
                        PrintStream out = System.out;
                        out.println("Log: " + msg);
                    }, CelleryUtils::printWarning);

            waitForJobCompletion(jobName, podName, testName);
            deleteTestCell(testName);
        } catch (IOException e) {
            String errMsg = "Error occurred while writing cell yaml " + targetPath;
            log.error(errMsg, e);
            throw new BallerinaCelleryException(errMsg);
        } catch (InterruptedException e) {
            String errMsg = "Error waiting for test completion. " + targetPath;
            log.error(errMsg, e.getMessage());
            throw new BallerinaCelleryException(errMsg);
        }
    }
}
