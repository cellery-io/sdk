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

import io.cellery.CelleryUtils;
import io.fabric8.kubernetes.api.model.PersistentVolumeClaim;
import io.fabric8.kubernetes.api.model.Secret;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BLangVMErrors;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.LinkedHashMap;

import static io.cellery.CelleryConstants.TARGET;
import static io.cellery.CelleryConstants.VOLUMES;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.executeShellCommand;
import static io.cellery.CelleryUtils.getSecret;
import static io.cellery.CelleryUtils.getVolumeClaim;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;

/**
 * createSecret implementation.
 */
@BallerinaFunction(
        orgName = "celleryio", packageName = "cellery:0.0.0",
        functionName = "createSecret",
        args = {@Argument(name = "secret", type = TypeKind.RECORD)},
        returnType = {@ReturnType(type = TypeKind.ERROR)},
        isPublic = true
)
public class CreateSecret extends BlockingNativeCallableUnit {

    @Override
    public void execute(Context context) {
        LinkedHashMap secretMap = ((BMap) context.getNullableRefArgument(0)).getMap();
        Secret secret = getSecret(secretMap);
        final String targetDirectory = System.getProperty("user.dir") + File.separator + TARGET;
        final String targetFile = targetDirectory + File.separator + VOLUMES +
                File.separator + secret.getMetadata().getName() + YAML;
        try {
            writeToFile(toYaml(secret), targetFile);
            String output = executeShellCommand("kubectl create -f " + targetFile, Paths.get(targetDirectory),
                    CelleryUtils::printDebug, CelleryUtils::printWarning);
            if (output.contains("created")) {
                Files.delete(Paths.get(targetFile));
            }
        } catch (IOException e) {
            context.setError(BLangVMErrors.createError(context, "Unable to create secret in path " + targetFile));
        } catch (BallerinaException e) {
            context.setError(BLangVMErrors.createError(context, "Unable to deploy secret from file " + targetFile));
        }

    }
}
