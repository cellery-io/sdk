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

package io.cellery.components.test.scenarios.employee;

import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.ComponentSpec;
import io.fabric8.kubernetes.api.model.Container;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;

import static io.cellery.components.test.utils.CelleryTestConstants.ARTIFACTS;
import static io.cellery.components.test.utils.CelleryTestConstants.BAL;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_MESH_VERSION;
import static io.cellery.components.test.utils.CelleryTestConstants.EMPLOYEE_PORTAL;
import static io.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static io.cellery.components.test.utils.CelleryTestConstants.YAML;

public class HRTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH =
            SAMPLE_DIR.resolve(EMPLOYEE_PORTAL + File.separator + CELLERY + File.separator + "hr");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "hr-img", "1.0.0", "hr-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "hr" + BAL,
                cellImageInfo), 0);
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
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "build")
    public void validateBuildTimeGatewayTemplate() {
        final API api = cell.getSpec().getGateway().getSpec().getIngress().getHttp().get(0);
        Assert.assertEquals(api.getDestination().getHost(), "hr");
        Assert.assertEquals(api.getContext(), "/hr");
        Assert.assertEquals(api.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api.getDefinitions().get(0).getPath(), "/");
        Assert.assertTrue(api.isGlobal());
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(cell.getSpec().getComponents().get(0).getMetadata().getName(), "hr");
        final ComponentSpec spec = cell.getSpec().getComponents().get(0).getSpec();
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getName(), "stock_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(1).
                getName(), "employee_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getImage(),
                "wso2cellery/sampleapp-hr:0.3.0");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getPorts().get(0).
                getContainerPort().intValue(), 8080);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        CellImageInfo employeeDep = new CellImageInfo("myorg", "employee", "1.0.0", "emp-inst");
        dependencyCells.put("employeeCellDep", employeeDep);
        CellImageInfo stockDep = new CellImageInfo("myorg", "stock", "1.0.0", "stock-inst");
        dependencyCells.put("stockCellDep", stockDep);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "hr" + BAL, cellImageInfo,
                dependencyCells, tmpDir), 0);
        File newYaml = tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.readCellYaml(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateMetadata() throws IOException {
        Map<String, CellImageInfo> dependencyInfo = LangTestUtils.getDependencyInfo(SOURCE_DIR_PATH);
        CellImageInfo employeeImage = dependencyInfo.get("employeeCellDep");
        Assert.assertEquals(employeeImage.getOrg(), "myorg");
        Assert.assertEquals(employeeImage.getName(), "employee");
        Assert.assertEquals(employeeImage.getVer(), "1.0.0");

        CellImageInfo stockImage = dependencyInfo.get("stockCellDep");
        Assert.assertEquals(stockImage.getOrg(), "myorg");
        Assert.assertEquals(stockImage.getName(), "stock");
        Assert.assertEquals(stockImage.getVer(), "1.0.0");
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
        Assert.assertEquals(runtimeCell.getMetadata().getName(), cellImageInfo.getInstanceName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeGatewayTemplate() {
        final API api = runtimeCell.getSpec().getGateway().getSpec().getIngress().getHttp().get(0);
        Assert.assertEquals(api.getDestination().getHost(), "hr");
        Assert.assertEquals(api.getContext(), "/hr");
        Assert.assertEquals(api.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api.getDefinitions().get(0).getPath(), "/");
        Assert.assertTrue(api.isGlobal());
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        Assert.assertEquals(runtimeCell.getSpec().getComponents().get(0).getMetadata().getName(),
                "hr");
        final Container container =
                runtimeCell.getSpec().getComponents().get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getEnv().get(0).getName(), "stock_api_url");
        Assert.assertEquals(container.getEnv().get(1).getName(), "employee_api_url");
        Assert.assertEquals(container.getImage(), "wso2cellery/sampleapp-hr:0.3.0");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);

    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}

