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

package org.cellery.components.test.utils;

import java.io.File;

/**
 * Constants used in Cellery Lang Testing.
 */
public class CelleryTestConstants {
    public static final String CELLERY_REPO_PATH =
            System.getProperty("user.home") + File.separator + ".cellery" + File.separator + "repo";
    public static final String CELLERY = "cellery";
    public static final String TARGET = "target";
    public static final String ARTIFACTS = "artifacts";
    public static final String METADATA = "metadata.json";
    public static final String EMPLOYEE_PORTAL = "employee-portal";
    public static final String HELLO_WEB = "hello-web";
    public static final String PET_CARE_STORE = "pet-care-store";
    public static final String PET_SERVICE = "pet-service";
    public static final String TLS_WEB = "tls-web";
    public static final String DOCKER_SOURCE = "docker-source";
    public static final String PRODUCT_REVIEW = "product-review";
    public static final String CELLERY_IMAGE_ORG = "mesh.cellery.io/cell-image-org";
    public static final String CELLERY_IMAGE_NAME = "mesh.cellery.io/cell-image-name";
    public static final String CELLERY_IMAGE_VERSION = "mesh.cellery.io/cell-image-version";
    public static final String CELLERY_MESH_VERSION = "mesh.cellery.io/v1alpha1";
    public static final String YAML = ".yaml";
    public static final String BAL = ".bal";
    public static final String JSON = ".json";
}
