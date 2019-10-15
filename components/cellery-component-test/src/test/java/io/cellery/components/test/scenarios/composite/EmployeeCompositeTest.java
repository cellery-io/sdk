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
package io.cellery.components.test.scenarios.composite;

import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.CelleryTestConstants;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.Component;
import io.cellery.models.ComponentSpec;
import io.cellery.models.Composite;
import io.fabric8.kubernetes.api.model.Container;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
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

public class EmployeeCompositeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("composite" + File.separator + "employee");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(CelleryTestConstants.TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CelleryTestConstants.CELLERY);
    private Composite composite;
    private Composite runtimeComposite;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "employee-comp", "1.0.0", "emp-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "employee-comp" + CelleryTestConstants.BAL,
                cellImageInfo), 0);
        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + CelleryTestConstants.YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        composite = CelleryUtils.readCompositeYaml(CELLERY_PATH.resolve(cellImageInfo.getName() + CelleryTestConstants.YAML).toString());
    }

    @Test(groups = "build")
    public void validateBuildTimeCellAvailability() {
        Assert.assertNotNull(composite);
    }

    @Test(groups = "build")
    public void validateBuildTimeAPIVersion() {
        Assert.assertEquals(composite.getApiVersion(), CelleryConstants.CELLERY_API_VERSION);
    }

    @Test(groups = "build")
    public void validateBuildTimeMetaData() {
        Assert.assertEquals(composite.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        final List<Component> components = composite.getSpec().getComponents();
        Assert.assertEquals(components.get(0).getMetadata().getName(), "employee");
        Assert.assertEquals(components.get(0).getMetadata().getLabels().get("team"), "HR");
        final Container container = components.get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "{{instance_name}}--salary-service");
        Assert.assertEquals(container.getImage(), "wso2cellery/sampleapp-employee:0.3.0");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(components.get(0).getSpec().getPorts().get(0).getPort(), 80);
        Assert.assertEquals(components.get(1).getMetadata().getName(), "salary");
        Assert.assertEquals(components.get(1).getMetadata().getLabels().get("owner"), "Alice");
        Assert.assertEquals(components.get(1).getMetadata().getLabels().get("team"), "Finance");
        Assert.assertEquals(components.get(1).getSpec().getTemplate().getContainers().get(0).getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(components.get(1).getSpec().getPorts().get(0).getTargetPort(), 8080);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException  {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "employee-comp" + CelleryTestConstants.BAL,
                cellImageInfo, dependencyCells, tmpDir), 0);
        File newYaml = tempPath.resolve(CelleryTestConstants.ARTIFACTS).resolve(CelleryTestConstants.CELLERY).resolve(cellImageInfo.getName() + CelleryTestConstants.YAML).toFile();
        runtimeComposite = CelleryUtils.readCompositeYaml(newYaml.getAbsolutePath());
        Assert.assertNotNull(runtimeComposite);
    }

    @Test(groups = "run")
    public void validateRunTimeCellAvailability() {
        Assert.assertNotNull(runtimeComposite);
    }

    @Test(groups = "run")
    public void validateRunTimeAPIVersion() {
        Assert.assertEquals(runtimeComposite.getApiVersion(), CelleryTestConstants.CELLERY_MESH_VERSION);
    }

    @Test(groups = "run")
    public void validateRunTimeMetaData() {
        Assert.assertEquals(runtimeComposite.getMetadata().getName(), cellImageInfo.getInstanceName());
        Assert.assertEquals(runtimeComposite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(runtimeComposite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(runtimeComposite.getMetadata().getAnnotations().get(CelleryTestConstants.CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        final List<Component> components = runtimeComposite.getSpec().getComponents();
        Assert.assertEquals(components.get(0).getMetadata().getName(), "employee");
        Assert.assertEquals(components.get(0).getMetadata().getLabels().get("team"), "HR");
        final ComponentSpec spec = components.get(0).getSpec();
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getName(), "SALARY_HOST");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getEnv().get(0).getValue(), "emp-inst" +
                "--salary-service");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getImage(), "wso2cellery/sampleapp" +
                "-employee:0.3.0");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(spec.getPorts().get(0).getPort(), 80);
        final Component salaryComponent = components.get(1);
        Assert.assertEquals(salaryComponent.getMetadata().getName(), "salary");
        Assert.assertEquals(salaryComponent.getMetadata().getLabels().get("owner"), "Alice");
        Assert.assertEquals(salaryComponent.getMetadata().getLabels().get("team"), "Finance");
        Assert.assertEquals(salaryComponent.getSpec().getTemplate().getContainers().get(0).getPorts().get(0)
                .getContainerPort().intValue(), 8080);
        Assert.assertEquals(salaryComponent.getSpec().getPorts().get(0).getTargetPort(), 8080);
    }

    @Test(groups = "build")
    public void validateMetadataJSON() throws IOException {
        String metadataJsonPath =
                TARGET_PATH.toAbsolutePath().toString() + File.separator + CelleryTestConstants.CELLERY + File.separator + CelleryTestConstants.METADATA;
        try (InputStream input = new FileInputStream(metadataJsonPath)) {
            try (InputStreamReader inputStreamReader = new InputStreamReader(input)) {
                JsonElement parsedJson = new JsonParser().parse(inputStreamReader);
                JsonObject metadataJson = parsedJson.getAsJsonObject();
                Assert.assertEquals(metadataJson.size(), 7);
                Assert.assertEquals(metadataJson.get("org").getAsString(), "myorg");
                Assert.assertEquals(metadataJson.get("name").getAsString(), "employee-comp");
                Assert.assertEquals(metadataJson.get("ver").getAsString(), "1.0.0");
                Assert.assertFalse(metadataJson.get("zeroScalingRequired").getAsBoolean());
                Assert.assertFalse(metadataJson.get("autoScalingRequired").getAsBoolean());

                JsonObject components = metadataJson.getAsJsonObject("components");
                Assert.assertNotNull(components);
                Assert.assertEquals(components.size(), 2);
                {
                    JsonObject employeeComponent = components.getAsJsonObject("employee");
                    Assert.assertEquals(employeeComponent.get("dockerImage").getAsString(),
                            "wso2cellery/sampleapp-employee:0.3.0");
                    Assert.assertFalse(employeeComponent.get("isDockerPushRequired").getAsBoolean());
                    JsonObject labels = employeeComponent.getAsJsonObject("labels");
                    Assert.assertEquals(labels.size(), 1);
                    Assert.assertEquals(labels.get("team").getAsString(), "HR");

                    JsonObject dependencies = employeeComponent.getAsJsonObject("dependencies");
                    JsonArray componentDependencies = dependencies.getAsJsonArray("components");
                    Assert.assertEquals(componentDependencies.size(), 1);
                    Assert.assertEquals(componentDependencies.get(0).getAsString(), "salary");
                    JsonObject cellDependencies = dependencies.getAsJsonObject("cells");
                    Assert.assertEquals(cellDependencies.size(), 0);
                }
                {
                    JsonObject salaryComponent = components.getAsJsonObject("salary");
                    Assert.assertEquals(salaryComponent.get("dockerImage").getAsString(),
                            "wso2cellery/sampleapp-salary:0.3.0");
                    Assert.assertFalse(salaryComponent.get("isDockerPushRequired").getAsBoolean());

                    JsonObject labels = salaryComponent.getAsJsonObject("labels");
                    Assert.assertEquals(labels.size(), 2);
                    Assert.assertEquals(labels.get("owner").getAsString(), "Alice");
                    Assert.assertEquals(labels.get("team").getAsString(), "Finance");

                    JsonObject dependencies = salaryComponent.getAsJsonObject("dependencies");
                    JsonArray componentDependencies = dependencies.getAsJsonArray("components");
                    Assert.assertEquals(componentDependencies.size(), 0);
                    JsonObject cellDependencies = dependencies.getAsJsonObject("cells");
                    Assert.assertEquals(cellDependencies.size(), 0);
                }
            }
        }
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
