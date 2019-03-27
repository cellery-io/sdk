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
package org.cellery.components.test.utils;

import com.esotericsoftware.yamlbeans.YamlReader;
import io.cellery.models.Cell;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

import static io.cellery.CelleryConstants.CELL_YAML_PATH;

/**
 * Native function cellery:createInstance.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createInstance",
        args = {@Argument(name = "imageName", type = TypeKind.STRING),
                @Argument(name = "imageVersion", type = TypeKind.STRING),
                @Argument(name = "cellImage", type = TypeKind.OBJECT),
                @Argument(name = "instanceName", type = TypeKind.STRING)},
        returnType = {@ReturnType(type = TypeKind.STRING)},
        isPublic = true
)

public class CelleryUtils extends BlockingNativeCallableUnit {
    public Cell getInstance(String destinationPath) {
        Cell cell = null;
        try (InputStreamReader fileReader = new InputStreamReader(new FileInputStream(destinationPath),
                StandardCharsets.UTF_8)) {
            YamlReader reader = new YamlReader(fileReader);
            cell = reader.read(Cell.class);
        } catch (IOException e) {
            throw new BallerinaException("Unable to read Cell image file " + destinationPath + ". \nDid you " +
                    "pull/build" +
                    " the cell image ?");
        }
        if (cell == null) {
            throw new BallerinaException("Unable to extract Cell Image yaml " + CELL_YAML_PATH);
        }
        return cell;
    }

    @Override
    public void execute(Context context) {

    }
}
