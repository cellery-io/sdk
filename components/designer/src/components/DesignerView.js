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

import "./designer.css";
import "vis-network/dist/vis-network.min.css";
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';
import Grid from '@material-ui/core/Grid';
import Tooltip from "@material-ui/core/Tooltip/Tooltip";
import Typography from "@material-ui/core/Typography";
import React from "react";
import classNames from "classnames";
import ArtTrackRounded from "@material-ui/icons/ArtTrackRounded";
import Delete from "@material-ui/icons/DeleteOutlineRounded";
import vis from "vis-network";
import {withStyles} from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import {
    ZoomInRounded,
    ZoomOutRounded,
    ZoomOutMapOutlined,
    ArrowRightAlt,
    ArrowBackRounded
} from '@material-ui/icons';
import * as PropTypes from "prop-types";
import Divider from "@material-ui/core/Divider";
import Drawer from "@material-ui/core/Drawer";
import DesignerHeader from "./DesignerHeader";
import * as axios from "axios";
import Snackbar from '@material-ui/core/Snackbar';
import IconButton from '@material-ui/core/IconButton';
import CloseIcon from '@material-ui/icons/Close';

const drawerWidth = 242;

const styles = (theme) => ({
    root: {
        display: "flex"
    },
    graph: {
        height: "93vh",
        "& .vis-network": {
            outline: "none"
        }
    },
    toolbar: {
        paddingTop: 10,
        paddingLeft: 25,
        paddingBottom: 10,
        paddingRight: 25,
        borderBottom: "1px solid #eee",
        borderTop: "1px solid #eee"
    },
    toolIcon: {
        color: "#808080"
    },
    instructions: {
        display: "inline-block",
        fontWeight: 600
    },
    diagramTools: {
        marginLeft: 20
    },
    hide: {
        display: "none"
    },
    drawer: {
        width: drawerWidth,
        flexShrink: 0,
        position: "sticky"
    },
    drawerPaper: {
        width: drawerWidth,
        borderTopWidth: 1,
        position: "inherit"
    },
    drawerHeader: {
        display: "flex",
        alignItems: "center",
        padding: 5,
        justifyContent: "flex-start",

        minHeight: "fit-content"
    },
    content: {
        flexGrow: 1,
        transition: theme.transitions.create("margin", {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen
        }),
        marginLeft: Number(theme.spacing.unit),
        marginRight: -drawerWidth + theme.spacing.unit
    },
    contentShift: {
        transition: theme.transitions.create("margin", {
            easing: theme.transitions.easing.easeOut,
            duration: theme.transitions.duration.enteringScreen
        }),
        marginRight: theme.spacing.unit
    },
    sideBarHeading: {
        fontSize: 10,
        fontWeight: 500,
        marginLeft: 4,
        letterSpacing: 1,
        textTransform: "uppercase"
    },
    btnLegend: {
        float: "right",
        position: "sticky",
        bottom: 20,
        marginTop: 10,
        fontSize: 12
    },
    legendContent: {
        padding: theme.spacing.unit * 2
    },
    legendText: {
        display: "inline-flex",
        marginLeft: 5,
        fontSize: 12
    },
    legendIcon: {
        verticalAlign: "middle",
        marginLeft: 20
    },
    legendFirstEl: {
        verticalAlign: "middle"
    },
    graphContainer: {
        display: "flex",
        width: "100%"
    },
    diagram: {
        flexGrow: 1,
        borderLeft: "1px solid #0000003b"

    },
    elements: {
        width: 33
    },
    elementBtn: {
        minWidth: 0
    },
    inputFields: {
        '& .MuiTextField-root': {
            margin: theme.spacing(1),
            width: 200,
        },
    },
    noContentMsg: {
        fontSize: 12,
        padding: 20
    },
    back: {
        padding: 0,
        marginRight: 10
    },
    actions: {
        display: "inline-block"
    },
    toolGrid: {
        display: "contents"
    },
    toolBtn: {
        padding: "5.5px 6px"
    }
});

