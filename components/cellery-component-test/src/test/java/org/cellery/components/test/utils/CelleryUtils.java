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
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

/**
 * Native function cellery:createInstance.
 */

public class CelleryUtils {
    private CelleryUtils(){}

    public static Cell getInstance(String destinationPath) {
        Cell cell;
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
            throw new BallerinaException("Unable to extract Cell Image yaml ");
        }
        return cell;
    }
}
