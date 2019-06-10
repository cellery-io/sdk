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

package org.cellery.components.test.scenarios.dockersource;

import io.cellery.models.Cell;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.models.CellImageInfo;
import org.cellery.components.test.utils.CelleryUtils;
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
import static org.cellery.components.test.utils.CelleryTestConstants.DOCKER_SOURCE;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class DockerSourceTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve(DOCKER_SOURCE);
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "docker-source", "1.0.0", "docker-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "docker-source" + BAL,
                cellImageInfo), 0);
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
    public void validateBuildTimeGatewayTemplate() {
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getBackend(),
                "hello");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getContext(),
                "/hello");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getDefinitions().get(0)
                .getPath(), "/sayHello");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getDefinitions().get(0)
                        .getMethod(),
                "GET");
        Assert.assertTrue(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).isGlobal());

        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getBackend(),
                "hellox");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getContext(),
                "/hellox");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getDefinitions().get(0)
                .getPath(), "/sayHellox");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getDefinitions().get(0)
                .getMethod(), "GET");
        Assert.assertTrue(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).isGlobal());
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getMetadata().getName(), "hello");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getImage(),
                cellImageInfo.getOrg() + "/sampleapp-hello");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 9090);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getServicePort(), 80);

        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getMetadata().getName(), "hellox");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getSpec().getContainer().getImage(),
                cellImageInfo.getOrg() + "/sampleapp-hellox");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 9090);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getSpec().getServicePort(), 80);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "docker-source" + BAL,
                cellImageInfo, dependencyCells, tmpDir), 0);
        File newYaml =
                tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.getInstance(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateRunTimeCellAvailability() {
        Assert.assertNotNull(runtimeCell);
    }

    @Test(groups = "run")
    public void validateRunTimeAPIVersion() {
        Assert.assertEquals(runtimeCell.getApiVersion(), "mesh.cellery.io/v1alpha1");
    }

    @Test(groups = "run")
    public void validateRunTimeKind() {
        Assert.assertEquals(runtimeCell.getKind(), "Cell");
    }

    @Test(groups = "run")
    public void validateRunTimeMetaData() {
        Assert.assertEquals(runtimeCell.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeGatewayTemplate() {
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getBackend()
                , "hello");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getContext()
                , "/hello");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getDefinitions()
                .get(0)
                .getPath(), "/sayHello");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getDefinitions()
                        .get(0)
                        .getMethod(),
                "GET");
        Assert.assertTrue(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).isGlobal());

        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getBackend()
                , "hellox");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getContext()
                , "/hellox");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getDefinitions()
                .get(0)
                .getPath(), "/sayHellox");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).getDefinitions()
                .get(0)
                .getMethod(), "GET");
        Assert.assertTrue(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(1).isGlobal());
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getType(), "MicroGateway");
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getMetadata().getName(),
                "hello");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getImage(),
                cellImageInfo.getOrg() + "/sampleapp-hello");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getPorts()
                .get(0).
                        getContainerPort().intValue(), 9090);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getServicePort(),
                80);

        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(1).getMetadata().getName(),
                "hellox");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(1).getSpec().getContainer().getImage(),
                cellImageInfo.getOrg() + "/sampleapp-hellox");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(1).getSpec().getContainer().getPorts()
                .get(0).
                        getContainerPort().intValue(), 9090);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(1).getSpec().getServicePort(),
                80);
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(String.valueOf(TARGET_PATH));
    }
}
