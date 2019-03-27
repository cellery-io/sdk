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

import CellDiagram from "./CellDiagram";
import React from "react";
import * as PropTypes from "prop-types";

class CellDiagramView extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            focused: props.focusedCell
        };
    }

    onClickCell = (nodeId) => {
        this.setState({
            focused: nodeId
        });
    };

    render = () => {
        const {data} = this.props;

        return (
            <div>
                <CellDiagram data={data} focusedCell={this.state.focused} onClickNode={this.onClickCell}/>
            </div>);
    };

}

CellDiagramView.propTypes = {
    data: PropTypes.object.isRequired,
    focusedCell: PropTypes.string.isRequired
};

export default CellDiagramView;
