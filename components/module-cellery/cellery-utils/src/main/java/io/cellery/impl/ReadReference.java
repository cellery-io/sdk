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
import org.ballerinalang.jvm.BallerinaValues;
import org.ballerinalang.jvm.types.BPackage;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Enumeration;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;

import static io.cellery.CelleryConstants.CELLERY_PKG_NAME;
import static io.cellery.CelleryConstants.CELLERY_PKG_ORG;
import static io.cellery.CelleryConstants.CELLERY_PKG_VERSION;
import static io.cellery.CelleryConstants.CELLERY_REPO_PATH;
import static io.cellery.CelleryConstants.INSTANCE_NAME;
import static io.cellery.CelleryConstants.INSTANCE_NAME_PLACEHOLDER;
import static io.cellery.CelleryConstants.NAME;
import static io.cellery.CelleryConstants.ORG;
import static io.cellery.CelleryConstants.REFERENCE_FILE_NAME;
import static io.cellery.CelleryConstants.VERSION;

/**
 * Native function cellery:ReadReference.
 */
public class ReadReference {

    public static MapValue readReferenceExternal(MapValue nameStruct) {
        String orgName = nameStruct.getStringValue(ORG);
        String cellName = nameStruct.getStringValue(NAME);
        String cellVersion = nameStruct.getStringValue(VERSION);
        String instanceName = nameStruct.getStringValue(INSTANCE_NAME);
        String zipFilePath = CELLERY_REPO_PATH + File.separator + orgName + File.separator + cellName + File.separator
                + cellVersion + File.separator + cellName + ".zip";
        JSONObject jsonObject;
        try {
            jsonObject = readReferenceJSON(zipFilePath);
            if (jsonObject == null || jsonObject.isEmpty()) {
                throw new BallerinaException("Reference file is empty. " + zipFilePath);
            }
        } catch (IOException e) {
            throw new BallerinaException("Error while reading reference file. " + zipFilePath);
        }
        MapValue<String, Object> refMap = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                CELLERY_PKG_NAME, CELLERY_PKG_VERSION), CelleryConstants.REFERENCE_DEFINITION);
        jsonObject.keys().forEachRemaining(key -> refMap.put(key,
                jsonObject.get(key).toString().replace(INSTANCE_NAME_PLACEHOLDER, "{{" + instanceName + "}}")));
        return refMap;
    }

    private static JSONObject readReferenceJSON(String zipFilePath) throws IOException {
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
