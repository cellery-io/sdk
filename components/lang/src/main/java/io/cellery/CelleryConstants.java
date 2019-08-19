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
package io.cellery;

import java.io.File;

/**
 * Collected constants of Cellery.
 */
public class CelleryConstants {
    public static final String CELLERY_RESOURCE_VERSION = "mesh.cellery.io/v1alpha1";
    public static final String CELLERY_PACKAGE = "celleryio/cellery:0.0.0";
    public static final String RESOURCE_DEFINITION = "ResourceDefinition";
    public static final String API_DEFINITION = "ApiDefinition";
    public static final String REFERENCE_DEFINITION = "Reference";
    public static final String IMAGE_NAME_DEFINITION = "ImageName";
    public static final String CELLERY_REPO_PATH =
            System.getProperty("user.home") + File.separator + ".cellery" + File.separator + "repo";

    public static final String IMAGE_SOURCE = "source";
    public static final String AUTO_SCALING_METRIC_RESOURCE = "Resource";
    public static final String AUTO_SCALING_METRIC_RESOURCE_CPU = "cpu";
    public static final String AUTO_SCALING_METRIC_RESOURCE_MEMORY = "memory";
    public static final String ORG = "org";
    public static final String NAME = "name";
    public static final String VERSION = "ver";
    public static final String INSTANCE_NAME = "instanceName";
    public static final String BTYPE_STRING = "string";
    public static final String MAX_REPLICAS = "maxReplicas";
    public static final String CONCURRENCY_TARGET = "concurrencyTarget";

    //Dependency Constants
    public static final String DEPENDENCIES = "dependencies";
    public static final String CELLS = "cells";
    public static final String CELL = "Cell";
    public static final String COMPOSITES = "composites";
    public static final String COMPOSITE = "Composite";
    public static final String COMPONENTS = "components";

    // Ballerina object name constants for Resources
    public static final String LIMITS = "limits";
    public static final String REQUESTS = "requests";

    // Ballerina object name constants for probes
    public static final String INGRESSES = "ingresses";
    public static final String LABELS = "labels";
    public static final String SCALING_POLICY = "scalingPolicy";
    public static final String ENV_VARS = "envVars";
    public static final String PROBES = "probes";
    public static final String POD_RESOURCES = "resources";
    public static final String KIND = "kind";
    public static final String LIVENESS = "liveness";
    public static final String READINESS = "readiness";
    public static final String EXPOSE = "expose";

    // These should match the Ballerina object names of the Auto Scaling Metrics Objects
    public static final String ENVOY_GATEWAY = "Envoy";
    public static final String MICRO_GATEWAY = "MicroGateway";
    public static final String PROTO_FILE = "protoFile";
    public static final String YAML = ".yaml";

    //Ingress Constants
    public static final String PROTOCOL_TCP = "TCP";
    public static final String GATEWAY_PORT = "gatewayPort";
    public static final String PROTOCOL_GRPC = "GRPC";
    public static final String TARGET = "target";
    public static final String RESOURCES = "resources";
    public static final String DEFAULT_GATEWAY_PROTOCOL = "http";
    public static final int DEFAULT_GATEWAY_PORT = 80;
    public static final String DEFAULT_PARAMETER_VALUE = "";
    public static final String CELLERY_IMAGE_DIR_ENV_VAR = "CELLERY_IMAGE_DIR";
    public static final String GATEWAY_SERVICE = "--gateway-service";
    public static final String INSTANCE_NAME_PLACEHOLDER = "{{instance_name}}";
    public static final String REFERENCE_FILE_NAME = "reference.json";
    public static final String METADATA_FILE_NAME = "metadata.json";

    public static final String ANNOTATION_CELL_IMAGE_ORG = "mesh.cellery.io/cell-image-org";
    public static final String ANNOTATION_CELL_IMAGE_NAME = "mesh.cellery.io/cell-image-name";
    public static final String ANNOTATION_CELL_IMAGE_VERSION = "mesh.cellery.io/cell-image-version";
    public static final String ANNOTATION_CELL_IMAGE_DEPENDENCIES = "mesh.cellery.io/cell-dependencies";

    public static final String SERVICE_TYPE_JOB = "Job";
}
