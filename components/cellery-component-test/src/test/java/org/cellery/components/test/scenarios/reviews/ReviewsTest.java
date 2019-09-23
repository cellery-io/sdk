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

package org.cellery.components.test.scenarios.reviews;

import io.cellery.CelleryUtils;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.Component;
import io.cellery.models.GatewaySpec;
import io.fabric8.kubernetes.api.model.Container;
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
import static org.cellery.components.test.utils.CelleryTestConstants.PRODUCT_REVIEW;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class ReviewsTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH =
            SAMPLE_DIR.resolve(PRODUCT_REVIEW + File.separator + CELLERY + File.separator +
                    "reviews");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "reviews", "1.0.0", "review-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "reviews" + BAL,
                cellImageInfo)
                , 0);
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
        GatewaySpec cellGatewaySpec = cell.getSpec().getGateway().getSpec();
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(0).getDestination().getHost(), "reviews");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(0).getContext(), "reviews-1");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(0).getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(cellGatewaySpec.getIngress().getHttp().get(0).isAuthenticate());
        Assert.assertTrue(cellGatewaySpec.getIngress().getHttp().get(0).isGlobal());

        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(1).getDestination().getHost(), "ratings");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(1).getContext(), "ratings-1");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(cellGatewaySpec.getIngress().getHttp().get(1).getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(cellGatewaySpec.getIngress().getHttp().get(0).isAuthenticate());

    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        List<Component> components = cell.getSpec().getComponents();
        Assert.assertEquals(components.get(0).getMetadata().getName(), "reviews");

        final Container container = components.get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getEnv().get(0).getName(), "PRODUCTS_HOST");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "{{customerProduct}}--gateway-service");

        Assert.assertEquals(container.getEnv().get(1).getName(), "DATABASE_NAME");
        Assert.assertEquals(container.getEnv().get(1).getValue(), "reviews_db");

        Assert.assertEquals(container.getEnv().get(2).getName(), "PORT");
        Assert.assertEquals(container.getEnv().get(2).getValue(), "8080");

        Assert.assertEquals(container.getEnv().get(3).getName(), "DATABASE_HOST");
        Assert.assertEquals(container.getEnv().get(3).getValue(), "{{database}}--gateway-service");

        Assert.assertEquals(container.getEnv().get(4).getName(), "RATINGS_PORT");
        Assert.assertEquals(container.getEnv().get(4).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(5).getName(), "DATABASE_PORT");
        Assert.assertEquals(container.getEnv().get(5).getValue(), "31406");

        Assert.assertEquals(container.getEnv().get(6).getName(), "CUSTOMERS_PORT");
        Assert.assertEquals(container.getEnv().get(6).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(7).getName(), "CUSTOMERS_HOST");
        Assert.assertEquals(container.getEnv().get(7).getValue(), "{{customerProduct}}--gateway-service");

        Assert.assertEquals(container.getEnv().get(8).getName(), "CUSTOMERS_CONTEXT");
        Assert.assertEquals(container.getEnv().get(8).getValue(), "customers-1");

        Assert.assertEquals(container.getEnv().get(9).getName(), "PRODUCTS_CONTEXT");
        Assert.assertEquals(container.getEnv().get(9).getValue(), "products-1");

        Assert.assertEquals(container.getEnv().get(10).getName(), "DATABASE_USERNAME");
        Assert.assertEquals(container.getEnv().get(10).getValue(), "root");

        Assert.assertEquals(container.getEnv().get(11).getName(), "RATINGS_HOST");
        Assert.assertEquals(container.getEnv().get(11).getValue(), "{{instance_name}}--ratings-service");

        Assert.assertEquals(container.getEnv().get(12).getName(), "PRODUCTS_PORT");
        Assert.assertEquals(container.getEnv().get(12).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(13).getName(), "DATABASE_PASSWORD");
        Assert.assertEquals(container.getEnv().get(13).getValue(), "root");

        Assert.assertEquals(container.getImage(), "celleryio/samples-productreview-reviews");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);

        Assert.assertEquals(components.get(1).getMetadata().getName(), "ratings");
        Assert.assertEquals(components.get(1).getSpec().getTemplate().getContainers().get(0).getPorts().get(0).
                getContainerPort().intValue(), 8080);
        Assert.assertEquals(components.get(1).getSpec().getTemplate().getContainers().get(0).getEnv().get(0).getName(),
                "PORT");
        Assert.assertEquals(components.get(1).getSpec().getTemplate().getContainers().get(0).getEnv().get(0).getValue(),
                "8080");
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        CellImageInfo databaseDep = new CellImageInfo("myorg", "database", "1.0.0", "db-inst");
        dependencyCells.put("database", databaseDep);
        CellImageInfo customerProductDep = new CellImageInfo("myorg", "products", "1.0.0", "cust-inst");
        dependencyCells.put("customerProduct", customerProductDep);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "reviews" + BAL,
                cellImageInfo, dependencyCells, tmpDir), 0);
        File newYaml =
                tempPath.resolve(ARTIFACTS).resolve(CELLERY).resolve(cellImageInfo.getName() + YAML).toFile();
        runtimeCell = CelleryUtils.readCellYaml(newYaml.getAbsolutePath());
    }

    @Test(groups = "run")
    public void validateMetadata() throws IOException {
        Map<String, CellImageInfo> dependencyInfo = LangTestUtils.getDependencyInfo(SOURCE_DIR_PATH);
        CellImageInfo databaseImage = dependencyInfo.get("database");
        Assert.assertEquals(databaseImage.getOrg(), "myorg");
        Assert.assertEquals(databaseImage.getName(), "database");
        Assert.assertEquals(databaseImage.getVer(), "1.0.0");

        CellImageInfo customerImage = dependencyInfo.get("customerProduct");
        Assert.assertEquals(customerImage.getOrg(), "myorg");
        Assert.assertEquals(customerImage.getName(), "products");
        Assert.assertEquals(customerImage.getVer(), "1.0.0");
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
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                cellImageInfo.getOrg());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                cellImageInfo.getName());
        Assert.assertEquals(runtimeCell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                cellImageInfo.getVer());
    }

    @Test(groups = "run")
    public void validateRunTimeGatewayTemplate() {
        GatewaySpec runtimeGatewaySpec = runtimeCell.getSpec().getGateway().getSpec();
        final API api = runtimeGatewaySpec.getIngress().getHttp().get(0);
        Assert.assertEquals(api.getDestination().getHost(), "reviews");
        Assert.assertEquals(api.getContext(), "reviews-1");
        Assert.assertEquals(api.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api.getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(api.isAuthenticate());
        Assert.assertTrue(api.isGlobal());

        final API api1 = runtimeGatewaySpec.getIngress().getHttp().get(1);
        Assert.assertEquals(api1.getDestination().getHost(), "ratings");
        Assert.assertEquals(api1.getContext(), "ratings-1");
        Assert.assertEquals(api1.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api1.getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(api.isAuthenticate());
    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        List<Component> components = runtimeCell.getSpec().getComponents();
        Assert.assertEquals(components.get(0).getMetadata().getName(), "reviews");

        final Container container = components.get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getEnv().get(0).getName(), "PRODUCTS_HOST");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "cust-inst--gateway-service");

        Assert.assertEquals(container.getEnv().get(1).getName(), "DATABASE_NAME");
        Assert.assertEquals(container.getEnv().get(1).getValue(), "reviews_db");

        Assert.assertEquals(container.getEnv().get(2).getName(), "PORT");
        Assert.assertEquals(container.getEnv().get(2).getValue(), "8080");

        Assert.assertEquals(container.getEnv().get(3).getName(), "DATABASE_HOST");
        Assert.assertEquals(container.getEnv().get(3).getValue(), "db-inst--gateway-service");

        Assert.assertEquals(container.getEnv().get(4).getName(), "RATINGS_PORT");
        Assert.assertEquals(container.getEnv().get(4).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(5).getName(), "DATABASE_PORT");
        Assert.assertEquals(container.getEnv().get(5).getValue(), "31406");

        Assert.assertEquals(container.getEnv().get(6).getName(), "CUSTOMERS_PORT");
        Assert.assertEquals(container.getEnv().get(6).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(7).getName(), "CUSTOMERS_HOST");
        Assert.assertEquals(container.getEnv().get(7).getValue(), "cust-inst--gateway-service");

        Assert.assertEquals(container.getEnv().get(8).getName(), "CUSTOMERS_CONTEXT");
        Assert.assertEquals(container.getEnv().get(8).getValue(), "customers-1");

        Assert.assertEquals(container.getEnv().get(9).getName(), "PRODUCTS_CONTEXT");
        Assert.assertEquals(container.getEnv().get(9).getValue(), "products-1");

        Assert.assertEquals(container.getEnv().get(10).getName(), "DATABASE_USERNAME");
        Assert.assertEquals(container.getEnv().get(10).getValue(), "root");

        Assert.assertEquals(container.getEnv().get(11).getName(), "RATINGS_HOST");
        Assert.assertEquals(container.getEnv().get(11).getValue(), "review-inst--ratings-service");

        Assert.assertEquals(container.getEnv().get(12).getName(), "PRODUCTS_PORT");
        Assert.assertEquals(container.getEnv().get(12).getValue(), "80");

        Assert.assertEquals(container.getEnv().get(13).getName(), "DATABASE_PASSWORD");
        Assert.assertEquals(container.getEnv().get(13).getValue(), "root");

        Assert.assertEquals(container.getImage(), "celleryio/samples-productreview-reviews");
        Assert.assertEquals(container.getPorts().get(0).
                getContainerPort().intValue(), 8080);
        Assert.assertEquals(components.get(1).getMetadata().getName(), "ratings");
        final Container container1 = components.get(1).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container1.getPorts().get(0).getContainerPort().intValue(), 8080);
        Assert.assertEquals(container1.getEnv().get(0).getName(), "PORT");
        Assert.assertEquals(container1.getEnv().get(0).getValue(), "8080");
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
