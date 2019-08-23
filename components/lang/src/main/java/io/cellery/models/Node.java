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
package io.cellery.models;

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

/**
 * Tree node data model Class.
 *
 * @param <T> data structure
 */
@Data
public class Node<T> {
    private T data;
    private List<Node<T>> children;
    private Node parent;

    public Node(T data) {
        this.data = data;
        this.children = new ArrayList<>();
    }

    /**
     * Add a child node to a given node.
     *
     * @param data data encapsulated by the node
     * @return created child node
     */
    public Node<T> addChild(T data) {
        Node<T> child = new Node<>(data);
        child.setParent(this);
        this.children.add(child);
        return child;
    }
}
