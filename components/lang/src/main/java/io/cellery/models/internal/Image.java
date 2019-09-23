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
 */
package io.cellery.models.internal;

import io.cellery.models.Test;
import lombok.Data;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Cell/Composite Image model Class.
 */
@Data
public class Image {
    private Map<String, ImageComponent> componentNameToComponentMap;
    private String orgName;
    private String cellName;
    private String cellVersion;
    private List<Dependency> dependencies;
    private boolean zeroScaling;
    private boolean autoScaling;
    private boolean compositeImage;
    private Test test;

    public Image() {
        componentNameToComponentMap = new HashMap<>();
        dependencies = new ArrayList<>();
    }

    public Map<String, ImageComponent> getComponentNameToComponentMap() {
        return componentNameToComponentMap;
    }

    public void addComponent(ImageComponent component) {
        if (componentNameToComponentMap.containsKey(component.getName())) {
            throw new BallerinaException("Two components with same name exists " + component.getName());
        }
        this.componentNameToComponentMap.put(component.getName(), component);
    }

    public void addDependency(Dependency dependency) {
        dependencies.add(dependency);
    }
}
