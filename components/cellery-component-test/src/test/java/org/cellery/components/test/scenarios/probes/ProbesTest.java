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
package org.cellery.components.test.scenarios.probes;

import io.cellery.CelleryUtils;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.ServiceTemplate;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.Probe;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.models.CellImageInfo;
import org.cellery.components.test.utils.LangTestUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.cellery.components.test.utils.CelleryTestConstants.ARTIFACTS;
import static org.cellery.components.test.utils.CelleryTestConstants.BAL;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_MESH_VERSION;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class ProbesTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("probes");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "probes", "1.0.0", "probe-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "probes" + BAL, cellImageInfo), 0);
        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        cell = CelleryUtils.readCellYaml(CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toString());
    }

    @Test(groups = "build")
    public void validateBuildTimeCellAvailability() {
        Assert.assertNotNull(cell);
    }

    @Test(groups = "build")
    public void validateBuildTimeAPIVersion() {
        Assert.assertEquals(cell.getApiVersion(), CELLERY_MESH_VERSION);
    }

    @Test(groups = "build")
    public void validateBuildTimeMetaData() {
        Assert.assertEquals(cell.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
    }

    @Test(groups = "build")
    public void validateBuildTimeGatewayTemplate() {
        final List<API> httpAPI = cell.getSpec().getGatewayTemplate().getSpec().getHttp();
        Assert.assertEquals(httpAPI.get(0).getBackend(), "employee");
        Assert.assertEquals(httpAPI.get(0).getContext(), "employee");
        Assert.assertEquals(httpAPI.get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(httpAPI.get(0).getDefinitions().get(0).getPath(), "/details");
        Assert.assertEquals(httpAPI.get(1).getBackend(), "salary");
        Assert.assertEquals(httpAPI.get(1).getContext(), "payroll");
        Assert.assertEquals(httpAPI.get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(httpAPI.get(1).getDefinitions().get(0).getPath(), "salary");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        final List<ServiceTemplate> servicesTemplates = cell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        final Container container = servicesTemplates.get(0).getSpec().getContainer();
        Assert.assertEquals(container.getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "{{instance_name}}--salary-service");
        Assert.assertEquals(container.getImage(), "docker.io/celleryio/sampleapp-employee");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getServicePort(), 80);
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getName(), "salary");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getPorts()
                .get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getServicePort(), 80);
    }

    @Test(groups = "build")
    public void validateBuildTimeProbes() {
        final List<ServiceTemplate> servicesTemplates = cell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        final Container empContainer = servicesTemplates.get(0).getSpec().getContainer();
        final Probe livenessProbe = empContainer.getLivenessProbe();
        Assert.assertNotNull(livenessProbe);
        Assert.assertEquals(livenessProbe.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(livenessProbe.getInitialDelaySeconds(), new Integer(30));
        Assert.assertEquals(livenessProbe.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(livenessProbe.getTimeoutSeconds(), new Integer(1));
        Assert.assertEquals(livenessProbe.getFailureThreshold(), new Integer(3));
        Assert.assertEquals(livenessProbe.getTcpSocket().getPort().getIntVal(), new Integer(8080));

        final Probe readinessProbe = empContainer.getReadinessProbe();
        Assert.assertNotNull(readinessProbe);
        Assert.assertEquals(readinessProbe.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(readinessProbe.getInitialDelaySeconds(), new Integer(10));
        Assert.assertEquals(readinessProbe.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(readinessProbe.getTimeoutSeconds(), new Integer(50));
        Assert.assertEquals(readinessProbe.getFailureThreshold(), new Integer(3));
        Assert.assertEquals(readinessProbe.getExec().getCommand().size(), 3);

        final Container salaryContainer = servicesTemplates.get(1).getSpec().getContainer();
        final Probe livenessProbeSalary = salaryContainer.getLivenessProbe();
        Assert.assertNotNull(livenessProbeSalary);
        Assert.assertEquals(livenessProbeSalary.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(livenessProbeSalary.getInitialDelaySeconds(), new Integer(30));
        Assert.assertEquals(livenessProbeSalary.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(livenessProbeSalary.getTimeoutSeconds(), new Integer(1));
        Assert.assertEquals(livenessProbeSalary.getFailureThreshold(), new Integer(3));
        Assert.assertEquals(livenessProbeSalary.getTcpSocket().getPort().getIntVal(), new Integer(8080));
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "probes" + BAL,
                cellImageInfo, dependencyCells, tmpDir), 0);
        File newYaml = tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.readCellYaml(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateRunTimeCellAvailability() {
        Assert.assertNotNull(runtimeCell);
    }

    @Test(groups = "run")
    public void validateRunTimeAPIVersion() {
        Assert.assertEquals(runtimeCell.getApiVersion(), CELLERY_MESH_VERSION);
    }

    @Test(groups = "run")
    public void validateRunTimeMetaData() {
        final ObjectMeta metadata = runtimeCell.getMetadata();
        Assert.assertEquals(metadata.getName(), cellImageInfo.getName());
        Assert.assertEquals(metadata.getAnnotations().get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(metadata.getAnnotations().get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(metadata.getAnnotations().get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeGatewayTemplate() {
        final List<API> httpList = runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp();
        Assert.assertEquals(httpList.get(0).getBackend(), "employee");
        Assert.assertEquals(httpList.get(0).getContext(), "employee");
        Assert.assertEquals(httpList.get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(httpList.get(0).getDefinitions().get(0).getPath(), "/details");
        Assert.assertEquals(httpList.get(1).getBackend(), "salary");
        Assert.assertEquals(httpList.get(1).getContext(), "payroll");
        Assert.assertEquals(httpList.get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(httpList.get(1).getDefinitions().get(0).getPath(), "salary");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        final List<ServiceTemplate> servicesTemplates = runtimeCell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        final Container container = servicesTemplates.get(0).getSpec().getContainer();
        Assert.assertEquals(container.getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "probe-inst--salary-service");
        Assert.assertEquals(container.getImage(), "docker.io/celleryio/sampleapp-employee");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getServicePort(), 80);
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getName(), "salary");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getPorts()
                .get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getServicePort(), 80);
    }


    @Test(groups = "run")
    public void validateRuntimeProbes() {
        final List<ServiceTemplate> servicesTemplates = runtimeCell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        final Container empContainer = servicesTemplates.get(0).getSpec().getContainer();
        final Probe livenessProbe = empContainer.getLivenessProbe();
        Assert.assertNotNull(livenessProbe);
        Assert.assertEquals(livenessProbe.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(livenessProbe.getInitialDelaySeconds(), new Integer(30));
        Assert.assertEquals(livenessProbe.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(livenessProbe.getTimeoutSeconds(), new Integer(1));
        Assert.assertEquals(livenessProbe.getFailureThreshold(), new Integer(5));
        Assert.assertEquals(livenessProbe.getTcpSocket().getPort().getIntVal(), new Integer(8080));

        final Probe readinessProbe = empContainer.getReadinessProbe();
        Assert.assertNotNull(readinessProbe);
        Assert.assertEquals(readinessProbe.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(readinessProbe.getInitialDelaySeconds(), new Integer(10));
        Assert.assertEquals(readinessProbe.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(readinessProbe.getTimeoutSeconds(), new Integer(50));
        Assert.assertEquals(readinessProbe.getFailureThreshold(), new Integer(3));
        Assert.assertEquals(readinessProbe.getExec().getCommand().size(), 3);

        final Container salaryContainer = servicesTemplates.get(1).getSpec().getContainer();
        final Probe livenessProbeSalary = salaryContainer.getLivenessProbe();
        Assert.assertNotNull(livenessProbeSalary);
        Assert.assertEquals(livenessProbeSalary.getPeriodSeconds(), new Integer(10));
        Assert.assertEquals(livenessProbeSalary.getInitialDelaySeconds(), new Integer(30));
        Assert.assertEquals(livenessProbeSalary.getSuccessThreshold(), new Integer(1));
        Assert.assertEquals(livenessProbeSalary.getTimeoutSeconds(), new Integer(1));
        Assert.assertEquals(livenessProbeSalary.getFailureThreshold(), new Integer(3));
        Assert.assertEquals(livenessProbeSalary.getTcpSocket().getPort().getIntVal(), new Integer(8080));
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
