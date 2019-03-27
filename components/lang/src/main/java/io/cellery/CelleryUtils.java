/*
 * Copyright (c) 2018, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
 */
package io.cellery;

import com.esotericsoftware.yamlbeans.YamlWriter;
import io.cellery.models.Component;
import org.apache.commons.io.FileUtils;
import org.ballerinalang.model.values.BMap;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.io.File;
import java.io.IOException;
import java.io.StringWriter;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.LinkedHashMap;
import java.util.Locale;

import static io.cellery.CelleryConstants.DEFAULT_PARAMETER_VALUE;
import static io.cellery.CelleryConstants.RESOURCES;
import static io.cellery.CelleryConstants.TARGET;

/**
 * Cellery Utility methods.
 */
public class CelleryUtils {

    /**
     * Returns swagger file as a String.
     *
     * @param path     swagger file path
     * @param encoding string encoding
     * @return swagger file as a String
     * @throws IOException if unable to read file
     */
    public static String readSwaggerFile(String path, Charset encoding) throws IOException {
        byte[] encoded = Files.readAllBytes(Paths.get(path));
        return new String(encoded, encoding);
    }

    /**
     * Returns valid kubernetes name.
     *
     * @param name actual value
     * @return valid name
     */
    public static String getValidName(String name) {
        return name.toLowerCase(Locale.getDefault()).replaceAll("\\P{Alnum}", "-");
    }


    public static void processParameters(Component component, LinkedHashMap<?, ?> parameters) {
        parameters.forEach((k, v) -> {
            if (((BMap) v).getMap().get("value") != null) {
                if (!((BMap) v).getMap().get("value").toString().isEmpty()) {
                    component.addEnv(k.toString(), ((BMap) v).getMap().get("value").toString());
                }
            } else {
                component.addEnv(k.toString(), DEFAULT_PARAMETER_VALUE);
            }
            //TODO:Handle secrets
        });
    }

    /**
     * Write content to a File. Create the required directories if they don't not exists.
     *
     * @param context    context of the file
     * @param targetPath target file path
     * @throws IOException If an error occurs when writing to a file
     */
    public static void writeToFile(String context, String targetPath) throws IOException {
        File newFile = new File(targetPath);
        // delete if file exists
        if (newFile.exists()) {
            Files.delete(Paths.get(newFile.getPath()));
        }
        //create required directories
        if (newFile.getParentFile().mkdirs()) {
            Files.write(Paths.get(targetPath), context.getBytes(StandardCharsets.UTF_8));
            return;
        }
        Files.write(Paths.get(targetPath), context.getBytes(StandardCharsets.UTF_8));
    }

    /**
     * Generates Yaml from a object.
     *
     * @param object Object
     * @param <T>    Any Object type
     * @return Yaml as a string.
     */
    public static <T> String toYaml(T object) {
        try (StringWriter stringWriter = new StringWriter()) {
            YamlWriter writer = new YamlWriter(stringWriter);
            writer.write(object);
            writer.getConfig().writeConfig.setWriteRootTags(false); //replaces only root tag
            writer.close(); //don't add this to finally, because the text will not be flushed
            return stringWriter.toString();
        } catch (IOException e) {
            throw new BallerinaException("Error occurred while generating yaml definition." + e.getMessage());
        }
    }

    /**
     * Copy file target/resources directory.
     *
     * @param sourcePath source file/directory path
     */
    public static void copyResourceToTarget(String sourcePath) {
        File src = new File(sourcePath);
        String targetPath = TARGET + File.separator + RESOURCES + File.separator + src.getName();
        File dst = new File(targetPath);
        // if source is file
        try {
            if (Files.isRegularFile(Paths.get(sourcePath))) {
                if (Files.isDirectory(dst.toPath())) {
                    // if destination is directory
                    FileUtils.copyFileToDirectory(src, dst);
                } else {
                    // if destination is file
                    FileUtils.copyFile(src, dst);
                }
            } else if (Files.isDirectory(Paths.get(sourcePath))) {
                FileUtils.copyDirectory(src, dst);
            }

        } catch (IOException e) {
            throw new BallerinaException("Error occured while copying resource file " + sourcePath +
                    ". " + e.getMessage());
        }
    }
}