class DesignerView extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            isAddClicked: false,
            isNodeSelected: false,
            isEdgeSelected: false,
            helpText: "",
            offsetX: 0,
            offsetY: 0,
            scale: 1.3,
            positionX: 0,
            positionY: 0,
            open: true,
            openSnackBar: false,
            errorContent: "",
            nodeType: 'none',
            jsonFile: {}
        };
        this.graph = {
            nodes: [],
            edges: []
        };
        this.nodesData = {};
        this.edgesData = {};
        this.dependencyGraph = React.createRef();

        this.fileReader = new FileReader();
        this.fileReader.onload = event => {
            this.setState({jsonFile: JSON.parse(event.target.result)}, () => {
                this.graph.nodes = [];
                this.graph.edges = [];
                this.buildNetworkData(this.state.jsonFile);
            });
        };
    }

    componentDidMount = () => {
        if (this.dependencyGraph.current) {
            this.draw();
        }
    };

    componentWillUnmount = () => {
        if (this.network) {
            this.network.destroy();
            this.network = null;
        }
    };

    handleDrawerOpen = () => {
        this.setState({open: true});
    };

    handleDrawerClose = () => {
        this.setState({open: false});
    };

    handleSnackBarClose = (event, reason) => {
        if (reason === 'clickaway') {
            return;
        }

        this.setState({
            openSnackBar: false
        });
    };


    static options = {
        physics: false,
        layout: {
            randomSeed: 10000
        },
        autoResize: true,
        nodes: {
            scaling: {
                min: 1,
                label: {
                    enabled: false,
                    min: 1,
                },
                customScalingFunction: function (min, max, total, value) {
                    return value;

                }
            },
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
                to: {
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
        interaction: {
            multiselect: false,
        },
        manipulation: false
    };

    draw = () => {
        let {nodes, edges} = this.graph;

        this.nodesData = new vis.DataSet(nodes);
        this.edgesData = new vis.DataSet(edges);

        const data = {
            nodes: this.nodesData,
            edges: this.edgesData
        };

        this.network = new vis.Network(this.dependencyGraph.current, data, DesignerView.options);

        this.network.on('dragEnd', (params) => {
            this.graph.nodes.forEach((item) => {
                let position = this.network.getPositions(item.id);
                item.x = position[item.id].x;
                item.y = position[item.id].y;
            });
            this.nodesData.update(nodes);
        });

        this.network.on('selectNode', (params) => {

            this.setState({
                isNodeSelected: true,
                isEdgeSelected: false,
            });

            const selectedNode = this.nodesData.get(params.nodes[0]);
            if (selectedNode.type === "cell" || selectedNode.type === "composite") {
                this.setState({
                    nodeType: "cell"
                });
            } else if (selectedNode.type === "component") {
                this.setState({
                    nodeType: "component"
                });
            } else if (selectedNode.type === "gateway") {
                this.setState({
                    nodeType: "gateway"
                });
            }
        });

        this.network.on('dragStart', (params) => {
            this.setState({
                isNodeSelected: true,
                isEdgeSelected: false,
            });

            const selectedNode = this.nodesData.get(params.nodes[0]);
            if (selectedNode.type === "cell" || selectedNode.type === "composite") {
                this.setState({
                    nodeType: "cell"
                });
            } else if (selectedNode.type === "component") {
                this.setState({
                    nodeType: "component"
                });
            } else if (selectedNode.type === "gateway") {
                this.setState({
                    nodeType: "gateway"
                });
            }
        });

        this.network.on('deselectNode', (params) => {
            this.setState({
                isNodeSelected: false,
                nodeType: "none"
            });
        });

        this.network.on('selectEdge', (params) => {

            this.getEdges();

            if (this.network.getSelectedNodes().length >= 1) {
                this.setState({
                    isEdgeSelected: false,
                });
            } else {
                this.setState({
                    isEdgeSelected: true,
                    nodeType: "link"
                });
            }

        });

        this.network.on('deselectEdge', (params) => {
            this.setState({
                isEdgeSelected: false,
                nodeType: "none"
            });
        });
    };

    viewGenerator = (nodeType) => {
        let instanceView;

        if (nodeType === "cell") {
            instanceView =
                '<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlnsXlink="http://www.w3.org/1999/xlink" x="0px" y="0px"\n' +
                '\t width="14px" height="14px" viewBox="0 0 14 14" style="enable-background:new 0 0 14 14" xml:space="preserve">\n' +
                '\t<path fill="none"  stroke="#808080" stroke-width="0.2px"\n' +
                '\t\t  d="M9,0.8H5C4.7,0.8,4.3,1,4,1.3L1.3,4C1,4.3,0.8,4.6,0.8,5v4c0,0.4,0.2,0.7,0.4,1L4,12.8c0.3,0.3,0.6,0.4,1,0.4H9c0.4,0,0.7-0.1,1-0.4l2.8-2.8c0.3-0.3,0.4-0.6,0.4-1V5c0-0.4-0.2-0.7-0.4-1L10,1.3C9.7,1,9.3,0.8,9,0.8z"/>\n' +
                '</svg>';
        } else if (nodeType === "composite") {
            instanceView =
                '<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlnsXlink="http://www.w3.org/1999/xlink" x="0px" y="0px"\n' +
                '   viewBox="0 0 14 14" style="enable-background:new 0 0 14 14;" xml:space="preserve">\n' +
                '<style type="text/css">\n' +
                '\t.st0{fill:none;}\n' +
                '\t.st1{fill:none;stroke:#808080;stroke-width:0.2;stroke-dasharray:1.0231,1.0231;}\n' +
                '</style>\n' +
                '<g>\n' +
                '\t<path class="st0" d="M13.2,7c0,3.4-2.8,6.2-6.2,6.2S0.8,10.4,0.8,7S3.6,0.8,7,0.8S13.2,3.6,13.2,7z"/>\n' +
                '\t<g>\n' +
                '\t\t<circle class="st1" cx="7" cy="7" r="6.2"/>\n' +
                '\t</g>\n' +
                '</g>\n' +
                '</svg>';
        } else if (nodeType === "gateway") {
            instanceView =
                '<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlnsXlink="http://www.w3.org/1999/xlink" x="0px" y="0px"\n' +
                '     width="14px" height="14px" viewBox="0 0 13 13" style="enable-background:new 0 0 13 13" xml:space="preserve">\n' +
                '    <path fill="#fff" stroke="#808080" stroke-width="0.2px"\n' +
                '          d="M13,7a6,6,0,0,1-6,6.06A6,6,0,0,1,1,7,6,6,0,0,1,7,.94,6,6,0,0,1,13,7Z" transform="translate(-0.59 -0.49)"/>\n' +
                '    <path fill="#808080" stroke="#fff" stroke-width="0.1px"\n' +
                '          d="M6.39,6.29v1L4.6,5.47a.14.14,0,0,1,0-.21L6.39,3.47v1a.15.15,0,0,0,.15.14H9.3a.15.15,0,0,1,.14.15V6a.15.15,0,0,1-.14.15H6.54A.15.15,0,0,0,6.39,6.29ZM7.46,7.85H4.7A.15.15,0,0,0,4.56,8V9.27a.15.15,0,0,0,.14.15H7.46a.15.15,0,0,1,.15.14v1L9.4,8.74a.14.14,0,0,0,0-.21L7.61,6.74v1A.15.15,0,0,1,7.46,7.85Z" transform="translate(13.5 -0.49) rotate(90)"/>\n' +
                '</svg>';

        } else if (nodeType === "component") {
            instanceView =
                '<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"\n' +
                '\t\t  width="14px" height="14px" viewBox="0 0 13 13" style="enable-background:new 0 0 13 13" xml:space="preserve">\n' +
                '    <path fill="#fff" fill-opacity="1" stroke="#808080" stroke-width="0.2px"\n' +
                '          d="M13 7a6 6 0 0 1-6 6.06A6 6 0 0 1 1 7 6 6 0 0 1 7 .94 6 6 0 0 1 13 7Z" transform="translate(-0.59 -0.49)"/>\n' +
                '    <path fill="#808080" stroke="#fff" stroke-width="0.1px"\n' +
                '\t\t  d="M4.37,5c-.19.11-.19.28,0,.39L6.76,6.82a.76.76,0,0,0,.69,0L9.64,5.45a.23.23,0,0,0,0-.42L7.45,3.7a.76.76,0,0,0-.69,0Z" transform="translate(-0.59 -0.49)"/>\n' +
                '    <path fill="#808080" stroke="#fff" stroke-width="0.1px"\n' +
                '\t\t  d="M10,5.93c0-.22-.15-.31-.34-.19L7.45,7.1a.73.73,0,0,1-.69,0L4.37,5.73c-.19-.11-.35,0-.35.2V8a.88.88,0,0,0,.33.63l2.43,1.68a.61.61,0,0,0,.65,0L9.66,8.63A.9.9,0,0,0,10,8Z" transform="translate(-0.59 -0.49)"/>\n' +
                '    <text fill="#fff" font-size="1.63px" font-family="ArialMT, Arial" transform="translate(5.96 5.1) scale(0.98 1)">μ\n' +
                '\t</text>\'\n' +
                '</svg>';
        }
        return "data:image/svg+xml;charset=utf-8," + encodeURIComponent(instanceView);
    };

    onClickNode = (nodeType) => {
        let {nodes} = this.graph;

        this.setState({
            isAddClicked: true,
            isNodeSelected: false,
            isEdgeSelected: false,
            helpText: "Click on empty space to place the element",
            nodeType: "none"
        });

        this.network.off('click');
        this.network.on('click', (params) => {
            const maxId = nodes.reduce((max, node) => node.iterator > max ? node.iterator : max, 0);
            let nodeLabel = '';

            if (nodeType === "component" || nodeType === "gateway") {
                if (nodeType === "component") {
                    nodeLabel = "new-" + nodeType + "-" + maxId;
                } else {
                    nodeLabel = "gateway";
                }
                nodes.push({
                    label: nodeLabel,
                    id: (maxId + 1) + "-node",
                    x: params.pointer.canvas.x,
                    y: params.pointer.canvas.y,
                    shape: "image",
                    value: 1,
                    iterator: maxId + 1,
                    type: nodeType,
                    image: this.viewGenerator(nodeType),
                    parent: '',
                    name: nodeLabel,
                    sourceImage: ''
                });
            } else if ((nodeType === "cell" || nodeType === "composite")) {
                nodeLabel = "new-" + nodeType + "-" + maxId;
                nodes.unshift({
                    label: "my-org/" + nodeLabel + ":latest",
                    id: (maxId + 1) + "-node",
                    x: params.pointer.canvas.x,
                    y: params.pointer.canvas.y,
                    shape: "image",
                    iterator: maxId + 1,
                    value: 3,
                    type: nodeType,
                    image: this.viewGenerator(nodeType),
                    org: "my-org",
                    version: "latest",
                    name: nodeLabel
                });
            }

            this.nodesData.remove(this.getGroupNodesIds());
            this.nodesData.update(nodes);

            this.setState({
                isAddClicked: false,
                helpText: "",
                nodeType: "none"
            });

            this.network.off('click');

        });
    };

    onClickEdge = () => {

        this.network.unselectAll();

        this.network.on('select', (params) => {
            this.network.disableEditMode();
        });

        this.network.addEdgeMode();

        this.setState({
            isAddClicked: true,
            helpText: "Click on a node and drag the link to another node to connect them",
            isNodeSelected: false,
            nodeType: "none"
        });

        this.network.on('release', (params) => {
            this.setState({
                isAddClicked: false,
                helpText: ""
            });
        });
    };

    getEdges = () => {
        let {edges} = this.graph;
        const allEdges = this.edgesData.get({returnType: 'Object'});
        const connections = [];
        for (const edgeId in allEdges) {
            const edge = edges.find((item) => item.id === edgeId);
            connections.push({
                id: allEdges[edgeId].id,
                to: allEdges[edgeId].to,
                from: allEdges[edgeId].from,
                label: edge ? edge.label : "",
            })
        }
        this.graph.edges = connections;
        this.edgesData.update(edges);
    };

    fitToScreen = () => {
        this.network.fit();
    };

    updateNode = (nodeId, value, attr) => {
        let {nodes} = this.graph;

        nodes.forEach((item) => {
            if (item.id === nodeId) {
                if (attr === "value") {
                    item.value = value;
                } else if (attr === "label") {
                    item.label = value;
                } else if (attr === "name") {
                    item.name = value;
                    if(item.type === "cell" || item.type === "composite"){
                        item.label = item.org + "/" + value + ":" + item.version;
                    } else {
                        item.label = value;
                    }
                } else if (attr === "parent") {
                    item.parent = value;
                } else if (attr === "org") {
                    item.org = value;
                    item.label = value + "/" + item.name + ":" + item.version;
                } else if (attr === "version") {
                    item.version = value;
                    item.label = item.org + "/" + item.name + ":" + value;
                } else if (attr === "sourceImage") {
                    item.sourceImage = value;
                }
            }
        });

        this.nodesData.update(nodes);
    };

    updateEdge = (edgeId, value, attr) => {
        let {edges} = this.graph;

        edges.forEach((item) => {
            if (item.id === edgeId) {
                if (attr === "alias") {
                    item.label = value;
                }
            }
        });
        this.edgesData.update(edges);
    };

    increaseSize = () => {
        const selectedNodes = this.network.getSelectedNodes();

        for (let node of selectedNodes) {
            const resizedNode = [];
            const selectedNode = this.nodesData.get(node);
            selectedNode.value = selectedNode.value + 1;

            this.updateNode(selectedNode.id, selectedNode.value, "value");
            resizedNode.push(selectedNode);
            this.nodesData.update(resizedNode);
        }
    };

    decreaseSize = () => {
        const selectedNodes = this.network.getSelectedNodes();

        for (let node of selectedNodes) {
            const resizedNode = [];
            const selectedNode = this.nodesData.get(node);
            if (selectedNode.value > 1) {
                selectedNode.value = selectedNode.value - 1;
            }
            this.updateNode(selectedNode.id, selectedNode.value, "value");
            resizedNode.push(selectedNode);
            this.nodesData.update(resizedNode);
        }
    };

    zoomIn = () => {
        let {scale, positionX, positionY, offsetX, offsetY} = this.state;
        this.setState({
            scale: scale + 0.3
        });

        const options = {
            position: {x: positionX, y: positionY},
            scale: scale,
            offset: {x: offsetX, y: offsetY},
            animation: true
        };
        this.network.moveTo(options);
    };

    zoomOut = () => {
        let {scale, positionX, positionY, offsetX, offsetY} = this.state;

        if (scale > 0.3) {
            this.setState({
                scale: scale - 0.3
            });
        }

        const options = {
            position: {x: positionX, y: positionY},
            scale: scale,
            offset: {x: offsetX, y: offsetY},
            animation: true
        };
        this.network.moveTo(options);
    };

    getGroupNodesIds = () => {
        const output = [];
        this.nodesData.get({
            filter: function (item) {
                output.push(item.id);

            }
        });
        return output;
    };

    getNodeData = (data) => {
        let networkNodes = [];

        data.forEach((node) => {
            if (node) {
                networkNodes.push({
                    label: node.label,
                    id: node.id,
                    x: node.x,
                    y: node.y,
                    shape: node.shape,
                    iterator: node.iterator,
                    value: node.value,
                    type: node.type,
                    image: node.image,
                    org: node.org,
                    name: node.name,
                    version: node.version
                });

                if (node.components) {
                    node.components.forEach((childNode) => {
                        networkNodes.push(childNode);
                    });
                }
                if (Object.keys(node.gateway).length > 0) {
                    networkNodes.push(node.gateway);
                }
            }
        });
        return networkNodes;
    };

    getTypeNodesIds = (type) => {
        const output = [];
        this.nodesData.get({
            filter: function (item) {
                if (item.type === type) {
                    output.push(item.id);
                }

            }
        });
        return output;
    };

    getParentOfComponent = () => {
        const cellNodeIds = this.getTypeNodesIds("cell");
        const compositeNodeIds = this.getTypeNodesIds("composite");
        const parentNodes = cellNodeIds.concat(compositeNodeIds);
        const componentNodeIds = this.getTypeNodesIds("component");
        const gatewayNodeIds = this.getTypeNodesIds("gateway");
        const childNodes = componentNodeIds.concat(gatewayNodeIds);

        parentNodes.forEach((item) => {
            const bb = this.network.getBoundingBox(item);
            const selectedNode = this.nodesData.get(item);

            childNodes.forEach((node) => {
                const nodeData = this.nodesData.get(node);
                let xValue = this.network.getPositions(node)[node].x;
                let yValue = this.network.getPositions(node)[node].y;
                if (xValue > bb.left && xValue < bb.right && yValue > bb.top && yValue < bb.bottom) {
                    nodeData.parent = selectedNode.name;
                    this.updateNode(node, selectedNode.name, "parent");
                    this.nodesData.update(nodeData);
                }
            });
        });
    };

    getChangeHandler = (field) => (event) => {
        const selectedNode = this.network.getSelectedNodes();
        const selectedEdge = this.network.getSelectedEdges();

        if (field === "organization") {
            this.updateNode(selectedNode[0], event.target.value, "org");
        } else if (field === "label") {
            this.updateNode(selectedNode[0], event.target.value, "label");
        } else if (field === "name") {
            this.updateNode(selectedNode[0], event.target.value, "name");
        } else if (field === "version") {
            this.updateNode(selectedNode[0], event.target.value, "version");
        } else if (field === "sourceImage") {
            this.updateNode(selectedNode[0], event.target.value, "sourceImage");
        } else if (field === "alias") {
            this.updateEdge(selectedEdge[0], event.target.value, "alias");
        }
    };

    getDefaultNodeValue = (field) => {
        const selectedNode = this.network.getSelectedNodes();
        const selectedNodeData = this.nodesData.get(selectedNode[0]);

        if (field === "organization") {
            return selectedNodeData.org;
        } else if (field === "label") {
            return selectedNodeData.label;
        } else if (field === "version") {
            return selectedNodeData.version;
        } else if (field === "name") {
            return selectedNodeData.name;
        } else if (field === "sourceImage") {
            return selectedNodeData.sourceImage;
        }
    };

    getDefaultEdgeValue = () => {
        let {edges} = this.graph;
        const selectedEdge = this.network.getSelectedEdges()[0];
        const edgeLabel = edges.find((item) => item.id === selectedEdge).label
        return edgeLabel;
    };

    deleteNode = () => {
        let {nodes} = this.graph;
        const selectedNode = this.network.getSelectedNodes();
        this.network.deleteSelected();
        this.setState({
            isNodeSelected: false,
            nodeType: "none"
        });

        if (selectedNode.length > 0) {
            const filteredNodes = [];
            nodes.forEach((node, index) => {
                selectedNode.forEach((selected) => {
                    if (node.id === selected) {
                        nodes.splice(index, 1);
                    } else {
                        filteredNodes.push(node)
                    }
                })
            });

            this.nodesData.update(nodes);
        }

    };

    deleteEdge = () => {
        let {edges} = this.graph;
        const selectedEdges = this.network.getSelectedEdges();
        this.network.deleteSelected();
        this.setState({
            isEdgeSelected: false,
            nodeType: "none"
        });

        if (selectedEdges.length > 0) {
            const filteredEdges = [];
            edges.forEach((edge, index) => {
                selectedEdges.forEach((selected) => {
                    if (edge.id === selected) {
                        edges.splice(index, 1);
                    } else {
                        filteredEdges.push(edge)
                    }
                })
            });

            this.edgesData.update(edges);
        }
    };

    getNodeDataStructure = () => {
        let {nodes} = this.graph;

        this.getParentOfComponent();

        const parentNodes = nodes.filter((node) => (node.type === "cell" || node.type === "composite"));
        const childNodes = nodes.filter((node) => (node.type === "component"));
        const gatewayNodes = nodes.filter((node) => (node.type === "gateway"));

        childNodes.forEach((node, index) => {
            node.dependencies = {};
            this.graph.edges.forEach((edge, index) => {
                if (node.id === edge.from) {
                    node.dependencies[edge.label] = edge.to;
                }
            });
        });

        parentNodes.forEach((node, index) => {
            node.components = [];
            childNodes.forEach((childNode, index) => {
                if (childNode.parent === node.name) {
                    node.components.push(childNode)
                }
            });

            if (node.type === "cell") {
                const gatewayNode = gatewayNodes.filter((gateway) => (gateway.parent === node.name))[0];
                node.gateway = gatewayNode ? gatewayNode : {};
            } else {
                node.gateway = {};
            }
        });

        return parentNodes;
    };

    buildDataStructure = () => {
        this.getEdges();

        const graphData = {
            data: {
                path: "",
                nodes: this.getNodeDataStructure(),
                edges: this.graph.edges
            }
        };

        return graphData;
    };

    download = (filename, text) => {
        let element = document.createElement('a');
        element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
        element.setAttribute('download', filename);
        element.style.display = 'none';
        document.body.appendChild(element);
        element.click();
        document.body.removeChild(element);
    };

    buildNetworkData = (response) => {
        if (response) {
            this.graph.nodes = this.getNodeData(response.data.nodes);
            this.graph.edges = response.data.edges;
        }

        this.nodesData.update(this.graph.nodes);
        this.edgesData.update(this.graph.edges);
    };

    saveData = () => {
        const data = JSON.stringify(this.buildDataStructure());
        this.download("cellery-diagram.json", data);
    };

    getNodeFromId = (id) => {
        let output = {};
        this.nodesData.get({
            filter: function (item) {
                if (item.id === id) {
                    output = item;
                }
            }
        });
        return output;
    };

    getNodeParentFromId = (id) => {
        let parent = "";
        const networkDataStructure = this.buildDataStructure();

        networkDataStructure.data.nodes.forEach((node) => {
            if (node.id === id) {
                parent = node.name;
            } else if (node.gateway.id === id) {
                parent = node.name;
            } else {
                node.components.forEach((comp) => {
                    if (comp.id === id) {
                        parent = comp.parent
                    }
                });
            }
        });
        return parent;
    };

    validateGraph = () => {
        const errorMsgs = [];

        if (this.graph.nodes.length > 0) {
            this.graph.nodes.forEach((node) => {
                if ((node.type === "component" && node.parent === "") ||
                    (node.type === "gateway" && node.parent === "")) {
                    errorMsgs.push(node.name + " is not placed in a cell or composite");
                }
            });

            const networkDataStructure = this.buildDataStructure();
            networkDataStructure.data.nodes.forEach((node) => {
                if (node.type === "cell" && Object.keys(node.gateway).length === 0) {
                    errorMsgs.push(node.label + " does not have a gateway node properly placed inside the cell");
                }
            });

            networkDataStructure.data.edges.forEach((edge) => {
                if (((this.getNodeParentFromId(edge.from) !== this.getNodeParentFromId(edge.to)) && edge.label ===
                        "")) {
                    errorMsgs.push("Add Alias for the link from " + this.getNodeFromId(edge.from).label + " to " +
                        this.getNodeFromId(edge.to).label);
                }
            });

        } else {
            errorMsgs.push("No cell or composite data to generate code")
        }

        if (errorMsgs.length === 0) {
            this.generateCode();
        } else {
            this.setState({
                errorContent: errorMsgs[0],
                openSnackBar: true
            });
        }
    };

    generateCode = () => {

        const data = this.buildDataStructure();

        axios({
            method: 'post',
            url: 'http://127.0.0.1:8088/api/generate',
            data: data
        }).then(function (response) {
            let element = document.createElement('a');
            element.setAttribute('href', 'http://127.0.0.1:8088/api/download');
            element.setAttribute('download', "generated_code.zip");
            element.style.display = 'none';
            document.body.appendChild(element);
            element.click();
            document.body.removeChild(element);
        });
    };

    onBackClick = () => {

        this.setState({
            isAddClicked: false
        });
    };


    render() {
        const {classes} = this.props;
        const {isAddClicked, open, isNodeSelected, isEdgeSelected, helpText, nodeType, openSnackBar, errorContent} = this.state;

        return (
            <div>
                <DesignerHeader saveData={this.saveData} fileReader={this.fileReader}
                                validateGraph={this.validateGraph}/>
                <div className={classes.toolbar}>
                    <Grid container spacing={3}>
                        <Grid item xs={12} md={6}>
                            <Grid container spacing={1} direction="row" justify="flex-start" alignItems="center">
                                <Grid item>
                                    {
                                        isAddClicked
                                            ? (<div>
                                                <Typography variant="subtitle2" className={classes.instructions}>
                                                    {
                                                        helpText ?
                                                            (<span><IconButton onClick={this.onBackClick}
                                                                               className={classes.back}>
                                                                <ArrowBackRounded fontSize="small"/>
                                                            </IconButton>{helpText}</span>)
                                                            : null
                                                    }
                                                </Typography>
                                            </div>)
                                            : null
                                    }
                                </Grid>
                            </Grid>
                        </Grid>
                        <Grid item xs={12} md={6}>
                            <Grid container spacing={1} direction="row" justify="flex-end" alignItems="center">
                                <Grid item className={classes.toolGrid}>
                                    {
                                        isNodeSelected && (nodeType === "cell" || nodeType === "composite" ||
                                            nodeType === "gateway" || nodeType === "component")?
                                            (<ButtonGroup size="small" aria-label="small outlined button group"
                                                className={classes.actions}>
                                                <Tooltip title="Increase size" placement="bottom">
                                                    <Button onClick={this.increaseSize} className={classes.toolBtn}>
                                                        <img alt="increase size" src={require('../icons/increase-size.svg')}
                                                             height={15}/>
                                                    </Button>
                                                </Tooltip>
                                                <Tooltip title="Decrease size" placement="bottom" className={classes.toolBtn}>
                                                    <Button onClick={this.decreaseSize}>
                                                        <img alt="decrease size" src={require('../icons/decrease-size.svg')}
                                                             height={15}/>
                                                    </Button>
                                                </Tooltip>
                                                <Tooltip title="Delete" placement="bottom">
                                                    <Button onClick={this.deleteNode}><Delete
                                                        fontSize="small"/></Button>
                                                </Tooltip>
                                            </ButtonGroup>)
                                            : null
                                    }
                                    {
                                        (isEdgeSelected && (nodeType === "link")) ?
                                            (<ButtonGroup size="small" aria-label="small outlined button group">
                                                <Tooltip title="Delete" placement="bottom">
                                                    <Button onClick={this.deleteEdge}><Delete
                                                        fontSize="small"/></Button>
                                                </Tooltip>
                                            </ButtonGroup>)
                                            : null
                                    }
                                    <ButtonGroup size="small" aria-label="small outlined button group"
                                                 className={classes.diagramTools}>
                                        <Tooltip title="Zoom in" placement="bottom">
                                            <Button onClick={this.zoomIn}><ZoomInRounded fontSize="small"/></Button>
                                        </Tooltip>
                                        <Tooltip title="Zoom out" placement="bottom">
                                            <Button onClick={this.zoomOut}><ZoomOutRounded fontSize="small"/></Button>
                                        </Tooltip>
                                        <Tooltip title="Fit to screen" placement="bottom">
                                            <Button onClick={this.fitToScreen}><ZoomOutMapOutlined
                                                fontSize="small"/></Button>
                                        </Tooltip>
                                    </ButtonGroup>
                                </Grid>
                                <Grid item>
                                    <ButtonGroup size="small" aria-label="small outlined button group"
                                                 className={classes.diagramTools}>
                                        {
                                            open ?
                                                (
                                                    <Tooltip title="Open properties" placement="bottom">
                                                        <Button aria-label="Open drawer"
                                                                onClick={this.handleDrawerClose}
                                                                className={classNames(classes.menuButton)}
                                                        >
                                                            <ArtTrackRounded fontSize="small" color="primary"/>
                                                        </Button>
                                                    </Tooltip>)
                                                : (<Tooltip title="Close properties" placement="bottom">
                                                    <Button aria-label="Open drawer"
                                                            onClick={this.handleDrawerOpen}
                                                            className={classNames(classes.menuButton, open &&
                                                                classes.active)}
                                                    >
                                                        <ArtTrackRounded fontSize="small" color="inherit"/>
                                                    </Button>
                                                </Tooltip>)
                                        }
                                    </ButtonGroup>
                                </Grid>
                            </Grid>
                        </Grid>
                    </Grid>
                </div>

                <div className={classes.root}>
                    <div className={classes.elements}>
                        <div>
                            <Tooltip title="Add Cell" placement="bottom">
                                <Button onClick={() => this.onClickNode('cell')} className={classes.elementBtn}>
                                    <img src={require('../icons/cell.svg')} height={25} width={25} alt="cell"/>
                                </Button>
                            </Tooltip>
                            <Tooltip title="Add Composite" placement="bottom">
                                <Button onClick={() => this.onClickNode('composite')}
                                        className={classes.elementBtn}>
                                    <img alt="composite" src={require('../icons/composite.svg')} height={25}/>
                                </Button>
                            </Tooltip>
                            <Tooltip title="Add Component" placement="bottom">
                                <Button onClick={() => this.onClickNode('component')}
                                        className={classes.elementBtn}>
                                    <img alt="component" src={require('../icons/component.svg')} height={25}/>
                                </Button>
                            </Tooltip>
                            <Tooltip title="Add Gateway" placement="bottom">
                                <Button onClick={() => this.onClickNode('gateway')} className={classes.elementBtn}>
                                    <img src={require('../icons/gateway.svg')} height={25} alt="gateway"/>
                                </Button>
                            </Tooltip>
                            <Tooltip title="Add Link" placement="bottom">
                                <Button onClick={this.onClickEdge} className={classes.elementBtn}>
                                    <ArrowRightAlt className={classes.toolIcon}/></Button>
                            </Tooltip>
                        </div>
                    </div>
                    <div className={classNames(classes.content, {
                        [classes.contentShift]: open
                    })}>
                        <div className={classes.graphContainer}>
                            <div className={classes.diagram}>
                                <div className={classes.graph} ref={this.dependencyGraph}/>
                            </div>
                        </div>
                    </div>

                    <React.Fragment>

                        <Drawer className={classes.drawer} variant="persistent" anchor="right"
                                open={open}
                                classes={{
                                    paper: classes.drawerPaper
                                }}>
                            <div className={classes.drawerHeader}>
                                <Typography color="textSecondary"
                                            className={classes.sideBarHeading}>
                                    Properties
                                </Typography>
                            </div>
                            <Divider/>

                            <form className={classes.inputFields} noValidate autoComplete="off">
                                <div>
                                    {
                                        (isNodeSelected && (nodeType === "cell" || nodeType === "composite")) ?
                                            (
                                                <React.Fragment>
                                                    <TextField
                                                        id="standard-multiline-flexible"
                                                        label="Image Name"
                                                        multiline
                                                        rowsMax="4"
                                                        defaultValue={this.getDefaultNodeValue("name")}
                                                        onChange={this.getChangeHandler("name")}
                                                        size="small"
                                                        inputProps={{style: {fontSize: 12}}}
                                                        InputLabelProps={{style: {fontSize: 12}}}
                                                    />
                                                    <TextField
                                                        id="standard-multiline-flexible"
                                                        label="Organization"
                                                        multiline
                                                        rowsMax="2"
                                                        defaultValue={this.getDefaultNodeValue("organization")}
                                                        onChange={this.getChangeHandler("organization")}
                                                        size="small"
                                                        inputProps={{style: {fontSize: 12}}}
                                                        InputLabelProps={{style: {fontSize: 12}}}
                                                    />
                                                    <TextField
                                                        id="standard-multiline-flexible"
                                                        label="Version"
                                                        defaultValue={this.getDefaultNodeValue("version")}
                                                        onChange={this.getChangeHandler("version")}
                                                        size="small"
                                                        inputProps={{style: {fontSize: 12}}}
                                                        InputLabelProps={{style: {fontSize: 12}}}
                                                    />
                                                </React.Fragment>

                                            ) : null
                                    }
                                    {
                                        (isNodeSelected && (nodeType === "component")) ?
                                            (
                                                <React.Fragment>
                                                    <TextField
                                                        id="standard-multiline-flexible"
                                                        label="Name"
                                                        multiline
                                                        rowsMax="4"
                                                        defaultValue={this.getDefaultNodeValue("name")}
                                                        onChange={this.getChangeHandler("name")}
                                                        size="small"
                                                        inputProps={{style: {fontSize: 12}}}
                                                        InputLabelProps={{style: {fontSize: 12}}}
                                                    />
                                                    <TextField
                                                        id="standard-multiline-flexible"
                                                        label="Source image"
                                                        defaultValue={this.getDefaultNodeValue("sourceImage")}
                                                        onChange={this.getChangeHandler("sourceImage")}
                                                        size="small"
                                                        inputProps={{style: {fontSize: 12}}}
                                                        InputLabelProps={{style: {fontSize: 12}}}
                                                    />
                                                </React.Fragment>
                                            ) : null
                                    }
                                    {
                                        (isEdgeSelected && (nodeType === "link")) ?
                                            (
                                                <TextField
                                                    id="standard-multiline-flexible"
                                                    label="Alias"
                                                    defaultValue={this.getDefaultEdgeValue()}
                                                    onChange={this.getChangeHandler("alias")}
                                                    size="small"
                                                    inputProps={{style: {fontSize: 12}}}
                                                    InputLabelProps={{style: {fontSize: 12}}}
                                                />
                                            ) : null
                                    }
                                    {
                                        (isNodeSelected && (nodeType === "gateway")) ?
                                            (
                                                <TextField
                                                    id="standard-multiline-flexible"
                                                    label="Name"
                                                    defaultValue={this.getDefaultNodeValue("name")}
                                                    disabled
                                                    size="small"
                                                    inputProps={{style: {fontSize: 12}}}
                                                    InputLabelProps={{style: {fontSize: 12}}}
                                                />
                                            ) : null
                                    }
                                    {
                                        (nodeType === "none") ?
                                            (
                                                <div className={classes.noContentMsg}>No element selected.</div>
                                            ) : null
                                    }

                                </div>
                            </form>
                        </Drawer>
                    </React.Fragment>
                    <Snackbar
                        anchorOrigin={{
                            vertical: 'bottom',
                            horizontal: 'left',
                        }}
                        open={openSnackBar}
                        autoHideDuration={6000}
                        onClose={this.handleSnackBarClose}
                        ContentProps={{
                            'aria-describedby': 'error-message',
                        }}
                        message={<span id="message-id">{errorContent}</span>}
                        action={
                            <IconButton
                                key="close"
                                aria-label="close"
                                color="inherit"
                                className={classes.close}
                                onClick={this.handleSnackBarClose}
                            >
                                <CloseIcon/>
                            </IconButton>
                        }
                    />
                </div>
            </div>
        );
    };

}

DesignerView.propTypes = {
    classes: PropTypes.object.isRequired
};

export default withStyles(styles)(DesignerView);
