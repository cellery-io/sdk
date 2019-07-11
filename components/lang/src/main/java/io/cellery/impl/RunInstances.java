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
package io.cellery.impl;

import io.cellery.CelleryConstants;
import io.cellery.CelleryUtils;
import io.cellery.models.CellImage;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.connector.api.BLangConnectorSPIUtil;
import org.ballerinalang.model.types.BArrayType;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;

import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.VERSION;

/**
 * Native function cellery:runInstances.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "runInstances",
        args = {@Argument(name = "iName", type = TypeKind.RECORD),
                @Argument(name = "instances", type = TypeKind.MAP)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class RunInstances extends BlockingNativeCallableUnit {

    @Override
    public void execute(Context ctx) {
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(0)).getMap();
        LinkedHashMap depStruct = ((BMap) ctx.getNullableRefArgument(1)).getMap();
        Map<String, CellImage> instances = new HashMap<>();
        CellImage cellImage;
        BArrayType bArrayType =
                new BArrayType(ctx.getProgramFile().getPackageInfo(CelleryConstants.CELLERY_PACKAGE).getTypeDefInfo(
                        CelleryConstants.IMAGE_NAME_DEFINITION).typeInfo.getType());
        BValueArray bValueArray = new BValueArray(bArrayType);
        for (Object object : depStruct.entrySet()) {
            cellImage = new CellImage();
            LinkedHashMap depNameStruct = ((BMap) ((Map.Entry) object).getValue()).getMap();
            String depAlias = ((Map.Entry) object).getKey().toString();
            cellImage.setOrgName(depNameStruct.get(ORG).toString());
            cellImage.setCellName(depNameStruct.get(INSTANCE_NAME).toString());
            cellImage.setCellVersion(depNameStruct.get(VERSION).toString());
            instances.put(depAlias, cellImage);
        }
        String output = CelleryUtils.executeShellCommand("kubectl get cells", null, CelleryUtils::printDebug,
                CelleryUtils::printDebug);
        StringBuilder runCommand;
        String imageName = nameStruct.get(ORG).toString() + "/" + nameStruct.get(NAME)
                + ":" + nameStruct.get(VERSION);

        AtomicLong runCount = new AtomicLong(0L);
        if (!output.contains(nameStruct.get(INSTANCE_NAME).toString() + " ")) {
            CelleryUtils.printInfo(nameStruct.get(INSTANCE_NAME).toString() + " instance not found.");
            BMap<String, BValue> bmap = BLangConnectorSPIUtil.createBStruct(ctx,
                    CelleryConstants.CELLERY_PACKAGE,
                    CelleryConstants.IMAGE_NAME_DEFINITION,
                    nameStruct.get(ORG).toString(), nameStruct.get(NAME),
                    nameStruct.get(VERSION), nameStruct.get(INSTANCE_NAME));
            bValueArray.add(runCount.getAndIncrement(), bmap);
            runCommand = new StringBuilder(
                    "cellery run " + imageName + " -n " + nameStruct.get(INSTANCE_NAME) + " -d");
            instances.forEach((key, value) -> {
                runCommand.append(" -l ").append(key).append(":").append(value.getCellName());
                if (!output.contains(value.getCellName())) {
                    CelleryUtils.printInfo(value.getCellName() + " instance not found.");
                    BMap<String, BValue> bmap1 = BLangConnectorSPIUtil.createBStruct(ctx,
                            CelleryConstants.CELLERY_PACKAGE,
                            CelleryConstants.IMAGE_NAME_DEFINITION,
                            value.getOrgName(), value.getCellName(), value.getCellVersion(), value.getCellName());
                    bValueArray.add(runCount.getAndIncrement(), bmap1);
                }
            });
            runCommand.append(" -y");
            CelleryUtils.printInfo("Creating test instances...");
            CelleryUtils.executeShellCommand(runCommand.toString(), null, CelleryUtils::printDebug,
                    CelleryUtils::printWarning);
        } else {
            CelleryUtils.printInfo(nameStruct.get(INSTANCE_NAME) +
                    " instance found created. Using existing instances and dependencies for testing");
        }
        ctx.setReturnValues(bValueArray);
    }
}
