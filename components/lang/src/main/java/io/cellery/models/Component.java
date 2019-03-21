package io.cellery.models;

import lombok.Data;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Cell component.
 */
@Data
public class Component {
    String name;
    int replicas;
    String gatewayType;
    Map<String, String> envVars;
    Map<String, String> labels;
    List<API> apis;
    List<TCP> tcpList;
    List<GRPC> grpcList;
    String source;
    String service;
    String protocol;
    Map<Integer, Integer> containerPortToServicePortMap;
    AutoScaling autoScaling;

    public Component() {
        envVars = new HashMap<>();
        labels = new HashMap<>();
        apis = new ArrayList<>();
        tcpList = new ArrayList<>();
        grpcList = new ArrayList<>();
        containerPortToServicePortMap = new HashMap<>();
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

    public void addPorts(int containerMap, int servicePort) {
        this.containerPortToServicePortMap.put(containerMap, servicePort);
    }

    public void addEnv(String key, String value) {
        this.envVars.put(key, value);
    }

    public void addLabel(String key, String value) {
        this.labels.put(key, value);
    }
}
