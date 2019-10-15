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

package io.cellery.models;

import com.fasterxml.jackson.annotation.JsonInclude;
import io.fabric8.kubernetes.api.model.ConfigMap;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.Secret;
import lombok.Data;

import java.util.ArrayList;
import java.util.List;

/**
 * Component Spec.
 */
@Data
public class ComponentSpec {
    private String type;

    @JsonInclude(JsonInclude.Include.NON_NULL)
    private ScalingPolicy scalingPolicy;
    private PodSpec template;
    private List<Port> ports;
    private List<ConfigMap> configurations;
    private List<Secret> secrets;
    private List<VolumeClaim> volumeClaims;

    public ComponentSpec() {
        configurations = new ArrayList<>();
        secrets = new ArrayList<>();
        volumeClaims = new ArrayList<>();
        ports = new ArrayList<>();
    }

    public void addConfiguration(ConfigMap configMap) {
        this.configurations.add(configMap);
    }

    public void addSecrets(Secret secret) {
        this.secrets.add(secret);
    }

    public void addVolumeClaim(VolumeClaim volumeClaim) {
        this.volumeClaims.add(volumeClaim);
    }
}
