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

import io.fabric8.kubernetes.api.model.ObjectMeta;
import lombok.Data;
import lombok.EqualsAndHashCode;

import static io.cellery.CelleryConstants.CELLERY_RESOURCE_VERSION;

/**
 * Cell POJO model.
 */
@EqualsAndHashCode(callSuper = true)
@Data()
public class Cell extends Composite {
    private CellSpec spec;

    public Cell() {
        super("cell");
    }

    public Cell(ObjectMeta metadata, CellSpec spec) {
        kind = "Cell";
        apiVersion = CELLERY_RESOURCE_VERSION;
        this.metadata = metadata;
        this.spec = spec;
    }
}
