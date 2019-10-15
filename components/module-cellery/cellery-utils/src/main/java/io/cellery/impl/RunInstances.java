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
package io.cellery.impl;

import org.ballerinalang.jvm.values.MapValue;

/**
 * Native function cellery:runInstances.
 */
public class RunInstances {
    public static void runInstances(MapValue nameStruct, MapValue depStruct) {
//        Map<String, Image> instances = new HashMap<>();
//        Image image;
//        BArrayType bArrayType =
//                new BArrayType(ctx.getProgramFile().getPackageInfo(CelleryConstants.CELLERY_PACKAGE).getTypeDefInfo(
//                        CelleryConstants.IMAGE_NAME_DEFINITION).typeInfo.getType());
//        BValueArray bValueArray = new BValueArray(bArrayType);
//        for (Object object : depStruct.entrySet()) {
//            image = new Image();
//            LinkedHashMap depNameStruct = ((BMap) ((Map.Entry) object).getValue()).getMap();
//            String depAlias = ((Map.Entry) object).getKey().toString();
//            image.setOrgName(depNameStruct.get(CelleryConstants.ORG).toString());
//            image.setCellName(depNameStruct.get(CelleryConstants.INSTANCE_NAME).toString());
//            image.setCellVersion(depNameStruct.get(CelleryConstants.VERSION).toString());
//            instances.put(depAlias, image);
//        }
//        String output = CelleryUtils.executeShellCommand("kubectl get cells", null, CelleryUtils::printDebug,
//                CelleryUtils::printDebug);
//        StringBuilder runCommand;
//        String imageName = nameStruct.get(CelleryConstants.ORG).toString() + "/" + nameStruct.get(CelleryConstants
//        .NAME)
//                + ":" + nameStruct.get(CelleryConstants.VERSION);
//
//        AtomicLong runCount = new AtomicLong(0L);
//        if (!output.contains(nameStruct.get(CelleryConstants.INSTANCE_NAME).toString() + " ")) {
//            CelleryUtils.printInfo(nameStruct.get(CelleryConstants.INSTANCE_NAME).toString() + " instance not found
//            .");
//            BMap<String, BValue> bmap = BLangConnectorSPIUtil.createBStruct(ctx,
//                    CelleryConstants.CELLERY_PACKAGE,
//                    CelleryConstants.IMAGE_NAME_DEFINITION,
//                    nameStruct.get(CelleryConstants.ORG).toString(), nameStruct.get(CelleryConstants.NAME),
//                    nameStruct.get(CelleryConstants.VERSION), nameStruct.get(CelleryConstants.INSTANCE_NAME));
//            bValueArray.add(runCount.getAndIncrement(), bmap);
//            runCommand = new StringBuilder(
//                    "cellery run " + imageName + " -n " + nameStruct.get(CelleryConstants.INSTANCE_NAME) + " -d");
//            instances.forEach((key, value) -> {
//                runCommand.append(" -l ").append(key).append(":").append(value.getCellName());
//                if (!output.contains(value.getCellName())) {
//                    CelleryUtils.printInfo(value.getCellName() + " instance not found.");
//                    BMap<String, BValue> bmap1 = BLangConnectorSPIUtil.createBStruct(ctx,
//                            CelleryConstants.CELLERY_PACKAGE,
//                            CelleryConstants.IMAGE_NAME_DEFINITION,
//                            value.getOrgName(), value.getCellName(), value.getCellVersion(), value.getCellName());
//                    bValueArray.add(runCount.getAndIncrement(), bmap1);
//                }
//            });
//            runCommand.append(" -y");
//            CelleryUtils.printInfo("Creating test instances...");
//            CelleryUtils.executeShellCommand(runCommand.toString(), null, CelleryUtils::printDebug,
//                    CelleryUtils::printWarning);
//        } else {
//            CelleryUtils.printInfo(nameStruct.get(CelleryConstants.INSTANCE_NAME) +
//                    " instance found created. Using existing instances and dependencies for testing");
//        }
//        ctx.setReturnValues(bValueArray);
    }
}
