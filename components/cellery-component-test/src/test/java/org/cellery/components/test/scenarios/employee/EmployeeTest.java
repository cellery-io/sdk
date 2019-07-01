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
package org.cellery.components.test.scenarios.employee;

import com.google.gson.JsonElement;
import com.google.gson.JsonParser;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.ServiceTemplate;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.models.CellImageInfo;
import org.cellery.components.test.utils.CelleryUtils;
import org.cellery.components.test.utils.LangTestUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
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
import static org.cellery.components.test.utils.CelleryTestConstants.EMPLOYEE_PORTAL;
import static org.cellery.components.test.utils.CelleryTestConstants.METADATA;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class EmployeeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH =
            SAMPLE_DIR.resolve(EMPLOYEE_PORTAL + File.separator + CELLERY + File.separator + "employee");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "employee", "1.0.0", "emp-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "employee" + BAL, cellImageInfo),
                0);
        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        cell = CelleryUtils.getInstance(CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toString());
    }

    @Test(groups = "build")
    public void validateBuildTimeCellAvailability() {
        Assert.assertNotNull(cell);
    }

    @Test(groups = "build")
    public void validateBuildTimeAPIVersion() {
        Assert.assertEquals(cell.getApiVersion(), "mesh.cellery.io/v1alpha1");
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
        final List<API> http = cell.getSpec().getGatewayTemplate().getSpec().getHttp();
        Assert.assertEquals(http.get(0).getBackend(), "employee");
        Assert.assertEquals(http.get(0).getContext(), "employee");
        Assert.assertEquals(http.get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(http.get(0).getDefinitions().get(0).getPath(), "/details");
        Assert.assertEquals(http.get(1).getBackend(), "salary");
        Assert.assertEquals(http.get(1).getContext(), "payroll");
        Assert.assertEquals(http.get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(http.get(1).getDefinitions().get(0).getPath(), "salary");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        final List<ServiceTemplate> servicesTemplates = cell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getLabels().get("team"), "HR");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0).getValue(),
                "{{instance_name}}--salary-service");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getImage(), "wso2cellery/sampleapp" +
                "-employee:0.3.0");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getServicePort(), 80);
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getName(), "salary");
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getLabels().get("owner"), "Alice");
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getLabels().get("team"), "Finance");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getServicePort(), 80);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "employee" + BAL, cellImageInfo,
                dependencyCells, tmpDir), 0);
        File newYaml = tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.getInstance(newYaml.getAbsolutePath());
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
        Assert.assertEquals(runtimeCell.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeGatewayTemplate() {
        final List<API> http = runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp();
        Assert.assertEquals(http.get(0).getBackend(), "employee");
        Assert.assertEquals(http.get(0).getContext(), "employee");
        Assert.assertEquals(http.get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(http.get(0).getDefinitions().get(0).getPath(), "/details");
        Assert.assertEquals(http.get(1).getBackend(), "salary");
        Assert.assertEquals(http.get(1).getContext(), "payroll");
        Assert.assertEquals(http.get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(http.get(1).getDefinitions().get(0).getPath(), "salary");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        final List<ServiceTemplate> servicesTemplates = runtimeCell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "employee");
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getLabels().get("team"), "HR");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0).getValue(), "emp-inst" +
                "--salary-service");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getImage(), "wso2cellery/sampleapp" +
                "-employee:0.3.0");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getServicePort(), 80);
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getName(), "salary");
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getLabels().get("owner"), "Alice");
        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getLabels().get("team"), "Finance");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getServicePort(), 80);
    }

    @Test(groups = "build")
    public void validateMetadataJSON() throws IOException {
        String metadataJsonPath =
                TARGET_PATH.toAbsolutePath().toString() + File.separator + CELLERY + File.separator + METADATA;
        try (InputStream input = new FileInputStream(metadataJsonPath)) {
            try (InputStreamReader inputStreamReader = new InputStreamReader(input)) {
                JsonElement parsedJson = new JsonParser().parse(inputStreamReader);
                Assert.assertEquals(parsedJson.getAsJsonObject().getAsJsonArray("exposed").size(), 2);
                Assert.assertFalse(parsedJson.getAsJsonObject().get("componentDep").getAsJsonObject()
                        .get("employee").isJsonNull());
            }
        }
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
