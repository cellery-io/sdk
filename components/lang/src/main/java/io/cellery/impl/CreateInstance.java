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

import com.esotericsoftware.yamlbeans.YamlReader;
import io.cellery.models.Cell;
import io.cellery.models.Component;
import io.cellery.models.ComponentHolder;
import org.ballerinalang.bre.Context;
import org.ballerinalang.bre.bvm.BlockingNativeCallableUnit;
import org.ballerinalang.model.types.TypeKind;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.model.values.BString;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.LinkedHashMap;
import java.util.Map;

import static io.cellery.CelleryConstants.CELLERY_HOME;
import static io.cellery.CelleryConstants.CELL_YAML_PATH;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.processParameters;
import static io.cellery.CelleryUtils.toYaml;
import static io.cellery.CelleryUtils.writeToFile;
import static org.apache.commons.lang3.StringUtils.removePattern;

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
public class CreateInstance extends BlockingNativeCallableUnit {
    private ComponentHolder componentHolder = new ComponentHolder();

    public void execute(Context ctx) {
        String[] cellNameData = ctx.getStringArgument(0).split("/");
        String cellName = cellNameData[1];
        String destinationPath = CELLERY_HOME + File.separator + "tmp" + File.separator + cellName + File.separator +
                "artifacts" + File.separator + "cellery" + File.separator + cellName + YAML;
        Cell cell = getInstance(destinationPath);
        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        processComponents((BMap) refArgument.getMap().get("components"));
        cell.getSpec().getServicesTemplates().forEach(serviceTemplate -> {
            String componentName = serviceTemplate.getMetadata().getName();
            Map<String, String> updatedParams =
                    componentHolder.getComponentNameToComponentMap().get(componentName).getEnvVars();
            //Replace parameter values defined in the YAML.
            serviceTemplate.getSpec().getContainer().getEnv().forEach(envVar -> {
                if (updatedParams.containsKey(envVar.getName()) && !updatedParams.get(envVar.getName()).isEmpty()) {
                    envVar.setValue(updatedParams.get(envVar.getName()));
                }
            });

        });
        try {
            writeToFile(removeTags(toYaml(cell)), destinationPath);
        } catch (IOException e) {
            throw new BallerinaException("Unable to persist updated cell yaml " + destinationPath);
        }
    }

    private Cell getInstance(String destinationPath) {
        Cell cell = null;
        try (InputStreamReader fileReader = new InputStreamReader(new FileInputStream(destinationPath),
                StandardCharsets.UTF_8)) {
            YamlReader reader = new YamlReader(fileReader);
            cell = reader.read(Cell.class);
        } catch (IOException e) {
            throw new BallerinaException("Unable to read Cell image file " + destinationPath + ". \nDid you " +
                    "pull/build the cell image ?");
        }
        if (cell == null) {
            throw new BallerinaException("Unable to extract Cell Image yaml " + CELL_YAML_PATH);
        }
        return cell;
    }

    private void processComponents(BMap<?, ?> components) {
        components.getMap().forEach((key, value) -> {
            LinkedHashMap componentValues = ((BMap) value).getMap();
            Component component = new Component();
            // Mandatory fields
            component.setName(((BString) componentValues.get("name")).stringValue());
            if (componentValues.containsKey("parameters")) {
                processParameters(component, ((BMap<?, ?>) componentValues.get("parameters")).getMap());
            }
            componentHolder.addComponent(component);
        });
    }


    private String removeTags(String string) {
        //a tag is a sequence of characters starting with ! and ending with whitespace
        return removePattern(string, " ![^\\s]*");
    }
}
