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

package io.cellery.models;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.Singular;

import java.util.ArrayList;
import java.util.List;

/**
 * Gateway Spec.
 */
@Data
@AllArgsConstructor
public class GatewaySpec {
    String type;
    String hostName;
    String tlsSecretName;
    private @Singular("http")
    List<API> http;
    private @Singular("tcp")
    List<TCP> tcp;
    private @Singular("grpc")
    List<GRPC> grpc;

    public GatewaySpec() {
        http = new ArrayList<>();
        tcp = new ArrayList<>();
        grpc = new ArrayList<>();
    }

    public void addHttpAPI(List<API> api) {
        http.addAll(api);
    }

    public void addTCP(List<TCP> tcpIngress) {
        tcp.addAll(tcpIngress);
    }

    public void addGRPC(List<GRPC> grpcIngress) {
        grpc.addAll(grpcIngress);
    }

}
