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
        const {data, focusedCell, onClickNode} = this.props;
        const componentNodes = [];
        const dataEdges = [];
        const cellNodes = [];
        const availableCells = data.cells;
        const focusCellIngressTypes = data.metaInfo[focusedCell].ingresses;
        const componentDependencyLinks = data.metaInfo[focusedCell].componentDependencyLinks;
        const availableComponents = data.components.filter((component) => (focusedCell === component.cell))
            .map((component) => `${component.cell}${CellDiagram.CELL_COMPONENT_SEPARATOR}${component.name}`);

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

        const getIngressTypes = (cell) => data.metaInfo[cell].ingresses.join(", ");

        const drawOnCanvas = (from, to, ctx, radius) => {
            const arrow = {
                h: 3,
                w: 10
            };
            const ptCircleFrom = getPointOnCircle(radius, from, to);
            const ptCircleTo = getPointOnCircle(radius, to, from);
            const ptArrow = getPointOnCircle(radius + arrow.w + 10, to, from);

            drawCircle(ctx, from, radius);
            drawCircle(ctx, to, radius);
            drawLine(ctx, ptCircleFrom, ptCircleTo);
            drawArrow(ctx, arrow, ptArrow, ptCircleTo);
        };

        const drawArrow = (ctx, arrow, ptArrow, endPt) => {
            const angleInDegrees = getAngleBetweenPoints(ptArrow, endPt);
            ctx.save();
            ctx.translate(ptArrow.x, ptArrow.y);
            ctx.rotate(angleInDegrees * Math.PI / 180);
            ctx.beginPath();
            ctx.moveTo(0, 0);
            ctx.lineTo(0, -arrow.h);
            ctx.lineTo(arrow.w, 0);
            ctx.lineTo(0, Number(arrow.h));
            ctx.closePath();
            ctx.fillStyle = "#ccc7c7";
            ctx.stroke();
            ctx.fill();
            ctx.restore();
        };

        const drawCircle = (ctx, circle, radius) => {
            ctx.beginPath();
            ctx.fillStyle = "#808080";
            ctx.strokeStyle = "transparent";
            ctx.arc(circle.x, circle.y, radius, 0, 2 * Math.PI, false);
            ctx.stroke();
            ctx.closePath();
        };

        const drawLine = (ctx, startPt, endPt) => {
            ctx.beginPath();
            ctx.moveTo(startPt.x, startPt.y);
            ctx.lineTo(endPt.x, endPt.y);
            ctx.strokeStyle = "#ccc7c7";
            ctx.stroke();
            ctx.closePath();
        };

        const getPointOnCircle = (radius, originPt, endPt) => {
            const angleInDegrees = getAngleBetweenPoints(originPt, endPt);
            const x = radius * Math.cos(angleInDegrees * Math.PI / 180) + originPt.x;
            const y = radius * Math.sin(angleInDegrees * Math.PI / 180) + originPt.y;
            return {x: x, y: y};
        };

        const getAngleBetweenPoints = (originPt, endPt) => {
            const interPt = {
                x: endPt.x - originPt.x,
                y: endPt.y - originPt.y
            };
            return Math.atan2(interPt.y, interPt.x) * 180 / Math.PI;
        };

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
                        label: `${cell}\n<b>(${getIngressTypes(cell)})</b>`,
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
                        from: edge.to,
                        to: edge.from,
                        label: edge.alias
                    });
                }
            });
        }

        const nodes = new vis.DataSet(cellNodes);
        nodes.add(componentNodes);
        nodes.add({
            id: `${focusedCell}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`,
            label: CellDiagram.NodeType.GATEWAY,
            shape: "image",
            image: "./gateway.svg",
            group: CellDiagram.NodeType.GATEWAY,
            fixed: true
        });

        const incomingNodes = [];
        dataEdges.forEach((edge, index) => {
            if (edge.from === focusedCell) {
                edge.from = `${focusedCell}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`;
                incomingNodes.push(edge.from);
            }
        });
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
                ctx.fillText(focusedCell, focusCellLabelPoint.x, focusCellLabelPoint.y + 20);
                ctx.font = "bold 0.8rem Arial";
                ctx.textAlign = "center";
                ctx.fillStyle = "#777777";
                ctx.fillText(`(${focusCellIngressTypes.join(", ")})`, focusCellLabelPoint.x,
                    focusCellLabelPoint.y + 45);

                componentDependencyLinks.forEach((link) => {
                    const from = this.network.getPositions([link.from]);
                    const to = this.network.getPositions([link.to]);
                    if (from[link.from] && to[link.to]) {
                        drawOnCanvas(from[link.from], to[link.to], ctx, 37);
                    }
                });
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

                const focusedNode = nodes.get(focusedCell);
                focusedNode.size = size;
                focusedNode.label = undefined;
                focusedNode.fixed = true;
                if (componentNodes.length === 1) {
                    focusedNode.mass = 5;
                } else {
                    focusedNode.mass = polygonRadius / 10;
                }
                this.network.moveNode(focusedCell, centerPoint.x, centerPoint.y);
                updatedNodes.push(focusedNode);

                // Placing gateway node
                const gatewayNode = nodes.get(
                    `${focusedCell}${CellDiagram.CELL_COMPONENT_SEPARATOR}${CellDiagram.NodeType.GATEWAY}`);
                const gatewayPoint = findPoint(centerPoint.x, centerPoint.y, 90, size * Math.cos(180 / 8));

                if (gatewayNode) {
                    const x = gatewayPoint.x;
                    const y = gatewayPoint.y;
                    this.network.moveNode(gatewayNode.id, x, y);
                    updatedNodes.push(gatewayNode);
                }
                nodes.update(updatedNodes);
            });

            this.network.on("selectNode", (event) => {
                this.network.unselectAll();
                const clickedNode = nodes.get(event.nodes[0]);
                if (clickedNode.group === CellDiagram.NodeType.CELL && clickedNode.id !== focusedCell) {
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
    focusedCell: PropTypes.string,
    onClickNode: PropTypes.func
};

export default withStyles(styles)(CellDiagram);
