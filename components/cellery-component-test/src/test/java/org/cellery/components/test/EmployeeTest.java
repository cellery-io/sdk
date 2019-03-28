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
package org.cellery.components.test;

import io.cellery.models.Cell;
import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
import org.ballerinax.kubernetes.utils.KubernetesUtils;
import org.cellery.components.test.utils.CelleryUtils;
import org.cellery.components.test.utils.LangTestUtils;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;

import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static org.cellery.components.test.utils.CelleryTestConstants.EMPLOYEE_PORTAL;
import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;

public class EmployeeTest {
    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve(EMPLOYEE_PORTAL+"/"+CELLERY+"/"+"employee");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Cell cell;

    @BeforeClass
    public void compileSample() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileBallerinaFile(SOURCE_DIR_PATH, "employee.bal",
                "test-org", "test-img", "1.3.5"), 0);
        File artifactYaml = TARGET_PATH.resolve("emp.yaml").toFile();
        Assert.assertTrue(artifactYaml.exists());
        cell = CelleryUtils.getInstance(CELLERY_PATH.resolve("emp.yaml").toString());
    }

    @Test
    public void validateCell() {
        Assert.assertNotNull(cell);
    }

    @AfterClass
    public void cleanUp() throws KubernetesPluginException {
        KubernetesUtils.deleteDirectory(String.valueOf(TARGET_PATH));
    }
}
