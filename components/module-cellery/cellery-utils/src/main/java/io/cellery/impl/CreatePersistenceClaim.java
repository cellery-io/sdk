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
import io.cellery.exception.BallerinaCelleryException;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaim;
import org.ballerinalang.jvm.values.MapValue;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;

import static io.cellery.CelleryUtils.executeShellCommand;
import static io.cellery.CelleryUtils.getVolumeClaim;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * CreatePersistenceClaim implementation.
 */
public class CreatePersistenceClaim {

    public static void createPersistenceClaim(MapValue volumeClaimMap) throws BallerinaCelleryException {
        PersistentVolumeClaim volumeClaim = getVolumeClaim(volumeClaimMap);
        final String targetDirectory = System.getProperty("user.dir") + File.separator + CelleryConstants.TARGET;
        final String targetFile = targetDirectory + File.separator + CelleryConstants.VOLUMES +
                File.separator + volumeClaim.getMetadata().getName() + CelleryConstants.YAML;
        try {
            writeToFile(toYaml(volumeClaim), targetFile);
            String output = executeShellCommand("kubectl create -f " + targetFile, Paths.get(targetDirectory),
                    CelleryUtils::printDebug, CelleryUtils::printWarning);
            if (output.contains("created")) {
                Files.delete(Paths.get(targetFile));
            }
        } catch (IOException e) {
            throw new BallerinaCelleryException("Unable to create volume claim in path " + targetFile);
        } catch (BallerinaException e) {
            throw new BallerinaCelleryException("Unable to deploy volume claim from file " + targetFile);
        }
    }
}
