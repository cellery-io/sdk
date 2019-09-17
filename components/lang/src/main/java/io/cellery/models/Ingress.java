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

package io.cellery.models;

import lombok.Data;
import org.apache.commons.lang3.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Ingress POJO.
 */
@Data
public class Ingress {
    private Extension extensions;
    private List<API> http;
    private List<GRPC> grpc;
    private List<TCP> tcp;

    public Ingress() {
        http = new ArrayList<>();
        grpc = new ArrayList<>();
        tcp = new ArrayList<>();
    }

    public void addHttpAPI(List<API> apis) {
        http.addAll(apis.stream().filter(a -> StringUtils.isNotEmpty(a.getContext())).collect(Collectors.toList()));
    }

    public void addTCP(List<TCP> tcpIngress) {
        tcp.addAll(tcpIngress);
    }

    public void addGRPC(List<GRPC> grpcIngress) {
        grpc.addAll(grpcIngress);
    }
}
