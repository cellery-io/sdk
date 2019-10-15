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

package org.cellery.models.internal;

import org.cellery.models.API;
import org.cellery.models.GRPC;
import org.cellery.models.Port;
import org.cellery.models.ScalingPolicy;
import org.cellery.models.TCP;
import org.cellery.models.Web;
import io.fabric8.kubernetes.api.model.Probe;
import io.fabric8.kubernetes.api.model.ResourceRequirements;
import lombok.Data;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Cell component.
 */
@Data
public class ImageComponent {
    private String name;
    private int replicas;
    private String gatewayType;
    private String type;
    private Map<String, String> envVars;
    private Map<String, String> labels;
    private List<API> apis;
    private List<TCP> tcpList;
    private List<GRPC> grpcList;
    private Web web;
    private String source;
    private boolean isDockerPushRequired;
    private String service;
    private String protocol;
    private List<Port> ports;
    private int containerPort;
    private ScalingPolicy scalingPolicy;
    private List<String> unsecuredPaths;
    private Probe readinessProbe;
    private Probe livenessProbe;
    private ResourceRequirements resources;
    private List<VolumeInfo> volumes;

    public ImageComponent() {
        ports = new ArrayList<>();
        envVars = new HashMap<>();
        labels = new HashMap<>();
        apis = new ArrayList<>();
        tcpList = new ArrayList<>();
        grpcList = new ArrayList<>();
        unsecuredPaths = new ArrayList<>();
        volumes = new ArrayList<>();
        replicas = 1;
    }

    public void addApi(API api) {
        this.apis.add(api);
    }

    public void addTCP(TCP tcp) {
        this.tcpList.add(tcp);
    }

    public void addGRPC(GRPC grpc) {
        this.grpcList.add(grpc);
    }

    public void addEnv(String key, String value) {
        this.envVars.put(key, value);
    }

    public void addLabel(String key, String value) {
        this.labels.put(key, value);
    }

    public void addUnsecuredPaths(String context) {
        this.unsecuredPaths.add(context);
    }

    public void addPort(Port port) {
        this.ports.add(port);
    }

    public void addVolumeInfo(VolumeInfo volumeInfo) {
        this.volumes.add(volumeInfo);
    }
}
