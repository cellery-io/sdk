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

import org.ballerinalang.jvm.values.ArrayValue;

/**
 * Native function cellery:stopInstances.
 */
public class StopInstances {

    public static void stopInstances(ArrayValue instanceList) {
//        BRefType<?> iNameRefType;
//        LinkedHashMap nameStruct = null;
//        String instanceName;
//        boolean wasRunning;
//
//        for (BRefType<?> refType : instanceList.getValues()) {
//            iNameRefType = (BRefType<?>) ((BMap) refType).getMap().get("iName");
//            if (((BMap) iNameRefType).getMap().get(INSTANCE_NAME) == null) {
//                break;
//            }
//            nameStruct = ((BMap) iNameRefType).getMap();
//            wasRunning = ((BBoolean) ((BMap) refType).getMap().get("isRunning")).booleanValue();
//            if (!wasRunning) {
//                instanceName = ((BString) nameStruct.get(INSTANCE_NAME)).stringValue();
//                printInfo("Deleting " + instanceName + " instance...");
//                CelleryUtils.executeShellCommand("kubectl delete cells.mesh.cellery.io " + instanceName, null,
//                        CelleryUtils::printDebug, CelleryUtils::printWarning);
//            }
//        }
    }
}
