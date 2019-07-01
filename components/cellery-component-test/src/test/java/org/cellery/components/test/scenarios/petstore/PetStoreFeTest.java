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

package org.cellery.components.test.scenarios.petstore;

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
import static org.cellery.components.test.utils.CelleryTestConstants.PET_CARE_STORE;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

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
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHost(), "pet-store.com");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getBackend(),
                "portal");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getContext(),
                "/");
        Assert.assertTrue(cell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).isGlobal());
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getBaseUrl(), "http" +
                "://pet-store.com/");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getClientId(),
                "petstoreapplication");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getDcrPassword(),
                "admin");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getDcrUser(), "admin");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().
                getNonSecurePaths().iterator().next(), "/app/*");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getProviderUrl(),
                "https://idp.cellery-system/oauth2/token");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getRedirectUrl(), "http" +
                "://pet-store.com/_auth/callback");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getOidc().getSubjectClaim(),
                "given_name");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getType(), "Envoy");
    }

    @Test(groups = "build")
    public void validateBuildTimeServicesTemplates() {
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getMetadata().getName(), "portal");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-portal");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 80);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getServicePort(), 80);

        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(0)
                .getName(), "PORTAL_PORT");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(0).
                getValue(), "80");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(1).
                getName(), "BASE_PATH");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(1).
                getValue(), ".");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(2).
                getName(), "PET_STORE_CELL_URL");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(2).
                getValue(), "http://{{petstorebackend}}--gateway-service:80/controller");
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
        runtimeCell = CelleryUtils.getInstance(newYaml.getAbsolutePath());
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
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHost(), "pet-store.com");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getBackend(),
                "portal");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).getContext(),
                "/");
        Assert.assertTrue(runtimeCell.getSpec().getGatewayTemplate().getSpec().getHttp().get(0).isGlobal());
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getBaseUrl(),
                "http://pet-store" +
                        ".com/");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getClientId(),
                "petstoreapplication");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getDcrPassword()
                , "admin");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getDcrUser(),
                "admin");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().
                getNonSecurePaths().iterator().next(), "/app/*");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getProviderUrl()
                , "https://idp" +
                        ".cellery-system/oauth2/token");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getRedirectUrl()
                , "http://pet" +
                        "-store.com/_auth/callback");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getOidc().getSubjectClaim(),
                "given_name");
        Assert.assertEquals(runtimeCell.getSpec().getGatewayTemplate().getSpec().getType(), "Envoy");
    }

    @Test(groups = "run")
    public void validateRunTimeServicesTemplates() {
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getMetadata().getName(),
                "portal");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-portal");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getPorts()
                .get(0).getContainerPort().intValue(), 80);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getServicePort(),
                80);

        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().
                get(0).getName(), "PORTAL_PORT");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv()
                .get(0).getValue(), "80");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv()
                .get(1).getName(), "BASE_PATH");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv()
                .get(1).getValue(), ".");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv()
                .get(2).getName(), "PET_STORE_CELL_URL");
        Assert.assertEquals(runtimeCell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv()
                .get(2).getValue(), "http://petbe-inst--gateway-service:80/controller");
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
