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
import io.cellery.models.GatewaySpec;
import io.cellery.models.ServiceTemplate;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.models.CellImageInfo;
import org.cellery.components.test.utils.CelleryUtils;
import org.cellery.components.test.utils.LangTestUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;

import static org.cellery.components.test.utils.CelleryTestConstants.BAL;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static org.cellery.components.test.utils.CelleryTestConstants.PET_CARE_STORE;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class PetStoreBeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve(PET_CARE_STORE);
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "petbe", "1.0.0");

    @BeforeClass
    public void compileSample() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "pet-be" + BAL, cellImageInfo), 0);
        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        cell = CelleryUtils.getInstance(CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toString());
    }

    @Test
    public void validateCellAvailability() {
        Assert.assertNotNull(cell);
    }

    @Test
    public void validateAPIVersion() {
        Assert.assertEquals(cell.getApiVersion(), "mesh.cellery.io/v1alpha1");
    }

    @Test
    public void validateKind() {
        Assert.assertEquals(cell.getKind(), "Cell");
    }

    @Test
    public void validateMetaData() {
        Assert.assertEquals(cell.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test
    public void validateGatewayTemplate() {
        GatewaySpec gatewaySpec = cell.getSpec().getGatewayTemplate().getSpec();
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getBackend(), "controller");
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getContext(), "controller");
        Assert.assertTrue(gatewaySpec.getHttp().get(0).isGlobal());
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getDefinitions().get(0).getPath(), "/catalog");
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getDefinitions().get(1).getMethod(), "GET");
        Assert.assertEquals(gatewaySpec.getHttp().get(0).getDefinitions().get(1).getPath(), "/orders");
        Assert.assertEquals(gatewaySpec.getType(), "MicroGateway");
    }

    @Test
    public void validateServicesTemplates() {
        List<ServiceTemplate> servicesTemplates = cell.getSpec().getServicesTemplates();
        Assert.assertEquals(servicesTemplates.get(0).getMetadata().getName(), "controller");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-controller");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 80);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getServicePort(), 80);

        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0)
                .getName(), "CATALOG_PORT");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(0).
                getValue(), "80");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(1).
                getName(), "ORDER_PORT");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(1).
                getValue(), "80");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(2).
                getName(), "ORDER_HOST");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(2).
                getValue(), "{{instance_name}}--orders-service");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(3).
                getName(), "CUSTOMER_PORT");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(3).
                getValue(), "80");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(4).
                getName(), "CATALOG_HOST");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(4).
                getValue(), "{{instance_name}}--catalog-service");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(5).
                getName(), "CUSTOMER_HOST");
        Assert.assertEquals(servicesTemplates.get(0).getSpec().getContainer().getEnv().get(5).
                getValue(), "{{instance_name}}--customers-service");


        Assert.assertEquals(servicesTemplates.get(1).getMetadata().getName(), "catalog");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-catalog");
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 80);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(1).getSpec().getServicePort(), 80);

        Assert.assertEquals(servicesTemplates.get(2).getMetadata().getName(), "orders");
        Assert.assertEquals(servicesTemplates.get(2).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-orders");
        Assert.assertEquals(servicesTemplates.get(2).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 80);
        Assert.assertEquals(servicesTemplates.get(2).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(2).getSpec().getServicePort(), 80);

        Assert.assertEquals(servicesTemplates.get(3).getMetadata().getName(), "customers");
        Assert.assertEquals(servicesTemplates.get(3).getSpec().getContainer().getImage(),
                "wso2cellery/samples-pet-store-customers");
        Assert.assertEquals(servicesTemplates.get(3).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 80);
        Assert.assertEquals(servicesTemplates.get(3).getSpec().getReplicas(), 1);
        Assert.assertEquals(servicesTemplates.get(3).getSpec().getServicePort(), 80);

    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(String.valueOf(TARGET_PATH));
    }
}
