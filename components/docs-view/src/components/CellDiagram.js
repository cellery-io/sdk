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

/* eslint max-lines: ["error", 600] */

import "vis/dist/vis-network.min.css";
import CircularProgress from "@material-ui/core/CircularProgress";
import Constants from "../constants";
import React from "react";
import vis from "vis";
import {withStyles} from "@material-ui/core/styles";
import * as PropTypes from "prop-types";

const styles = {
    progress: {
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        textAlign: "center"
    }
};

class CellDiagram extends React.Component {

    static NodeType = {
        CELL: "cell",
        COMPONENT: "component",
        GATEWAY: "gateway",
        COMPOSITE: "composite"
    };

    static GRAPH_OPTIONS = {
        nodes: {
            size: 40,
            font: {
                size: 15,
                color: "#000000",
                multi: true,
                bold: {
                    color: "#777777",
                    size: 12
                }
            },
            scaling: {
                max: 10000
            }
        },
        edges: {
            width: 2,
            smooth: {
                type: "continuous",
                forceDirection: "none",
                roundness: 0
            },
            color: {
                inherit: false,
                color: "#ccc7c7"
            },

            /*
             * Arrows 'from' configuration will switch the direction of the arrow 'from' and 'to' in edge data.
             *the config will show the arrow label closer to 'to' node
             */
            arrows: {
                from: {
                    enabled: true,
                    scaleFactor: 0.5
                }
            },
            font: {
                align: "horizontal",
                size: 12,
                color: "#777777"
            }
        },
        layout: {
            improvedLayout: false
        },
        autoResize: true,
        physics: {
            enabled: true,
            forceAtlas2Based: {
                gravitationalConstant: -800,
                centralGravity: 0.1,
                avoidOverlap: 1
            },
            solver: "forceAtlas2Based",
            stabilization: {
                enabled: true,
                iterations: 25,
                fit: true
            }
        },
        interaction: {
            selectConnectedEdges: false,
            hover: true
        }
    };

    static CELL_COMPONENT_SEPARATOR = " ";

    constructor(props) {
        super(props);
        this.dependencyGraph = React.createRef();
        this.loader = React.createRef();
    }

    componentDidMount = () => {
        if (this.dependencyGraph.current && this.loader.current) {
            this.draw();
        }
    };

    componentDidUpdate = () => {
        if (this.network) {
            this.network.destroy();
            this.network = null;

            if (this.dependencyGraph.current && this.loader.current) {
                this.draw();
            }
        }
    };

    componentWillUnmount = () => {
        if (this.network) {
            this.network.destroy();
            this.network = null;
        }
    };

