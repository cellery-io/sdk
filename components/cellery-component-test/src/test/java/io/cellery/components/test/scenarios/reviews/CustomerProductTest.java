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

package io.cellery.components.test.scenarios.reviews;

import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.API;
import io.cellery.models.Cell;
import io.cellery.models.GRPC;
import io.cellery.models.Ingress;
import io.fabric8.kubernetes.api.model.Container;
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

import static io.cellery.components.test.utils.CelleryTestConstants.ARTIFACTS;
import static io.cellery.components.test.utils.CelleryTestConstants.BAL;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_MESH_VERSION;
import static io.cellery.components.test.utils.CelleryTestConstants.PRODUCT_REVIEW;
import static io.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static io.cellery.components.test.utils.CelleryTestConstants.YAML;
import static io.cellery.components.test.utils.LangTestUtils.deleteDirectory;

public class CustomerProductTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH =
            SAMPLE_DIR.resolve(PRODUCT_REVIEW + File.separator + CELLERY + File.separator +
                    "customer-products");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private Cell runtimeCell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "products", "1.0.0", "cust-inst");
    private Map<String, CellImageInfo> dependencyCells = new HashMap<>();

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "customer-products" + BAL
                , cellImageInfo), 0);
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
        final GRPC grpc = cell.getSpec().getGateway().getSpec().getIngress().getGrpc().get(0);
        Assert.assertEquals(grpc.getDestination().getHost(), "categories");
        Assert.assertEquals(grpc.getDestination().getPort(), 8000);
        Assert.assertEquals(grpc.getPort(), 8000);

        final API api = cell.getSpec().getGateway().getSpec().getIngress().getHttp().get(0);
        Assert.assertEquals(api.getDestination().getHost(), "customers");
        Assert.assertEquals(api.getContext(), "/customers-1");
        Assert.assertEquals(api.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api.getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(api.isAuthenticate());

        final API api1 = cell.getSpec().getGateway().getSpec().getIngress().getHttp().get(1);
        Assert.assertEquals(api1.getDestination().getHost(), "products");
        Assert.assertEquals(api1.getContext(), "/products-1");
        Assert.assertEquals(api1.getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(api1.getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(api.isAuthenticate());

    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(cell.getSpec().getComponents().get(0).getMetadata().getName(),
                "customers");
        final Container container =
                cell.getSpec().getComponents().get(0).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container.getEnv().get(0).getName(), "PORT");
        Assert.assertEquals(container.getEnv().get(0).getValue(), "8080");
        Assert.assertEquals(container.getImage(), "celleryio/samples-productreview-customers");
        Assert.assertEquals(container.getPorts().get(0).getContainerPort().intValue(), 8080);

        Assert.assertEquals(cell.getSpec().getComponents().get(1).getMetadata().getName(),
                "categories");
        final List<Container> containers =
                cell.getSpec().getComponents().get(1).getSpec().getTemplate().getContainers();
        Assert.assertEquals(containers.get(0).getEnv().get(0).getName(), "PORT");
        Assert.assertEquals(containers.get(0).getEnv().get(0).getValue(), "8000");
        Assert.assertEquals(containers.get(0).getImage(), "celleryio/samples-productreview-categories");
        Assert.assertEquals(containers.get(0).getPorts().get(0).getContainerPort().intValue(), 8000);

        Assert.assertEquals(cell.getSpec().getComponents().get(2).getMetadata().getName(), "products");
        final Container container1 =
                cell.getSpec().getComponents().get(2).getSpec().getTemplate().getContainers().get(0);
        Assert.assertEquals(container1.getEnv().get(0).getName(), "CATEGORIES_PORT");
        Assert.assertEquals(container1.getEnv().get(0).getValue(), "8000");
        Assert.assertEquals(container1.getEnv().get(1).getName(), "PORT");
        Assert.assertEquals(container1.getEnv().get(1).getValue(), "8080");
        Assert.assertEquals(container1.getEnv().get(2).getName(), "CATEGORIES_HOST");
        Assert.assertEquals(container1.getEnv().get(2).getValue(), "{{instance_name}}--categories-service");
        Assert.assertEquals(container1.getImage(), "celleryio/samples-productreview-products");
        Assert.assertEquals(container1.getPorts().get(0).getContainerPort().intValue(), 8080);
    }

    @Test(groups = "run")
    public void compileCellRun() throws IOException, InterruptedException {
        String tmpDir = LangTestUtils.createTempImageDir(SOURCE_DIR_PATH, cellImageInfo.getName());
        Path tempPath = Paths.get(tmpDir);
        Assert.assertEquals(LangTestUtils.compileCellRunFunction(SOURCE_DIR_PATH, "customer-products" + BAL
                , cellImageInfo, dependencyCells, tmpDir), 0);
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
        Assert.assertEquals(ingress.getGrpc().get(0).getDestination().getHost(), "categories");
        Assert.assertEquals(ingress.getGrpc().get(0).getDestination().getPort(), 8000);
        Assert.assertEquals(ingress.getGrpc().get(0).getPort(), 8000);

        Assert.assertEquals(ingress.getHttp().get(0).getDestination().getHost(), "customers");
        Assert.assertEquals(ingress.getHttp().get(0).getContext(), "/customers-1");
        Assert.assertEquals(ingress.getHttp().get(0).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(ingress.getHttp().get(0).getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(ingress.getHttp().get(0).isAuthenticate());

        Assert.assertEquals(ingress.getHttp().get(1).getDestination().getHost(), "products");
        Assert.assertEquals(ingress.getHttp().get(1).getContext(), "/products-1");
        Assert.assertEquals(ingress.getHttp().get(1).getDefinitions().get(0).getMethod(), "GET");
        Assert.assertEquals(ingress.getHttp().get(1).getDefinitions().get(0).getPath(), "/*");
        Assert.assertTrue(ingress.getHttp().get(0).isAuthenticate());

    }

    @Test(groups = "run")
    public void validateRunTimeServiceTemplates() {
        Assert.assertEquals(runtimeCell.getSpec().getComponents().get(0).getMetadata().getName(),
                "customers");
        final List<Container> containers =
                runtimeCell.getSpec().getComponents().get(0).getSpec().getTemplate().getContainers();
        Assert.assertEquals(containers.get(0).getEnv().get(0).getName(), "PORT");
        Assert.assertEquals(containers.get(0).getEnv().get(0).getValue(), "8080");
        Assert.assertEquals(containers.get(0).getImage(), "celleryio/samples-productreview-customers");
        Assert.assertEquals(containers.get(0).getPorts().get(0).getContainerPort().intValue(), 8080);

        Assert.assertEquals(runtimeCell.getSpec().getComponents().get(1).getMetadata().getName(),
                "categories");
        final List<Container> containers1 =
                runtimeCell.getSpec().getComponents().get(1).getSpec().getTemplate().getContainers();
        Assert.assertEquals(containers1.get(0).getEnv().get(0).getName(), "PORT");
        Assert.assertEquals(containers1.get(0).getEnv().get(0).getValue(), "8000");
        Assert.assertEquals(containers1.get(0).getImage(), "celleryio/samples-productreview-categories");
        Assert.assertEquals(containers1.get(0).getPorts().get(0).getContainerPort().intValue(), 8000);

        Assert.assertEquals(runtimeCell.getSpec().getComponents().get(2).getMetadata().getName(),
                "products");
        final List<Container> containers2 =
                runtimeCell.getSpec().getComponents().get(2).getSpec().getTemplate().getContainers();
        Assert.assertEquals(containers2.get(0).getEnv().get(0).getName(), "CATEGORIES_PORT");
        Assert.assertEquals(containers2.get(0).getEnv().get(0).getValue(), "8000");
        Assert.assertEquals(containers2.get(0).getEnv().get(1).getName(), "PORT");
        Assert.assertEquals(containers2.get(0).getEnv().get(1).getValue(), "8080");
        Assert.assertEquals(containers2.get(0).getEnv().get(2).getName(), "CATEGORIES_HOST");
        Assert.assertEquals(containers2.get(0).getEnv().get(2).getValue(), "cust-inst--categories-service");
        Assert.assertEquals(containers2.get(0).getImage(), "celleryio/samples-productreview-products");
        Assert.assertEquals(containers2.get(0).getPorts().get(0).getContainerPort().intValue(), 8080);
    }

    @AfterClass
    public void cleanUp() throws IOException {
        deleteDirectory(TARGET_PATH);
    }
}
