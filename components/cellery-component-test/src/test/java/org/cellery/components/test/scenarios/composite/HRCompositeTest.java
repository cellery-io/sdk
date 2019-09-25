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

package org.cellery.components.test.scenarios.composite;

import io.cellery.CelleryUtils;
import io.cellery.models.ComponentSpec;
import io.cellery.models.Composite;
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

public class HRCompositeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("composite" + File.separator + "hr");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Composite composite;
    private Composite runtimeComposite;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "hr-comp", "1.0.0", "hr-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "hr-comp" + BAL,
                cellImageInfo), 0);
        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        composite = CelleryUtils.readCompositeYaml(CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toString());
    }

    @Test(groups = "build")
    public void validateBuildTimeCellAvailability() {
        Assert.assertNotNull(composite);
    }

    @Test(groups = "build")
    public void validateBuildTimeAPIVersion() {
        Assert.assertEquals(composite.getApiVersion(), CELLERY_MESH_VERSION);
    }

    @Test(groups = "build")
    public void validateBuildTimeMetaData() {
        Assert.assertEquals(composite.getMetadata().getName(), cellImageInfo.getName());
        final Map<String, String> annotations = composite.getMetadata().getAnnotations();
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
        Assert.assertEquals(annotations.get("mesh.cellery.io/cell-dependencies").length(), 199);
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(composite.getSpec().getComponents().get(0).getMetadata().getName(), "hr");
        final ComponentSpec spec = composite.getSpec().getComponents().get(0).getSpec();
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getName(), "stock_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(1).getName(), "employee_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getImage(), "wso2cellery/sampleapp-hr:0.3.0");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getPorts().get(0).getContainerPort().intValue()
                , 8080);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        CellImageInfo employeeDep = new CellImageInfo("myorg", "employee-comp", "1.0.0", "emp-inst");
        dependencyCells.put("employeeCompDep", employeeDep);
        CellImageInfo stockDep = new CellImageInfo("myorg", "stock-comp", "1.0.0", "stock-inst");
        dependencyCells.put("stockCompDep", stockDep);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "hr-comp" + BAL, cellImageInfo,
                dependencyCells, tmpDir), 0);
        File newYaml = tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeComposite = CelleryUtils.readCompositeYaml(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateMetadata() throws IOException {
        Map<String, CellImageInfo> dependencyInfo = LangTestUtils.getDependencyInfo(SOURCE_DIR_PATH);
        CellImageInfo employeeImage = dependencyInfo.get("employeeCompDep");
        Assert.assertEquals(employeeImage.getOrg(), "myorg");
        Assert.assertEquals(employeeImage.getName(), "employee-comp");
        Assert.assertEquals(employeeImage.getVer(), "1.0.0");

        CellImageInfo stockImage = dependencyInfo.get("stockCompDep");
        Assert.assertEquals(stockImage.getOrg(), "myorg");
        Assert.assertEquals(stockImage.getName(), "stock-comp");
        Assert.assertEquals(stockImage.getVer(), "1.0.0");
    }

    @Test(groups = "run")
    public void validateRunTimeCellAvailability() {
        Assert.assertNotNull(runtimeComposite);
    }

    @Test(groups = "run")
    public void validateRunTimeAPIVersion() {
        Assert.assertEquals(runtimeComposite.getApiVersion(), CELLERY_MESH_VERSION);
    }

    @Test(groups = "run")
    public void validateRunTimeMetaData() {
        Assert.assertEquals(runtimeComposite.getMetadata().getName(), cellImageInfo.getInstanceName());
        final Map<String, String> annotations = runtimeComposite.getMetadata().getAnnotations();
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(annotations.get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get("mesh.cellery.io/cell-dependencies").length(),
                199);
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        Assert.assertEquals(runtimeComposite.getSpec().getComponents().get(0).getMetadata().getName(), "hr");
        final ComponentSpec spec = runtimeComposite.getSpec().getComponents().get(0).getSpec();
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getName(), "stock_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getValue(), "http://stock-inst" +
                "--stock-service:80");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(1).getName(), "employee_api_url");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(1).getValue(), "http://emp-inst" +
                "--employee-service:80");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getImage(), "wso2cellery/sampleapp-hr:0.3.0");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getPorts().get(0).getContainerPort().intValue()
                , 8080);
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}

