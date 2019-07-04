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

import io.cellery.CelleryUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BRefType;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;

import static io.cellery.CelleryConstants.INSTANCE_NAME;

/**
 * Native function cellery:stopInstances.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "stopInstances",
        args = {@Argument(name = "iName", type = TypeKind.RECORD),
                @Argument(name = "instanceList", type = TypeKind.ARRAY)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class StopInstances extends BlockingNativeCallableUnit {

    @Override
    public void execute(Context ctx) {
        BRefType<?>[] instanceList = ((BValueArray) ctx.getNullableRefArgument(1)).getValues();
        String instanceName;
        for (BRefType<?> refType: instanceList) {

            if (((BMap) refType).getMap().get(INSTANCE_NAME) == null) {
                break;
            }
            instanceName = ((BMap) refType).getMap().get(INSTANCE_NAME).toString();

            CelleryUtils.printInfo("Deleting " + instanceName + " instance...");
            CelleryUtils.executeShellCommand("kubectl delete cells.mesh.cellery.io " + instanceName, null,
                    CelleryUtils::printDebug, CelleryUtils::printWarning);
        }
    }
}
