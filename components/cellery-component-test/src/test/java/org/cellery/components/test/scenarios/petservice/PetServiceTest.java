///*
// * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
// *
// * WSO2 Inc. licenses this file to you under the Apache License,
// * Version 2.0 (the "License"); you may not use this file except
// * in compliance with the License.
// * You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on an
// * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// * KIND, either express or implied.  See the License for the
// * specific language governing permissions and limitations
// * under the License.
// */
//
//package org.cellery.components.test.scenarios.petservice;
//
//import io.cellery.CelleryUtils;
//import io.cellery.models.AutoScalingPolicy;
//import io.cellery.models.Cell;
//import io.cellery.models.ServiceTemplateSpec;
//import org.ballerinax.kubernetes.exceptions.KubernetesPluginException;
//import org.ballerinax.kubernetes.utils.KubernetesUtils;
//import org.cellery.components.test.models.CellImageInfo;
//import org.cellery.components.test.utils.LangTestUtils;
//import org.testng.Assert;
//import org.testng.annotations.AfterClass;
//import org.testng.annotations.Test;
//
//import java.io.File;
//import java.io.IOException;
//import java.nio.file.Path;
//import java.nio.file.Paths;
//
//import static io.cellery.CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_CPU;
//import static io.cellery.CelleryConstants.AUTO_SCALING_METRIC_RESOURCE_MEMORY;
//import static org.cellery.components.test.utils.CelleryTestConstants.BAL;
//import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY;
//import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_NAME;
//import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_ORG;
//import static org.cellery.components.test.utils.CelleryTestConstants.CELLERY_IMAGE_VERSION;
//import static org.cellery.components.test.utils.CelleryTestConstants.PET_SERVICE;
//import static org.cellery.components.test.utils.CelleryTestConstants.TARGET;
//import static org.cellery.components.test.utils.CelleryTestConstants.YAML;
//
//public class PetServiceTest {
//
//    private static final Path SAMPLE_DIR = Paths.get(System.getProperty("sample.dir"));
//    private static final Path SOURCE_DIR_PATH = SAMPLE_DIR.resolve(PET_SERVICE);
//    private static final Path TARGET_PATH = SOURCE_DIR_PATH.resolve(TARGET);
//    private static final Path CELLERY_PATH = TARGET_PATH.resolve(CELLERY);
//    private Cell cell;
//    private CellImageInfo cellImageInfo = new CellImageInfo("myorg", "petservice", "1.0.0", "petsvc-inst");
//
//    @Test(groups = "build")
//    public void compileCellBuild() throws IOException, InterruptedException {
//        Assert.assertEquals(LangTestUtils.compileCellBuildFunction(SOURCE_DIR_PATH, "pet-cell" + BAL, cellImageInfo),
//                0);
//        File artifactYaml = CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toFile();
//        Assert.assertTrue(artifactYaml.exists());
//        cell = CelleryUtils.readCellYaml(CELLERY_PATH.resolve(cellImageInfo.getName() + YAML).toString());
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeCellAvailability() {
//        Assert.assertNotNull(cell);
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeAPIVersion() {
//        Assert.assertEquals(cell.getApiVersion(), "mesh.cellery.io/v1alpha1");
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeKind() {
//        Assert.assertEquals(cell.getKind(), "Cell");
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeMetaData() {
//        Assert.assertEquals(cell.getMetadata().getName(), cellImageInfo.getName());
//        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_ORG), cellImageInfo.getOrg());
//        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_NAME), cellImageInfo.getName());
//        Assert.assertEquals(cell.getMetadata().getAnnotations().get(CELLERY_IMAGE_VERSION), cellImageInfo.getVer());
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeGatewayTemplate() {
//        Assert.assertEquals(cell.getSpec().getGateway().getSpec().getHttp().get(0).getContext(),
//                "petsvc");
//        Assert.assertEquals(cell.getSpec().getGateway().getSpec().getType(), "MicroGateway");
//    }
//
//    @Test(groups = "build")
//    public void validateBuildTimeServiceTemplates() {
//        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(0).getMetadata().getName(), "debug");
//        final ServiceTemplateSpec debugSpec = cell.getSpec().getServicesTemplates().get(0).getSpec();
//        Assert.assertEquals(debugSpec.getContainer().getImage(), "docker.io/mirage20/k8s-debug-tools");
//        Assert.assertEquals(debugSpec.getContainer().getPorts().size(), 0);
//        Assert.assertEquals(debugSpec.getReplicas(), 1);
//        Assert.assertTrue(debugSpec.getAutoscaling().isOverridable());
//        final AutoScalingPolicy autoscalePolicy = debugSpec.getAutoscaling().getPolicy();
//        Assert.assertEquals(autoscalePolicy.getMaxReplicas(), 10);
//        Assert.assertEquals(autoscalePolicy.getMinReplicas(), 1);
//        Assert.assertEquals(autoscalePolicy.getMetrics().get(0).getType(), "Resource");
//        Assert.assertEquals(autoscalePolicy.getMetrics().get(0).getResource().getName(),
//                AUTO_SCALING_METRIC_RESOURCE_CPU);
//        Assert.assertEquals(autoscalePolicy.getMetrics().get(0).getResource().getTargetAverageValue(), "500m");
//        Assert.assertEquals(autoscalePolicy.getMetrics().get(1).getResource().getName(),
//                AUTO_SCALING_METRIC_RESOURCE_MEMORY);
//        Assert.assertEquals(autoscalePolicy.getMetrics().get(1).getResource().getTargetAverageUtilization(), 50);
//
//
//        final ServiceTemplateSpec petSpec = cell.getSpec().getServicesTemplates().get(1).getSpec();
//        Assert.assertEquals(cell.getSpec().getServicesTemplates().get(1).getMetadata().getName(), "pet-service");
//        Assert.assertEquals(petSpec.getContainer().getImage(), "docker.io/isurulucky/pet-service");
//        Assert.assertEquals(petSpec.getContainer().getPorts().get(0).getContainerPort().intValue(), 9090);
//        Assert.assertEquals(petSpec.getReplicas(), 1);
//        Assert.assertEquals(petSpec.getServicePort(), 80);
//        Assert.assertFalse(petSpec.getAutoscaling().isOverridable());
//        final AutoScalingPolicy zeroScalePolicy = petSpec.getAutoscaling().getPolicy();
//        Assert.assertEquals(zeroScalePolicy.getMaxReplicas(), 10);
//        Assert.assertEquals(zeroScalePolicy.getConcurrency(), 25);
//        Assert.assertEquals(zeroScalePolicy.getMinReplicas(), 0);
//    }
//
//    @AfterClass
//    public void cleanUp() throws KubernetesPluginException {
//        KubernetesUtils.deleteDirectory(TARGET_PATH);
//    }
//}
