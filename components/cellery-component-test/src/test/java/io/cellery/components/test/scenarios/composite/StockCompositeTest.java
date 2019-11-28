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

package io.cellery.components.test.scenarios.composite;

import io.cellery.CelleryUtils;
import io.cellery.components.test.models.CellImageInfo;
import io.cellery.components.test.utils.LangTestUtils;
import io.cellery.models.ComponentSpec;
import io.cellery.models.Composite;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;

import static io.cellery.components.test.utils.CelleryTestConstants.BAL;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
import static io.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
import static io.cellery.components.test.utils.CelleryTestConstants.TARGET;
import static io.cellery.components.test.utils.CelleryTestConstants.YAML;
import static io.cellery.components.test.utils.LangTestUtils.deleteDirectory;

public class StockCompositeTest {

    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve("composite" + File.separator + "stock");
    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
    private Composite composite;
    private CellImageInfo imageInfo = new CellImageInfo("myorg", "stock-comp", "1.0.0", "stock-inst");

    @Test(groups = "build")
    public void compileCellBuild() throws IOException, InterruptedException {
        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "stock-comp" + BAL,
                imageInfo), 0);
        File artifactYaml = CELLERY_PATH.resolve(imageInfo.getName() + YAML).toFile();
        Assert.assertTrue(artifactYaml.exists());
        composite = CelleryUtils.readCompositeYaml(CELLERY_PATH.resolve(imageInfo.getName() + YAML).toString());
    }

    @Test(groups = "build")
    public void validateBuildTimeCellAvailability() {
        Assert.assertNotNull(composite);
    }

    @Test(groups = "build")
    public void validateBuildTimeAPIVersion() {
        Assert.assertEquals(composite.getApiVersion(), "mesh.cellery.io/v1alpha2");
    }

    @Test(groups = "build")
    public void validateBuildTimeMetaData() {
        Assert.assertEquals(composite.getMetadata().getName(), imageInfo.getName());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG),
                imageInfo.getOrg());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME),
                imageInfo.getName());
        Assert.assertEquals(composite.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION),
                imageInfo.getVer());
    }

    @Test(groups = "build")
    public void validateBuildTimeServiceTemplates() {
        Assert.assertEquals(composite.getSpec().getComponents().get(0).getMetadata().getName(), "stock");
        final ComponentSpec spec = composite.getSpec().getComponents().get(0).getSpec();
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getImage(),
                "wso2cellery/sampleapp-stock:0.3.0");
        Assert.assertEquals(spec.getTemplate().getContainers().get(0).getPorts().get(0).getContainerPort().intValue(),
                8080);
        Assert.assertEquals(spec.getPorts().get(0).getPort(), 80);
        Assert.assertEquals(spec.getPorts().get(0).getName(), "stock8080");
        Assert.assertEquals(spec.getPorts().get(0).getTargetPort(), 8080);
    }

    @AfterClass
    public void cleanUp() throws IOException {
        deleteDirectory(TARGET_PATH);
    }
}

