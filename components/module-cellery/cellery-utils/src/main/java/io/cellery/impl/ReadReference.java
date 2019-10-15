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
import org.apache.commons.io.IOUtils;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.connector.api.BLangConnectorSPIUtil;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.model.values.BValue;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Enumeration;
import java.util.LinkedHashMap;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;

import static io.cellery.CelleryConstants.CELLERY_REPO_PATH;
import static io.cellery.CelleryConstants.INSTANCE_NAME_PLACEHOLDER;
import static io.cellery.CelleryConstants.REFERENCE_FILE_NAME;

/**
 * Native function cellery:ReadReference.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "readReference",
        args = {@Argument(name = "cellImage", type = TypeKind.RECORD),
                @Argument(name = "dependencyName", type = TypeKind.STRING)},
        returnType = {@ReturnType(type = TypeKind.OBJECT), @ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class ReadReference extends BlockingNativeCallableUnit {

    public void execute(Context ctx) {
        LinkedHashMap nameStruct = ((BMap) ctx.getNullableRefArgument(0)).getMap();
        String orgName = ((BString) nameStruct.get("org")).stringValue();
        String cellName = ((BString) nameStruct.get("name")).stringValue();
        String cellVersion = ((BString) nameStruct.get("ver")).stringValue();
        String instanceName = ((BString) nameStruct.get("instanceName")).stringValue();
        String zipFilePath = CELLERY_REPO_PATH + File.separator + orgName + File.separator + cellName + File.separator
                + cellVersion + File.separator + cellName + ".zip";
        JSONObject jsonObject;
        try {
            jsonObject = readReferenceJSON(zipFilePath);
            if (jsonObject == null || jsonObject.isEmpty()) {
                ctx.setError(BLangVMErrors.createError(ctx, "Reference file is empty. " + zipFilePath));
                return;
            }
        } catch (IOException e) {
            ctx.setError(BLangVMErrors.createError(ctx, e.getMessage()));
            return;
        }
        BMap<String, BValue> refMap = BLangConnectorSPIUtil.createBStruct(ctx, CelleryConstants.CELLERY_PACKAGE,
                CelleryConstants.REFERENCE_DEFINITION);
        jsonObject.keys().forEachRemaining(key -> refMap.put(key,
                new BString(jsonObject.get(key).toString().replace(INSTANCE_NAME_PLACEHOLDER,
                        "{{" + instanceName + "}}"))));
        ctx.setReturnValues(refMap);
    }

    private JSONObject readReferenceJSON(String zipFilePath) throws IOException {
        try (ZipFile zipFile = new ZipFile(zipFilePath)) {
            Enumeration<? extends ZipEntry> entries = zipFile.entries();
            String fileName = "artifacts" + File.separator + "ref" + File.separator + REFERENCE_FILE_NAME;
            while (entries.hasMoreElements()) {
                ZipEntry entry = entries.nextElement();
                if (entry.getName().matches(fileName)) {
                    try (InputStream stream = zipFile.getInputStream(entry)) {
                        return new JSONObject(IOUtils.toString(stream, StandardCharsets.UTF_8));
                    }
                }
            }
        }
        return null;
    }
}
