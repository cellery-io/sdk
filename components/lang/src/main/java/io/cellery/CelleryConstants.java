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
    public static final String CELLERY_PACKAGE = "celleryio/cellery:0.0.0";
    public static final String RECORD_NAME_DEFINITION = "Definition";

    public static final String CELL_REFERENCE_TEMPLATE_FILE = "cell_reference.bal.mustache";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_NAME = "cellName";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_VERSION = "cellVersion";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_GATEWAY_PORT = "cellGatewayPort";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_GATEWAY_PROTOCOL = "cellGatewayProtocol";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_COMPONENTS = "components";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_HANDLE_API_NAME = "handleApiName";
    public static final String CELL_REFERENCE_TEMPLATE_CONTEXT_HANDLE_TYPE_NAME = "handleTypeName";

    public static final String AUTO_SCALING_METRIC_RESOURCE = "Resource";
    public static final String AUTO_SCALING_METRIC_RESOURCE_CPU = "cpu";

    // These should match the Ballerina object names of the Auto Scaling Metrics Objects
    public static final String AUTO_SCALING_METRIC_OBJECT_CPU_UTILIZATION_PERCENTAGE = "CpuUtilizationPercentage";
    public static final String YAML = ".yaml";
    public static final String TARGET = "target";
    public static final String RESOURCES = "resources";
    public static final String DEFAULT_GATEWAY_PROTOCOL = "http";
    public static final int DEFAULT_GATEWAY_PORT = 80;
    public static final String DEFAULT_PARAMETER_VALUE = "";
    public static final String CELLERY_HOME = System.getProperty("user.home") + File.separator +
            ".cellery";
    public static final String CELLERY_REPO_PATH = CELLERY_HOME + File.separator + "repo" + File.separator;
    public static final String CELL_YAML_PATH = "artifacts" + File.separator + "cellery" + File.separator;
}
