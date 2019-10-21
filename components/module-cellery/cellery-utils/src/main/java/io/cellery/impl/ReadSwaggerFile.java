/*
 *   Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  WSO2 Inc. licenses this file to you under the Apache License,
 *  Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package io.cellery.impl;

import io.cellery.exception.BallerinaCelleryException;
import io.swagger.models.Swagger;
import io.swagger.parser.SwaggerParser;
import org.ballerinalang.jvm.BallerinaErrors;
import org.ballerinalang.jvm.BallerinaValues;
import org.ballerinalang.jvm.types.BArrayType;
import org.ballerinalang.jvm.types.BPackage;
import org.ballerinalang.jvm.values.ArrayValue;
import org.ballerinalang.jvm.values.MapValue;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryConstants.API_DEFINITION;
import static io.cellery.CelleryConstants.CELLERY_PKG_NAME;
import static io.cellery.CelleryConstants.CELLERY_PKG_ORG;
import static io.cellery.CelleryConstants.CELLERY_PKG_VERSION;
import static io.cellery.CelleryConstants.RESOURCE_DEFINITION;
import static io.cellery.CelleryUtils.copyResourceToTarget;

/**
 * Class to Parse Swagger.
 */
public class ReadSwaggerFile {

    public static MapValue readSwaggerFileInternal(String swaggerFilePath) {
        String specification = "";
        try {
            specification = readSwagger(swaggerFilePath, Charset.defaultCharset());
            copyResourceToTarget(swaggerFilePath);
        } catch (IOException e) {
            throw BallerinaErrors.createInteropError(new BallerinaCelleryException(e.getMessage()));
        }
        final Swagger swagger = new SwaggerParser().parse(specification);
        String basePath = swagger.getBasePath();
        if (basePath.isEmpty() || "/".equals(basePath)) {
            basePath = "";
        }
        AtomicLong runCount = new AtomicLong(0L);
        String finalBasePath = basePath;
        ArrayValue arrayValue = new ArrayValue(new BArrayType(BallerinaValues.createRecordValue(new BPackage(
                CELLERY_PKG_ORG, CELLERY_PKG_NAME, CELLERY_PKG_VERSION), RESOURCE_DEFINITION).getType()));
        MapValue<String, Object> apiDefinitions = BallerinaValues.createRecordValue(new BPackage(CELLERY_PKG_ORG,
                CELLERY_PKG_NAME, CELLERY_PKG_VERSION), API_DEFINITION);
        swagger.getPaths().forEach((path, pathDefinition) ->
                pathDefinition.getOperationMap().forEach((httpMethod, operation) -> {
                    MapValue<String, Object> resource = BallerinaValues.createRecordValue(new BPackage("celleryio",
                            "cellery", "0.5.0"), "ResourceDefinition");
                    resource.put("path", finalBasePath + path);
                    resource.put("method", httpMethod.toString());
                    arrayValue.add(runCount.getAndIncrement(), resource);
                }));
        apiDefinitions.put("resources", arrayValue);
        return apiDefinitions;
    }

    /**
     * Returns swagger file as a String.
     *
     * @param path     swagger file path
     * @param encoding string encoding
     * @return swagger file as a String
     * @throws IOException if unable to read file
     */
    private static String readSwagger(String path, Charset encoding) throws IOException {
        byte[] encoded = Files.readAllBytes(Paths.get(path));
        return new String(encoded, encoding);
    }

}
