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

import "vis/dist/vis-network.min.css";
import CircularProgress from "@material-ui/core/CircularProgress";
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
        GATEWAY: "gateway"
    };

    static GRAPH_OPTIONS = {
        nodes: {
            shapeProperties: {
                borderRadius: 10
            },
            borderWidth: 1,
            size: 40,
            font: {
                size: 15,
                color: "#000000"
            },
            scaling: {
                max: 10000
            },
            chosen: false
        },
        edges: {
            width: 2,
            smooth: false,
            color: {
                inherit: false,
                color: "#ccc7c7"
            },
            arrows: {
                to: {
                    enabled: true,
                    scaleFactor: 0.5
                }
            }, scaling: {
                max: 10000
            }
        },
        layout: {
            improvedLayout: false
        },
        autoResize: true,
        physics: {
            enabled: true,
            forceAtlas2Based: {
                gravitationalConstant: -500,
                centralGravity: 0.075,
                avoidOverlap: 1
            },
            solver: "forceAtlas2Based",
            stabilization: {
                enabled: true,
                iterations: 10,
                fit: true
            }
        },
        interaction: {
            selectConnectedEdges: false
        }
    };

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
        const {data, focusedCell, onClickNode} = this.props;
        const componentNodes = [];
        const dataEdges = [];
        const cellNodes = [];
        const availableCells = data.cells;
        const availableComponents = data.components.filter((component) => (focusedCell === component.cell))
            .map((component) => component);

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

        if (availableComponents) {
            availableComponents.forEach((node, index) => {
                componentNodes.push({
                    id: node.name,
                    label: node.name,
                    shape: "image",
                    image: "./component.svg",
                    group: CellDiagram.NodeType.COMPONENT
                });
            });
        }

        if (availableCells) {
            for (const cell of availableCells) {
                if (cell === focusedCell) {
                    cellNodes.push({
                        id: cell,
                        label: cell,
                        shape: "image",
                        image: "./focusedCell.svg",
                        group: CellDiagram.NodeType.CELL
                    });
                } else {
                    cellNodes.push({
                        id: cell,
                        label: cell,
                        shape: "image",
                        image: "./cell.svg",
                        group: CellDiagram.NodeType.CELL
                    });
                }
            }
        }

        if (data.dependencyLinks) {
            data.dependencyLinks.forEach((edge, index) => {
                // Finding distinct links
                const linkMatches = dataEdges.find(
                    (existingEdge) => existingEdge.from === edge.from && existingEdge.to === edge.to);

                if (!linkMatches) {
                    dataEdges.push({
                        id: index,
                        from: edge.from,
                        to: edge.to
                    });
                }
            });
        }

        const nodes = new vis.DataSet(cellNodes);
        const edges = new vis.DataSet(dataEdges);
        nodes.add(componentNodes);

        const graphData = {
            nodes: nodes,
            edges: edges
        };

        if (!this.network) {
            this.network = new vis.Network(this.dependencyGraph.current, graphData, CellDiagram.GRAPH_OPTIONS);

            const spacing = 150;
            const allNodes = nodes.get({returnType: "Object"});
            const updatedNodes = [];

            this.network.on("stabilized", () => {
                const nodeIds = nodes.getIds();

                this.network.fit({
                    nodes: nodeIds
                });
                this.loader.current.style.visibility = "hidden";
                this.loader.current.style.height = "0vh";
                this.dependencyGraph.current.style.visibility = "visible";

                window.onresize = () => {
                    this.network.fit({
                        nodes: nodeIds
                    });
                };
            });

            this.network.on("stabilizationIterationsDone", () => {
                const centerPoint = getPolygonCentroid(getGroupNodePositions(CellDiagram.NodeType.COMPONENT));
                const polygonRadius = getDistance(getGroupNodePositions(CellDiagram.NodeType.COMPONENT),
                    centerPoint);
                const size = polygonRadius + spacing;

                const focusedNode = nodes.get(focusedCell);
                focusedNode.size = size;
                focusedNode.fixed = true;
                focusedNode.interaction = {
                    selectable: false,
                    draggable: false
                };
                focusedNode.mass = polygonRadius / 10;
                this.network.moveNode(focusedCell, centerPoint.x, centerPoint.y);

                for (const nodeId in allNodes) {
                    if (allNodes[nodeId].group === CellDiagram.NodeType.COMPONENT) {
                        allNodes[nodeId].fixed = true;
                        if (allNodes.hasOwnProperty(nodeId)) {
                            updatedNodes.push(allNodes[nodeId]);
                        }
                    }
                }

                updatedNodes.push(focusedNode);
                nodes.update(updatedNodes);
            });

            this.network.on("click", (event) => {
                onClickNode(event.nodes[0]);
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
                <div style={{visibility: "hidden", height: "100vh"}} ref={this.dependencyGraph}/>
            </div>);
    };

}

CellDiagram.propTypes = {
    classes: PropTypes.object.isRequired,
    data: PropTypes.arrayOf(PropTypes.object),
    focusedCell: PropTypes.arrayOf(PropTypes.object),
    onClickNode: PropTypes.func,
};

export default withStyles(styles)(CellDiagram);
