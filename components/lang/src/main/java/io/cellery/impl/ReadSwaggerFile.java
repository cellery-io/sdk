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

import io.cellery.CelleryConstants;
import io.swagger.models.Swagger;
import io.swagger.parser.SwaggerParser;
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
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.IOException;
import java.nio.charset.Charset;
import java.util.concurrent.atomic.AtomicLong;

import static io.cellery.CelleryUtils.copyResourceToTarget;
import static io.cellery.CelleryUtils.readSwaggerFile;

/**
 * Class to Parse Swagger.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "readSwaggerFile",
        args = {@Argument(name = "swaggerFilePath", type = TypeKind.STRING)},
        returnType = {@ReturnType(type = TypeKind.ARRAY), @ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class ReadSwaggerFile extends BlockingNativeCallableUnit {

    public void execute(Context ctx) {
        BArrayType bArrayType =
                new BArrayType(ctx.getProgramFile().getPackageInfo(CelleryConstants.CELLERY_PACKAGE).getTypeDefInfo(
                        CelleryConstants.RECORD_NAME_DEFINITION).typeInfo.getType());
        BValueArray bValueArray = new BValueArray(bArrayType);
        String swaggerFilePath = ctx.getNullableStringArgument(0);
        final String specification;
        try {
            specification = readSwaggerFile(swaggerFilePath, Charset.defaultCharset());
            copyResourceToTarget(swaggerFilePath);
        } catch (IOException e) {
            throw new BallerinaException("Unable to read swagger file. " + swaggerFilePath);
        }
        final Swagger swagger = new SwaggerParser().parse(specification);
        String basePath = swagger.getBasePath();
        if (basePath.isEmpty() || "/".equals(basePath)) {
            basePath = "";
        }
        AtomicLong runCount = new AtomicLong(0L);
        String finalBasePath = basePath;
        swagger.getPaths().forEach((path, pathDefinition) ->
                pathDefinition.getOperationMap().forEach((httpMethod, operation) -> {
                    BMap<String, BValue> bmap = BLangConnectorSPIUtil.createBStruct(ctx,
                            CelleryConstants.CELLERY_PACKAGE,
                            CelleryConstants.RECORD_NAME_DEFINITION,
                            finalBasePath + path, httpMethod.toString());
                    bValueArray.add(runCount.getAndIncrement(), bmap);

                }));
        ctx.setReturnValues(bValueArray);
    }
}
