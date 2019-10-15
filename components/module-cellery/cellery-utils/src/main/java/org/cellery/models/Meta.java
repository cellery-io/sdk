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
package org.cellery.models;

import lombok.Data;

import java.util.HashMap;
import java.util.Map;

/**
 * Cell Meta data model Class.
 */
@Data
public class Meta {
    private String org;
    private String name;
    private String ver;
    private String instanceName;
    private String alias;
    private String kind;
    private boolean isRunning;
    private boolean shared;
    private Map<String, ComponentMeta> components;
    private Map<String, String> environmentVariables;

    public Meta() {
        components = new HashMap<>();
        environmentVariables = new HashMap<>();
        alias = "";
    }

    /**
     *  Get dependent cells of a given cell.
     *
     * @return Map of dependent cells
     */
    public Map<String, Meta> getDependencies() {
        Map<String, Meta> dependencies = new HashMap<>();
        // Iterate the map of components
        for (Map.Entry<String, ComponentMeta> component : components.entrySet()) {
            // Iterate cell dependencies
            for (Map.Entry<String, Meta> dependentCell : component.getValue().getDependencies().getCells()
                    .entrySet()) {
                dependencies.put(dependentCell.getKey(), dependentCell.getValue());
            }
            // Iterate composite dependencies
            for (Map.Entry<String, Meta> dependentComponent : component.getValue().getDependencies().getComposites()
                    .entrySet()) {
                dependencies.put(dependentComponent.getKey(), dependentComponent.getValue());
            }
        }
        return dependencies;
    }
}
