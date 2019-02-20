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
import org.ballerinalang.model.values.BRefType;
import org.ballerinalang.model.values.BValueArray;
import org.ballerinalang.natives.annotations.Argument;
import org.ballerinalang.natives.annotations.BallerinaFunction;
import org.ballerinalang.natives.annotations.ReturnType;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.Enumeration;
import java.util.Map;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;

import static io.cellery.CelleryConstants.CELLERY_HOME;
import static io.cellery.CelleryConstants.CELLERY_REPO_PATH;
import static io.cellery.CelleryConstants.CELL_YAML_PATH;
import static io.cellery.CelleryConstants.DEFAULT_PARAMETER_VALUE;
import static io.cellery.CelleryConstants.YAML;
import static io.cellery.CelleryUtils.getValidName;
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
        String cellVersion = ctx.getStringArgument(1);
        String orgName = cellNameData[0];
        String cellName = cellNameData[1];
        String instanceName = ctx.getStringArgument(2);
        String cellImagePath = CELLERY_REPO_PATH + orgName + File.separator + cellName + File.separator
                + cellVersion + File.separator + cellName + ".zip";
        Cell cell = new CreateInstance().getInstance(cellImagePath, cellName);

        final BMap refArgument = (BMap) ctx.getNullableRefArgument(0);
        processComponents(((BValueArray) refArgument.getMap().get("components")).getValues());
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
        //Replace cell name with instance name
        cell.getMetadata().setName(instanceName);
        String destinationPath = CELLERY_HOME + File.separator + "tmp" + File.separator
                + "instances" + File.separator + instanceName + File.separator + instanceName + YAML;
        try {
            writeToFile(removeTags(toYaml(cell)), destinationPath);
        } catch (IOException e) {
            throw new BallerinaException("Unable to persist updated cell file " + destinationPath);
        }
    }

    private Cell getInstance(String cellImagePath, String cellName) {
        Cell cell = null;
        try (ZipFile zipFile = new ZipFile(cellImagePath)) {
            Enumeration<? extends ZipEntry> entries = zipFile.entries();
            while (entries.hasMoreElements()) {
                ZipEntry entry = entries.nextElement();
                if ((CELL_YAML_PATH + cellName + ".yaml").equals(entry.getName())) {
                    InputStream stream = zipFile.getInputStream(entry);
                    YamlReader reader = new YamlReader(new InputStreamReader(stream, StandardCharsets.UTF_8));
                    cell = reader.read(Cell.class);
                    break;
                }
            }
        } catch (IOException e) {
            throw new BallerinaException("Unable to read Cell image file " + cellImagePath + ". \nDid you pull/build" +
                    " the cell image ?");
        }
        if (cell == null) {
            throw new BallerinaException("Unable to extract Cell Image yaml " + CELL_YAML_PATH + cellName);
        }
        return cell;
    }

    private void processComponents(BRefType<?>[] components) {
        for (BRefType componentDefinition : components) {
            if (componentDefinition == null) {
                continue;
            }
            Component component = new Component();
            ((BMap<?, ?>) componentDefinition).getMap().forEach((key, value) -> {
                switch (key.toString()) {
                    case "name":
                        component.setName(value.toString());
                        component.setService(getValidName(value.toString()));
                        break;
                    case "parameters":
                        ((BMap<?, ?>) value).getMap().forEach((k, v) -> {
                            if (((BMap) v).getMap().get("value") != null) {
                                if (!((BMap) v).getMap().get("value").toString().isEmpty()) {
                                    component.addEnv(k.toString(), ((BMap) v).getMap().get("value").toString());
                                }
                            } else {
                                component.addEnv(k.toString(), DEFAULT_PARAMETER_VALUE);
                            }
                        });
                        break;
                    case "labels":
                        ((BMap<?, ?>) value).getMap().forEach((labelKey, labelValue) ->
                                component.addLabel(labelKey.toString(), labelValue.toString()));
                        break;
                    default:
                        break;
                }
            });
            componentHolder.addComponent(component);
        }
    }


    private String removeTags(String string) {
        //a tag is a sequence of characters starting with ! and ending with whitespace
        return removePattern(string, " ![^\\s]*");
    }
}
