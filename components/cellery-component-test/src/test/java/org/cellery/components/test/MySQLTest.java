/*
 * Copyright (c) 2018, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package org.cellery.components.test;

import io.cellery.models.Cell;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.utils.CelleryUtils;
import org.cellery.components.test.utils.KubernetesTestUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;

public class MySQLTest implements SampleTest {
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("employee-portal/cellery/mysql");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve("target");
    private static final Path CELLERY_PATH = TARGET_PATH.resolve("cellery");
    private Cell cell;

    @BeforeClass
    public void compileSample() throws IOException, InterruptedException {
        Assert.assertEquals(KubernetesTestUtils.compileBallerinaFile(SOURCE_DIR_PATH, "mysql.bal",
                "test-org", "mysql-img", "1.3.5"), 0);
        File artifactYaml = CELLERY_PATH.resolve("mysql-img.yaml").toFile();
        Assert.assertTrue(artifactYaml.exists());
        CelleryUtils cellObj = new CelleryUtils();
        cell = cellObj.getInstance(CELLERY_PATH.resolve("mysql-img.yaml").toString());
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
    public void validateMetaData() {
        Assert.assertEquals(cell.getMetadata().getName(), "mysql-img");
        Assert.assertEquals(cell.getMetadata().getAnnotations().get("mesh.cellery.io/cell-image-org"), "test-org");
        Assert.assertEquals(cell.getMetadata().getAnnotations().get("mesh.cellery.io/cell-image-name"), "mysql-img");
        Assert.assertEquals(cell.getMetadata().getAnnotations().get("mesh.cellery.io/cell-image-version"), "1.3.5");
    }

    @Test
    public void validateGatewayTemplate() {
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getTcp().get(0).getBackendHost(), "mysql");
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getTcp().get(0).getBackendPort(), 31406);
        Assert.assertEquals(cell.getSpec().getGatewayTemplate().getSpec().getTcp().get(0).getPort(), 3306);
    }

    @Test
    public void validateServicesTemplates() {
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getMetadata().getName(), "mysql");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(0).
                getName(), "MYSQL_ROOT_PASSWORD");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getEnv().get(0).
                getValue(), "root");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getImage(),
                "mirage20/samples-productreview-mysql");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getContainer().getPorts().get(0).
                getContainerPort().intValue(), 3306);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getProtocol(), "TCP");
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getReplicas(), 1);
        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getSpec().getServicePort(), 3306);
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(String.valueOf(TARGET_PATH));
    }
}

