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

package io.cellery.util;

import io.cellery.CelleryUtils;

import java.util.HashMap;
import java.util.Map;

/**
 * Cellery Utility methods.
 */
public class KubernetesClient {
    /**
     * Apply file.
     *
     * @param fileName File name
     */
    public static void apply(String fileName) {
        CelleryUtils.executeShellCommand("kubectl apply -f  " + fileName, null,
                CelleryUtils::printDebug, CelleryUtils::printWarning);
    }

    /**
     * Deletes a Kubernetes resource.
     *
     * @param name Resource name
     * @param type Resource type
     */
    public static void delete(String name, String type) {
        CelleryUtils.executeShellCommand("kubectl delete " + type + " " + name, null,
                CelleryUtils::printDebug, CelleryUtils::printWarning);
    }

    /**
     * Wait for condition.
     *
     * @param condition the condition that is being checked
     * @param timeoutSeconds waiting time for the condition
     * @param resourceName name of the resource
     */
    public static void waitFor(String condition, int timeoutSeconds, String resourceName) {
        CelleryUtils.executeShellCommand("kubectl wait --for=condition=" + condition +
                        " --timeout=" + timeoutSeconds + "s " +
                        resourceName,
                null,
                CelleryUtils::printDebug,
                CelleryUtils::printWarning);
    }

    /**
     * Get kubernetes resource.
     *
     * @param type type of resource
     * @param name name of resource
     * @return resource
     */
    private static String getResource(String type, String name) {
        String expectedException = "not found";
        String output;
        try {
            Map<String, String> environment = new HashMap<>();
            output = CelleryUtils.executeShellCommand(null, msg -> { }, msg -> { }, environment,
                    "kubectl", "get", type, name, "-o", "json");
        } catch (Exception ex) {
            if (!ex.getMessage().contains(expectedException)) {
                throw ex;
            } else {
                output = ex.getMessage();
            }
        }
        return output;
    }

    /**
     * Get cells.
     *
     * @param instance instance name.
     * @return cells
     */
    public static String getCells(String instance) {
        return getResource("cells", instance);
    }

    /**
     * Get composites.
     *
     * @param instance instance name
     * @return composites
     */
    public static String getComposites(String instance) {
        return getResource("composites", instance);
    }
}
