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

package org.cellery.impl;

import org.cellery.CelleryUtils;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaim;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;
import org.cellery.CelleryConstants;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.LinkedHashMap;

import static org.cellery.CelleryUtils.executeShellCommand;
import static org.cellery.CelleryUtils.getVolumeClaim;
import static org.cellery.CelleryUtils.toYaml;
import static org.cellery.CelleryUtils.writeToFile;

/**
 * createPersistenceClaim implementation.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createPersistenceClaim",
        args = {@Argument(name = "volumeClaim", type = TypeKind.RECORD)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class CreatePersistenceClaim extends BlockingNativeCallableUnit {

    @Override
    public void execute(Context context) {
        LinkedHashMap volumeClaimMap = ((BMap) context.getNullableRefArgument(0)).getMap();
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
            context.setError(BLangVMErrors.createError(context, "Unable to create volume claim in path " + targetFile));
        } catch (BallerinaException e) {
            context.setError(BLangVMErrors.createError(context, "Unable to deploy volume claim from file "
                    + targetFile));
        }
    }
}
