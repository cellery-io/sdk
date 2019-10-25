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

package io.cellery.components.test.scenarios.volume;

import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.Cell;
import io.cellery.models.Component;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaimSpec;
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
import static io.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static io.cellery.components.test.utils.CelleryTestConstants.YAML;

public class VolumeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("volumes");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "volume", "1.0.0", "volume-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "volume" + BAL,
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
    public void validateBuildTimeKind() {
        Assert.assertEquals(cell.getKind(), "Cell");
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
    public void validateBuildTimeServiceTemplates() {
        final Component component = cell.getSpec().getComponents().get(0);
        Assert.assertEquals(cell.getSpec().getGateway().getSpec().getIngress().getHttp().get(0).getVersion(), "v1.0.0");
        Assert.assertEquals(component.getMetadata().getName(), "hello-api");
        final Container container = component.getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getImage(),
                "docker.io/wso2cellery/samples-hello-world-api-hello-service");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 9090);
        Assert.assertEquals(container.getVolumeMounts().size(), 6);
        Assert.assertEquals(container.getVolumeMounts().get(0).getName(), "{{instance_name}}-my-secret");
        Assert.assertEquals(container.getVolumeMounts().get(0).getMountPath(), "/tmp/secret/");
        Assert.assertFalse(container.getVolumeMounts().get(0).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(1).getName(), "my-secret-shared");
        Assert.assertEquals(container.getVolumeMounts().get(1).getMountPath(), "/tmp/shared/secret");
        Assert.assertFalse(container.getVolumeMounts().get(1).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(2).getName(), "my-config");
        Assert.assertEquals(container.getVolumeMounts().get(2).getMountPath(), "/tmp/config");
        Assert.assertFalse(container.getVolumeMounts().get(2).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(3).getName(), "my-config-shared");
        Assert.assertEquals(container.getVolumeMounts().get(3).getMountPath(), "/tmp/shared/config");
        Assert.assertFalse(container.getVolumeMounts().get(3).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(4).getName(), "pv1");
        Assert.assertEquals(container.getVolumeMounts().get(4).getMountPath(), "/tmp/pvc/");
        Assert.assertFalse(container.getVolumeMounts().get(4).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(5).getName(), "pv2");
        Assert.assertEquals(container.getVolumeMounts().get(5).getMountPath(), "/tmp/pvc/shared");
        Assert.assertTrue(container.getVolumeMounts().get(5).getReadOnly());
        Assert.assertEquals(component.getSpec().getTemplate().getVolumes().size(), 3);
        Assert.assertEquals(component.getSpec().getVolumeClaims().size(), 1);
        Assert.assertEquals(component.getSpec().getConfigurations().size(), 1);
        Assert.assertEquals(component.getSpec().getSecrets().size(), 1);

        Assert.assertEquals(component.getSpec().getVolumeClaims().get(0).getName(), "pv1");
        Assert.assertEquals(component.getSpec().getConfigurations().get(0).getMetadata().getName(), "my-config");
        Assert.assertEquals(component.getSpec().getSecrets().get(0).getMetadata().getName(), "{{instance_name}}-my" +
                "-secret");

    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "volume" + BAL, cellImageInfo,
                dependencyCells, tmpDir), 0);
        File newYaml =
                tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
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
    public void validateRunTimeKind() {
        Assert.assertEquals(runtimeCell.getKind(), "Cell");
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
    public void validateRunTimeServiceTemplates() {
        Assert.assertEquals(cell.getSpec().getGateway().getSpec().getIngress().getHttp().get(0).getVersion(), "v1.0.0");
        final Component component = runtimeCell.getSpec().getComponents().get(0);
        Assert.assertEquals(component.getMetadata().getName(), "hello-api");
        final Container container = component.getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getImage(),
                "docker.io/wso2cellery/samples-hello-world-api-hello-service");
        Assert.assertEquals(container.getPorts()
                .get(0).getContainerPort().intValue(), 9090);
        Assert.assertEquals(container.getVolumeMounts().size(), 6);
        Assert.assertEquals(container.getVolumeMounts().get(0).getName(), "volume-inst-my-secret");
        Assert.assertEquals(container.getVolumeMounts().get(0).getMountPath(), "/tmp/secret/");
        Assert.assertFalse(container.getVolumeMounts().get(0).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(1).getName(), "my-secret-shared");
        Assert.assertEquals(container.getVolumeMounts().get(1).getMountPath(), "/tmp/shared/secret");
        Assert.assertFalse(container.getVolumeMounts().get(1).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(2).getName(), "my-config");
        Assert.assertEquals(container.getVolumeMounts().get(2).getMountPath(), "/tmp/config");
        Assert.assertFalse(container.getVolumeMounts().get(2).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(3).getName(), "my-config-shared");
        Assert.assertEquals(container.getVolumeMounts().get(3).getMountPath(), "/tmp/shared/config");
        Assert.assertFalse(container.getVolumeMounts().get(3).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(4).getName(), "pv1");
        Assert.assertEquals(container.getVolumeMounts().get(4).getMountPath(), "/tmp/pvc/");
        Assert.assertFalse(container.getVolumeMounts().get(4).getReadOnly());
        Assert.assertEquals(container.getVolumeMounts().get(5).getName(), "pv2");
        Assert.assertEquals(container.getVolumeMounts().get(5).getMountPath(), "/tmp/pvc/shared");
        Assert.assertTrue(container.getVolumeMounts().get(5).getReadOnly());
        Assert.assertEquals(component.getSpec().getVolumeClaims().get(0).getName(), "pv1");
        final PersistentVolumeClaimSpec pvSpec =
                component.getSpec().getVolumeClaims().get(0).getTemplate().getSpec();
        Assert.assertEquals(pvSpec.getStorageClassName(), "slow");
        Assert.assertEquals(pvSpec.getSelector().getMatchLabels().size(), 1);
        Assert.assertEquals(pvSpec.getSelector().getMatchExpressions().size(), 1);

        Assert.assertEquals(component.getSpec().getConfigurations().get(0).getMetadata().getName(), "my-config");
        Assert.assertEquals(component.getSpec().getSecrets().get(0).getMetadata().getName(), "volume-inst-my" +
                "-secret");

    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
