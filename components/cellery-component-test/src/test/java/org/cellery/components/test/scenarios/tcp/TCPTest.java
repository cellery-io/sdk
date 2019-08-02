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

package org.cellery.components.test.scenarios.tcp;

import io.cellery.CelleryUtils;
import io.cellery.models.Cell;
import io.cellery.models.GatewaySpec;
import io.cellery.models.ServiceTemplateSpec;
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

import static org.cellery.components.test.utils.CelleryTestConstants.BAL;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static org.cellery.components.test.utils.CelleryTestConstants.YAML;

public class TCPTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("grpc-tcp");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;
    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "database", "1.0.0", "db-inst");

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "grpc-tcp" + BAL, cellImageInfo),
                0);
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
        Assert.assertEquals(cell.getApiVersion(), "mesh.cellery.io/v1alpha1");
    }

    @Test(groups = "build")
    public void validateBuildTimeMetaData() {
        Assert.assertEquals(cell.getMetadata().getName(), cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
    }

    @Test(groups = "build")
    public void validateBuildTimeGatewayTemplate() {
        final GatewaySpec spec = cell.getSpec().getGatewayTemplate().getSpec();
        Assert.assertEquals(spec.getType(), "Envoy");
        Assert.assertEquals(spec.getTcp().get(0).getBackendPort(), 3306);
        Assert.assertEquals(spec.getTcp().get(0).getPort(), 31406);
        Assert.assertEquals(spec.getGrpc().size(), 0);
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getMetadata().getName(), "mysql");
        final ServiceTemplateSpec spec = cell.getSpec().getServicesTemplates().get(0).getSpec();
        Assert.assertEquals(spec.getContainer().getEnv().get(0).getName(), "MYSQL_ROOT_PASSWORD");
        Assert.assertEquals(spec.getContainer().getEnv().get(0).getValue(), "root");
        Assert.assertEquals(spec.getContainer().getImage(), "mirage20/samples-productreview-mysql");
        Assert.assertEquals(spec.getContainer().getPorts().get(0).getContainerPort().intValue(), 3306);
        Assert.assertEquals(spec.getProtocol(), "TCP");
        Assert.assertEquals(spec.getReplicas(), 1);
        Assert.assertEquals(spec.getServicePort(), 3306);

        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getMetadata().getName(), "grpc");
        final ServiceTemplateSpec grpcCompSpec = cell.getSpec().getServicesTemplates().get(1).getSpec();
        Assert.assertEquals(grpcCompSpec.getContainer().getImage(), "mirage20/samples-productreview-mysql");
        Assert.assertEquals(grpcCompSpec.getContainer().getPorts().get(0).getContainerPort().intValue(), 4406);
        Assert.assertEquals(grpcCompSpec.getProtocol(), "GRPC");
        Assert.assertEquals(grpcCompSpec.getReplicas(), 1);
        Assert.assertEquals(grpcCompSpec.getServicePort(), 4406);
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(TARGET_PATH);
    }
}
