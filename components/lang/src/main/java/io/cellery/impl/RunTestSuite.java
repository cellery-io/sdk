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
package io.cellery.impl;

import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.models.Cell;
import io.cellery.models.CellSpec;
import io.cellery.models.Image;
import io.cellery.models.STSTemplate;
import io.cellery.models.STSTemplateSpec;
import io.cellery.models.ServiceTemplate;
import io.cellery.models.ServiceTemplateSpec;
import io.cellery.models.Test;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import org.apache.commons.lang3.StringUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BRefType;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;

import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_NAME;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_ORG;
import static io.cellery.CelleryConstants.ANNOTATION_CELL_IMAGE_VERSION;
import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.SERVICE_TYPE_JOB;
import static io.cellery.CelleryConstants.VERSION;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.getValidName;
import static io.cellery.CelleryUtils.printInfo;
import static io.cellery.CelleryUtils.printWarning;
import static io.cellery.CelleryUtils.toYaml;

/**
 * Native function cellery:runTestSuite.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "runTestSuite",
        args = {@Argument(name = "instanceList", type = TypeKind.ARRAY),
                @Argument(name = "testSuite", type = TypeKind.RECORD)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class RunTestSuite extends BlockingNativeCallableUnit {
    private static final String OUTPUT_DIRECTORY = System.getProperty("user.dir");
    private static final Logger log = LoggerFactory.getLogger(RunTestSuite.class);

    @Override
    public void execute(Context ctx) {
        LinkedHashMap nameStruct = null;
        BRefType<?>[] instanceList = ((BValueArray) ctx.getNullableRefArgument(0)).getValues();
        BRefType<?> iNameRefType;

        for (BRefType<?> refType: instanceList) {
            iNameRefType = (BRefType<?>) ((BMap) refType).getMap().get("iName");
            if (((BMap) iNameRefType).getMap().get(INSTANCE_NAME) == null) {
                break;
            }
            String alias = ((BMap) refType).getMap().get("alias").toString();
            if (!"".equals(alias)) {
                break;
            }
            nameStruct = ((BMap) iNameRefType).getMap();
        }
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(1);
        BRefType<?>[] tests = ((BValueArray) refArgument.getMap().get("tests")).getValues();
        try {
            if (nameStruct != null) {
                executeTests(tests, nameStruct);
            }
        } catch (Exception e) {
            ctx.setReturnValues(BLangVMErrors.createError(ctx, e.getMessage()));
        }
    }

    private void executeTests(BRefType<?>[] tests, LinkedHashMap nameStruct) {
        for (BRefType<?> refType : tests) {
            String name = ((BMap) refType).getMap().get(NAME).toString();
            String instanceName = ((BString) nameStruct.get(INSTANCE_NAME)).stringValue();
            if (name.isEmpty()) {
                break;
            }
            Test test = new Test();
            test.setName(name);

            LinkedHashMap sourceMap = ((BMap) ((BMap) refType).getMap().get("source")).getMap();
            if (sourceMap.get("image") == null) {
                test.setSource(sourceMap.get("filepath").toString());
                runInlineTest(instanceName);
            } else {
                test.setSource(sourceMap.get("image").toString());
                LinkedHashMap envMap = ((BMap) ((BMap) refType).getMap().get("envVars")).getMap();
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
    private boolean waitForPodRunning(String podName, String podInfo, String instanceName) throws
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
    private void waitForJobCompletion(String jobName, String podName, String instanceName) throws InterruptedException {
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
            printWarning("Error getting status of job " + jobName + ". Skipping collection of logs.");
        } else {
            printInfo("Test execution completed. Collecting logs to logs/" +
                    instanceName + ".log");
            CelleryUtils.executeShellCommand(
                    "kubectl logs " + podName + " " + instanceName + " > logs/" + instanceName + ".log", null,
                    CelleryUtils::printDebug, CelleryUtils::printWarning);
        }
    }

    private String getPodName(String podInfo, String instanceName) throws InterruptedException {
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
    private void deleteTestCell(String instanceName) {
        printInfo("Deleting test cell " + instanceName);
        CelleryUtils.executeShellCommand("kubectl delete cells.mesh.cellery.io " + instanceName, null,
                CelleryUtils::printDebug, CelleryUtils::printWarning);
    }

    private void runInlineTest(String module) {
        Path workingDir = Paths.get(System.getProperty("user.dir"));
        if (Files.notExists(workingDir.resolve("Ballerina.toml"))) {
            CelleryUtils.executeShellCommand("ballerina init", workingDir, CelleryUtils::printInfo,
                    CelleryUtils::printWarning);
        }

        if (Files.exists(workingDir.resolve(CelleryConstants.TEMP_TEST_MODULE))) {
            module = CelleryConstants.TEMP_TEST_MODULE;
        }

        String cmdArgs;
        if ("true".equalsIgnoreCase(System.getenv("INCELL"))) {
            cmdArgs = "--groups incell ";
        } else {
            cmdArgs = "--disable-groups incell ";
        }
        CelleryUtils.executeShellCommand("ballerina test " + cmdArgs + module, workingDir,
                CelleryUtils::printInfo, CelleryUtils::printWarning);
    }

    private Cell generateTestCell(Test test, LinkedHashMap nameStruct) {
        Image cellImage = new Image();
        cellImage.setCellName(test.getName());
        cellImage.setTest(test);
        cellImage.setOrgName(((BString) nameStruct.get(ORG)).stringValue());
        cellImage.setCellVersion(((BString) nameStruct.get(VERSION)).stringValue());

        List<ServiceTemplate> serviceTemplateList = new ArrayList<>();
        List<String> unsecuredPaths = new ArrayList<>();
        STSTemplate stsTemplate = new STSTemplate();
        STSTemplateSpec stsTemplateSpec = new STSTemplateSpec();
        ServiceTemplateSpec templateSpec = new ServiceTemplateSpec();

        List<EnvVar> envVarList = new ArrayList<>();
        cellImage.getTest().getEnvVars().forEach((key, value) -> {
            if (StringUtils.isEmpty(value)) {
                printWarning("Value is empty for environment variable \"" + key + "\"");
            }
            envVarList.add(new EnvVarBuilder().withName(key).withValue(value).build());
        });
        templateSpec.setContainer(new ContainerBuilder()
                .withImage(cellImage.getTest().getSource())
                .withEnv(envVarList)
                .build());
        templateSpec.setType(SERVICE_TYPE_JOB);
        ServiceTemplate serviceTemplate = new ServiceTemplate();
        serviceTemplate.setMetadata(new ObjectMetaBuilder()
                .withName(cellImage.getTest().getName())
                .withLabels(cellImage.getTest().getLabels())
                .build());
        serviceTemplate.setSpec(templateSpec);
        serviceTemplateList.add(serviceTemplate);
        stsTemplateSpec.setUnsecuredPaths(unsecuredPaths);
        stsTemplate.setSpec(stsTemplateSpec);

        CellSpec cellSpec = new CellSpec();
        cellSpec.setServicesTemplates(serviceTemplateList);
        cellSpec.setStsTemplate(stsTemplate);
        ObjectMeta objectMeta = new ObjectMetaBuilder().withName(getValidName(cellImage.getCellName()))
                .addToAnnotations(ANNOTATION_CELL_IMAGE_ORG, cellImage.getOrgName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_NAME, cellImage.getCellName())
                .addToAnnotations(ANNOTATION_CELL_IMAGE_VERSION, cellImage.getCellVersion())
                .build();
        return new Cell(objectMeta, cellSpec);
    }

    private void runImageBasedTest(Cell testCell, String testName) {
        String targetPath = Paths.get(OUTPUT_DIRECTORY, "cellery", testName + YAML).toString();
        try {

            CelleryUtils.writeToFile(toYaml(testCell), targetPath);
            CelleryUtils.writeToFile(toYaml(testCell), targetPath);
            CelleryUtils.printDebug("Creating test cell " + testName);
            CelleryUtils.executeShellCommand("kubectl apply -f " + targetPath, null,
                    CelleryUtils::printDebug, CelleryUtils::printWarning);
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

            CelleryUtils.printDebug("podName is: " + podName);
            CelleryUtils.printDebug("Waiting for pod " + podName + " status to be 'Running'...");

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
            throw new BallerinaException(errMsg);
        } catch (InterruptedException e) {
            log.error("Error while waiting for test completion. ", e.getMessage());
            throw new BallerinaException(e);
        }
    }
}
