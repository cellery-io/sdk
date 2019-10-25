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

package io.cellery.components.test.scenarios.petstore;

import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.Cell;
import io.cellery.models.Ingress;
import io.cellery.models.OIDC;
import io.fabric8.kubernetes.api.model.Container;
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
import static io.cellery.components.test.utils.CelleryTestConstants.PET_CARE_STORE;
import static io.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static io.cellery.components.test.utils.CelleryTestConstants.YAML;
import static io.cellery.components.test.utils.LangTestUtils.deleteDirectory;

public class PetStoreFeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve(PET_CARE_STORE + File.separator + "pet" +
            "-fe");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "petfe", "1.0.0", "petfe-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "pet-fe" + BAL,
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
    public void validateBuildTimeGatewayTemplate() {
        final Ingress ingress = cell.getSpec().getGateway().getSpec().getIngress();
        Assert.assertEquals(ingress.getExtensions().getClusterIngress().getHost(), "pet-store.com");
        Assert.assertEquals(ingress.getHttp().get(0).getDestination().getHost(), "portal");
        Assert.assertEquals(ingress.getHttp().get(0).getContext(), "/");
        Assert.assertFalse(ingress.getHttp().get(0).isGlobal());
        final OIDC oidc = ingress.getExtensions().getOidc();
        Assert.assertEquals(oidc.getBaseUrl(), "http://pet-store.com/");
        Assert.assertEquals(oidc.getClientId(), "petstoreapplication");
        Assert.assertEquals(oidc.getDcrPassword(), "admin");
        Assert.assertEquals(oidc.getDcrUser(), "admin");
        Assert.assertEquals(oidc.getNonSecurePaths().iterator().next(), "/app/*");
        Assert.assertEquals(oidc.getProviderUrl(), "https://idp.cellery-system/oauth2/token");
        Assert.assertEquals(oidc.getRedirectUrl(), "http://pet-store.com/_auth/callback");
        Assert.assertEquals(oidc.getSubjectClaim(), "given_name");
    }

    @Test(groups = "build")
    public void validateBuildTimeServicesTemplates() {
        Assert.assertEquals(cell.getSpec().getComponents().get(0).getMetadata().getName(), "portal");
        final Container container =
                cell.getSpec().getComponents().get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getImage(), "wso2cellery/samples-pet-store-portal");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 80);

        Assert.assertEquals(container.getEnv().get(0).getName(), "PORTAL_PORT");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "80");
        Assert.assertEquals(container.getEnv().get(1).getName(), "BASE_PATH");
        Assert.assertEquals(container.getEnv().get(1).getValue(), ".");
        Assert.assertEquals(container.getEnv().get(2).getName(), "PET_STORE_CELL_URL");
        Assert.assertEquals(container.getEnv().get(2).getValue(), "" +
                "http://{{petstorebackend}}--gateway-service:80/controller");
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        CellImageInfo petbeDep = new CellImageInfo("myorg", "petbe", "1.0.0", "petbe-inst");
        dependencyCells.put("petstorebackend", petbeDep);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "pet-fe" + BAL,
                cellImageInfo, dependencyCells, tmpDir), 0);
        File newYaml =
                tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.readCellYaml(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateMetadata() throws IOException {
        Map<String, CellImageInfo> dependencyInfo = LangTestUtils.getDependencyInfo(SOURCE_DIR_PATH);
        CellImageInfo petbeImage = dependencyInfo.get("petstorebackend");
        Assert.assertEquals(petbeImage.getOrg(), "myorg");
        Assert.assertEquals(petbeImage.getName(), "petbe");
        Assert.assertEquals(petbeImage.getVer(), "1.0.0");
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
    public void validateRunTimeGatewayTemplate() {
        final Ingress ingress = runtimeCell.getSpec().getGateway().getSpec().getIngress();
        Assert.assertEquals(ingress.getExtensions().getClusterIngress().getHost(), "pet-store.com");
        Assert.assertEquals(ingress.getHttp().get(0).getDestination().getHost(), "portal");
        Assert.assertEquals(ingress.getHttp().get(0).getContext(), "/");
        Assert.assertFalse(ingress.getHttp().get(0).isGlobal());
        final OIDC oidc = ingress.getExtensions().getOidc();
        Assert.assertEquals(oidc.getBaseUrl(), "http://pet-store.com/");
        Assert.assertEquals(oidc.getClientId(), "petstoreapplication");
        Assert.assertEquals(oidc.getDcrPassword(), "admin");
        Assert.assertEquals(oidc.getDcrUser(), "admin");
        Assert.assertEquals(oidc.getNonSecurePaths().iterator().next(), "/app/*");
        Assert.assertEquals(oidc.getProviderUrl(), "https://idp.cellery-system/oauth2/token");
        Assert.assertEquals(oidc.getRedirectUrl(), "http://pet-store.com/_auth/callback");
        Assert.assertEquals(oidc.getSubjectClaim(), "given_name");
    }

    @Test(groups = "run")
    public void validateRunTimeServicesTemplates() {
        Assert.assertEquals(runtimeCell.getSpec().getComponents().get(0).getMetadata().getName(),
                "portal");
        final Container container =
                runtimeCell.getSpec().getComponents().get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getImage(), "wso2cellery/samples-pet-store-portal");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 80);

        Assert.assertEquals(container.getEnv().get(0).getName(), "PORTAL_PORT");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "80");
        Assert.assertEquals(container.getEnv().get(1).getName(), "BASE_PATH");
        Assert.assertEquals(container.getEnv().get(1).getValue(), ".");
        Assert.assertEquals(container.getEnv().get(2).getName(), "PET_STORE_CELL_URL");
        Assert.assertEquals(container.getEnv().get(2).getValue(), "http://petbe-inst--gateway-service:80/controller");
    }

    @AfterClass
    public void cleanUp() throws IOException {
        deleteDirectory(TARGET_PATH);
    }
}
