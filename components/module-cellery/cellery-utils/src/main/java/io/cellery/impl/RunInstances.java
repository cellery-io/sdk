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

import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.models.internal.Image;
import org.ballerinalang.jvm.BallerinaValues;
import org.ballerinalang.jvm.types.BArrayType;
import org.ballerinalang.jvm.types.BPackage;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryConstants.CELLERY_PKG_NAME;
import static io.cellery.CelleryConstants.CELLERY_PKG_ORG;
import static io.cellery.CelleryConstants.CELLERY_PKG_VERSION;
import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.VERSION;

/**
 * Native function cellery:runInstances.
 */
public class RunInstances {

    public static ArrayValue runInstances(MapValue nameStruct, MapValue<?, ?> depStruct) {
        Map<String, Image> instances = new HashMap<>();

        ArrayValue bValueArray = new ArrayValue(new BArrayType(BallerinaValues.createRecordValue(new BPackage(
                        CELLERY_PKG_ORG, CELLERY_PKG_NAME, CELLERY_PKG_VERSION),
                CelleryConstants.IMAGE_NAME_DEFINITION).getType()));
        depStruct.forEach((key, value) -> {
            Image image = new Image();
            MapValue depNameStruct = (MapValue<?, ?>) value;
            String depAlias = key.toString();
            image.setOrgName(depNameStruct.getStringValue(CelleryConstants.ORG));
            image.setCellName(depNameStruct.getStringValue(CelleryConstants.INSTANCE_NAME));
            image.setCellVersion(depNameStruct.getStringValue(CelleryConstants.VERSION));
            instances.put(depAlias, image);
        });

        String output = CelleryUtils.executeShellCommand("kubectl get cells", null, CelleryUtils::printDebug,
                CelleryUtils::printDebug);
        StringBuilder runCommand;
        String imageName = nameStruct.get(CelleryConstants.ORG).toString() + "/" + nameStruct.get(CelleryConstants
                .NAME) + ":" + nameStruct.get(CelleryConstants.VERSION);

        AtomicLong runCount = new AtomicLong(0L);
        if (!output.contains(nameStruct.get(CelleryConstants.INSTANCE_NAME).toString() + " ")) {
            CelleryUtils.printInfo(nameStruct.get(CelleryConstants.INSTANCE_NAME).toString() + " instance not found.");
            MapValue<String, Object> bmap = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                    CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.IMAGE_NAME_DEFINITION);
            bmap.put(ORG, nameStruct.getStringValue(CelleryConstants.ORG));
            bmap.put(NAME, nameStruct.getStringValue(CelleryConstants.NAME));
            bmap.put(VERSION, nameStruct.getStringValue(CelleryConstants.VERSION));
            bmap.put(INSTANCE_NAME, nameStruct.getStringValue(CelleryConstants.INSTANCE_NAME));
            bValueArray.add(runCount.getAndIncrement(), bmap);
            runCommand = new StringBuilder(
                    "cellery run " + imageName + " -n " + nameStruct.get(CelleryConstants.INSTANCE_NAME) + " -d");
            instances.forEach((key, value) -> {
                runCommand.append(" -l ").append(key).append(":").append(value.getCellName());
                if (!output.contains(value.getCellName())) {
                    CelleryUtils.printInfo(value.getCellName() + " instance not found.");
                    MapValue<String, Object> bmap1 = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                            CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.IMAGE_NAME_DEFINITION);
                    bmap1.put(ORG, value.getOrgName());
                    bmap1.put(NAME, value.getCellName());
                    bmap1.put(VERSION, value.getCellVersion());
                    bmap1.put(INSTANCE_NAME, value.getCellName());
                    bValueArray.add(runCount.getAndIncrement(), bmap1);
                }
            });
            runCommand.append(" -y");
            CelleryUtils.printInfo("Creating test instances...");
            CelleryUtils.executeShellCommand(runCommand.toString(), null, CelleryUtils::printDebug,
                    CelleryUtils::printWarning);
        } else {
            CelleryUtils.printInfo(nameStruct.get(CelleryConstants.INSTANCE_NAME) +
                    " instance found created. Using existing instances and dependencies for testing");
        }
        return bValueArray;
    }
}
