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
    private String name;
    private int replicas;
    private String gatewayType;
    private Map<String, String> envVars;
    private Map<String, String> labels;
    private List<API> apis;
    private List<TCP> tcpList;
    private List<GRPC> grpcList;
    private List<Web> webList;
    private String source;
    private String service;
    private String protocol;
    private int containerPort;
    private AutoScaling autoScaling;

    public Component() {
        envVars = new HashMap<>();
        labels = new HashMap<>();
        apis = new ArrayList<>();
        tcpList = new ArrayList<>();
        grpcList = new ArrayList<>();
        webList = new ArrayList<>();
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

    public void addWeb(Web webIngress) {
        this.webList.add(webIngress);
    }

    public void addEnv(String key, String value) {
        this.envVars.put(key, value);
    }

    public void addLabel(String key, String value) {
        this.labels.put(key, value);
    }
}