    draw = () => {
        const {data, focusedNode, onClickNode} = this.props;

        const componentNodes = [];
        const dataEdges = [];
        const parentNodes = [];
        const availableCells = data.cells;
        const availableComposites = data.composites;
        const focusCellIngressTypes = data.metaInfo[focusedNode].ingresses;
        const componentDependencyLinks = data.metaInfo[focusedNode].componentDependencyLinks;
        const nodeType = data.metaInfo[focusedNode].type;
        const availableComponents = data.components.filter((component) => (focusedNode === component.parent))
            .map((component) => `${component.parent}${CellDiagram.CELL_COMPONENT_SEPARATOR}${component.name}`);

        const getGroupNodesIds = (group) => {
            const output = [];
            nodes.get({
                filter: function(item) {
                    if (item.group === group) {
                        output.push(item.id);
                    }
                }
            });
            return output;
        };

        const getDistance = (pts, centroid) => {
            const cenX = centroid.x;
            const cenY = centroid.y;
            const distance = [];
            for (let i = 0; i < pts.length; i++) {
                distance.push(Math.hypot(pts[i].x - cenX, pts[i].y - cenY));
            }
            const dist = Math.max(...distance);
            return dist;
        };

        const getPolygonCentroid = (pts) => {
            let maxX;
            let maxY;
            let minX;
            let minY;
            for (let i = 0; i < pts.length; i++) {
                minX = (pts[i].x < minX || minX === undefined) ? pts[i].x : minX;
                maxX = (pts[i].x > maxX || maxX === undefined) ? pts[i].x : maxX;
                minY = (pts[i].y < minY || minY === undefined) ? pts[i].y : minY;
                maxY = (pts[i].y > maxY || maxY === undefined) ? pts[i].y : maxY;
            }
            return {x: (minX + maxX) / 2, y: (minY + maxY) / 2};
        };

        const getGroupNodePositions = (groupId) => {
            const groupNodes = getGroupNodesIds(groupId);
            const nodePositions = this.network.getPositions(groupNodes);
            return Object.values(nodePositions);
        };

        const findPoint = (x, y, angle, distance) => {
            const result = {};
            result.x = Math.round(Math.cos(angle * Math.PI / 180) * distance + x);
            result.y = Math.round(Math.sin(angle * Math.PI / 180) * distance + y);
            return result;
        };

        const getIngressTypes = (parent) => data.metaInfo[parent].ingresses.join(", ");

        const getAngleBetweenPoints = (cx, cy, ex, ey) => {
            const dy = ey - cy;
            const dx = ex - cx;
            let theta = Math.atan2(dy, dx);
            theta *= 180 / Math.PI;
            return theta;
        };

        // Add component nodes
        if (availableComponents) {
            availableComponents.forEach((node, index) => {
                componentNodes.push({
                    id: node,
                    label: node.split(CellDiagram.CELL_COMPONENT_SEPARATOR)[1],
                    shape: "image",
                    image: "./component.svg",
                    group: CellDiagram.NodeType.COMPONENT
                });
            });
        }

        // Add cell nodes
        if (availableCells) {
            for (const cell of availableCells) {
                if (cell === focusedNode) {
                    parentNodes.push({
                        id: cell,
                        label: cell,
                        shape: "image",
                        image: "./focusedCell.svg",
                        group: Constants.Type.CELL
                    });
                } else {
                    const cellIngressTypes = data.metaInfo[cell].ingresses;
                    let cellLabelText = "";

                    if (cellIngressTypes.length === 0) {
                        cellLabelText = cell;
                    } else {
                        cellLabelText = `${cell}\n<b>(${getIngressTypes(cell)})</b>`;
                    }

                    parentNodes.push({
                        id: cell,
                        label: cellLabelText,
                        shape: "image",
                        image: "./cell.svg",
                        group: Constants.Type.CELL
                    });
                }
            }
        }

        // Add composite nodes
        if (availableComposites) {
            for (const composite of availableComposites) {
                if (composite === focusedNode) {
                    parentNodes.push({
                        id: composite,
                        label: composite,
                        shape: "image",
                        image: "./focusedComposite.svg",
                        group: Constants.Type.COMPOSITE
                    });
                } else {
                    const compIngressTypes = data.metaInfo[composite].ingresses;
                    let compLabelText = "";

                    if (compIngressTypes.length === 0) {
                        compLabelText = composite;
                    } else {
                        compLabelText = `${composite}\n<b>(${getIngressTypes(composite)})</b>`;
                    }

                    parentNodes.push({
                        id: composite,
                        label: compLabelText,
                        shape: "image",
                        image: "./composite.svg",
                        group: Constants.Type.COMPOSITE
                    });
                }
            }
        }

        // Add edges
        if (data.dependencyLinks) {
            data.dependencyLinks.forEach((edge, index) => {
                // Finding distinct links
                const linkMatches = dataEdges.find(
                    (existingEdge) => (existingEdge.from.parent === edge.from.parent
                        && existingEdge.from.component === edge.from.component) && existingEdge.to === edge.to);

                if (!linkMatches) {
                    if (focusedNode === edge.to) {
                        if (nodeType === Constants.Type.COMPOSITE) {
                            // Add temporary node to border of the composite
                            parentNodes.push({
                                id: `temp-node-${edge.to}`,
                                label: "",
                                shape: "dot",
                                fixed: true,
                                color: {
                                    border: "#fff",
                                    background: "#fff"
                                }
                            });

                            /*
                             * Draw the edge from incoming node to temporary node.
                             * 'from' and 'to' values from edge data is switched and assigned dataEdges according to the
                             *arrow config in GRAPH_OPTIONS.
                             */
                            dataEdges.push({
                                from: `temp-node-${edge.to}`,
                                to: edge.from.parent,
                                label: edge.alias
                            });
                        }

                        // Add the gateway node incoming edges
                        dataEdges.push({
                            from: `${edge.to}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`,
                            to: edge.from.parent,
                            label: edge.alias
                        });
                    } else if (focusedNode === edge.from.parent) {
                        // Add component to cell edges
                        dataEdges.push({
                            from: edge.to,
                            to: `${edge.from.parent}${CellDiagram.CELL_COMPONENT_SEPARATOR}${edge.from.component}`,
                            label: edge.alias
                        });
                    } else {
                        // Add cell/composite dependencies
                        dataEdges.push({
                            from: edge.to,
                            to: edge.from.parent,
                            label: edge.alias
                        });
                    }
                }
            });
        }

        // Add component wise dependencies
        if (componentDependencyLinks) {
            componentDependencyLinks.forEach((edge, index) => {
                // Finding distinct links
                const linkMatches = dataEdges.find(
                    (existingEdge) => existingEdge.from === edge.from && existingEdge.to === edge.to);

                if (!linkMatches) {
                    dataEdges.push({
                        from: `${focusedNode}${CellDiagram.CELL_COMPONENT_SEPARATOR}${edge.to}`,
                        to: `${focusedNode}${CellDiagram.CELL_COMPONENT_SEPARATOR}${edge.from}`
                    });
                }
            });
        }

        const nodes = new vis.DataSet(parentNodes);
        nodes.add(componentNodes);

        const incomingNodes = [];
        if (nodeType === Constants.Type.CELL) {
            // Add gateway node
            nodes.add({
                id: `${focusedNode}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`,
                label: CellDiagram.NodeType.GATEWAY,
                shape: "image",
                image: "./gateway.svg",
                group: CellDiagram.NodeType.GATEWAY,
                fixed: true
            });

            // Check for incoming nodes of a cell to get the position to place the graph on the canvas
            dataEdges.forEach((edge, index) => {
                if (edge.from
                    === `${focusedNode}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`) {
                    incomingNodes.push(edge.to);
                }
            });
        }

        const edges = new vis.DataSet(dataEdges);

        const graphData = {
            nodes: nodes,
            edges: edges
        };

        if (!this.network) {
            this.network = new vis.Network(this.dependencyGraph.current, graphData, CellDiagram.GRAPH_OPTIONS);

            this.loader.current.style.visibility = "visible";
            this.loader.current.style.height = "60vh";
            this.dependencyGraph.current.style.visibility = "hidden";

            const spacing = 150;
            const allNodes = nodes.get({returnType: "Object"});
            const updatedNodes = [];

            this.network.on("stabilized", () => {
                const nodeIds = nodes.getIds();
                const centerPoint = getPolygonCentroid(getGroupNodePositions(CellDiagram.NodeType.COMPONENT));
                const polygonRadius = getDistance(getGroupNodePositions(CellDiagram.NodeType.COMPONENT), centerPoint);
                const size = polygonRadius + spacing;

                // Place temporary node on the border of the composite to get the incoming edge point to the composite
                if (nodeType === Constants.Type.COMPOSITE) {
                    dataEdges.forEach((edge, index) => {
                        if (edge.from === `temp-node-${focusedNode}`) {
                            const fromNode = edge.to;
                            const from = this.network.getPositions([fromNode]);
                            const angleInDegrees = getAngleBetweenPoints(centerPoint.x, centerPoint.y, from[fromNode].x,
                                from[fromNode].y);
                            // Get the coordinate of the connecting point of the composite
                            const incomingPoint = findPoint(centerPoint.x, centerPoint.y, angleInDegrees, size - 30);
                            const tempNode = nodes.get(edge.from);
                            // Place the temporary node on the composite border
                            tempNode.fixed = true;
                            tempNode.x = incomingPoint.x;
                            tempNode.y = incomingPoint.y;
                            tempNode.mass = 1;
                            tempNode.size = 1;
                            updatedNodes.push(tempNode);
                        }
                    });
                    nodes.update(updatedNodes);
                }

                this.network.fit({
                    nodes: nodeIds
                });
                this.loader.current.style.visibility = "hidden";
                this.loader.current.style.height = "0vh";
                this.dependencyGraph.current.style.visibility = "visible";
                this.network.setOptions({physics: false});
                window.onresize = () => {
                    this.network.fit({
                        nodes: nodeIds
                    });
                };
            });

            this.network.on("beforeDrawing", (ctx) => {
                ctx.fillStyle = "#ffffff";
                ctx.fillRect(-ctx.canvas.offsetWidth, -(ctx.canvas.offsetHeight + 20),
                    ctx.canvas.width, ctx.canvas.height);
            });

            this.network.on("afterDrawing", (ctx) => {
                const centerPoint = getPolygonCentroid(getGroupNodePositions(CellDiagram.NodeType.COMPONENT));
                const polygonRadius = getDistance(getGroupNodePositions(CellDiagram.NodeType.COMPONENT), centerPoint);
                const size = polygonRadius + spacing;
                const focusCellLabelPoint = findPoint(centerPoint.x, centerPoint.y, 270, size * Math.cos(180 / 8));
                ctx.font = "bold 1.3rem Arial";
                ctx.textAlign = "center";
                ctx.fillStyle = "#666666";
                ctx.fillText(focusedNode, focusCellLabelPoint.x, focusCellLabelPoint.y + 20);
                ctx.font = "bold 0.8rem Arial";
                ctx.textAlign = "center";
                ctx.fillStyle = "#777777";
                if (focusCellIngressTypes.length > 0) {
                    ctx.fillText(`(${focusCellIngressTypes.join(", ")})`, focusCellLabelPoint.x,
                        focusCellLabelPoint.y + 45);
                }
            });

            this.network.on("stabilizationIterationsDone", () => {
                let centerPoint = getPolygonCentroid(getGroupNodePositions(CellDiagram.NodeType.COMPONENT));
                const polygonRadius = getDistance(getGroupNodePositions(CellDiagram.NodeType.COMPONENT),
                    centerPoint);
                const size = polygonRadius + spacing;

                for (const nodeId in allNodes) {
                    if (allNodes[nodeId].group === CellDiagram.NodeType.COMPONENT) {
                        allNodes[nodeId].fixed = true;

                        if (incomingNodes.length > 0) {
                            const positions = this.network.getPositions(allNodes[nodeId].id);
                            if (componentNodes.length === 1) {
                                this.network.moveNode(allNodes[nodeId].id, positions[allNodes[nodeId].id].x, 100);
                            } else if (centerPoint.y > -100 && centerPoint.y < 50) {
                                this.network.moveNode(allNodes[nodeId].id,
                                    positions[allNodes[nodeId].id].x, positions[allNodes[nodeId].id].y + 100);
                            } else {
                                this.network.moveNode(allNodes[nodeId].id,
                                    positions[allNodes[nodeId].id].x, positions[allNodes[nodeId].id].y + 200);
                            }
                        }
                        if (allNodes.hasOwnProperty(nodeId)) {
                            updatedNodes.push(allNodes[nodeId]);
                        }
                    }
                }
                centerPoint = getPolygonCentroid(getGroupNodePositions(CellDiagram.NodeType.COMPONENT));

                const focused = nodes.get(focusedNode);
                focused.size = size;
                focused.label = undefined;
                focused.fixed = true;
                if (componentNodes.length === 1) {
                    focused.mass = 5;
                } else {
                    focused.mass = polygonRadius / 10;
                }
                this.network.moveNode(focusedNode, centerPoint.x, centerPoint.y);
                updatedNodes.push(focused);

                // Placing gateway node
                if (nodeType === Constants.Type.CELL) {
                    const gatewayNode = nodes.get(
                        `${focusedNode}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`);
                    const gatewayPoint = findPoint(centerPoint.x, centerPoint.y, 90, size * Math.cos(180 / 8));

                    if (gatewayNode) {
                        const x = gatewayPoint.x;
                        const y = gatewayPoint.y;
                        this.network.moveNode(gatewayNode.id, x, y);
                        updatedNodes.push(gatewayNode);
                    }
                }
                nodes.update(updatedNodes);
            });

            this.network.on("selectNode", (event) => {
                this.network.unselectAll();
                const clickedNode = nodes.get(event.nodes[0]);
                if (((clickedNode.group === Constants.Type.CELL)
                        || (clickedNode.group === Constants.Type.COMPOSITE)) && clickedNode.id !== focusedNode) {
                    onClickNode(event.nodes[0]);
                }
            });

            this.network.on("dragging", (event) => {
                this.network.unselectAll();
            });
        }
    };

    render = () => {
        const {classes} = this.props;

        return (
            <div className={classes.root}>
                <div ref={this.loader} style={{visibility: "visible", height: "60vh"}} className={classes.progress}>
                    <CircularProgress/>
                </div>
                <div style={{visibility: "hidden", height: "93vh"}} ref={this.dependencyGraph}/>
            </div>);
    };

}

CellDiagram.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.object,
    focusedNode: PropTypes.string.isRequired,
    onClickNode: PropTypes.func
};

export default withStyles(styles)(CellDiagram);
