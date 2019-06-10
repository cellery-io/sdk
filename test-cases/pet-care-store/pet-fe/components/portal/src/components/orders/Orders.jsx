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

import {ExpandMore} from "@material-ui/icons";
import Notification from "../common/Notification";
import React from "react";
import classNames from "classnames";
import {withStyles} from "@material-ui/core/styles";
import {
    ExpansionPanel, ExpansionPanelDetails, ExpansionPanelSummary, Grid, Table, TableBody, TableCell, TableRow,
    Typography
} from "@material-ui/core";
import * as PropTypes from "prop-types";
import * as utils from "../../utils";

const styles = (theme) => ({
    root: {
        backgroundColor: theme.palette.background.paper
    },
    titleContent: {
        maxWidth: 600,
        margin: "0 auto",
        padding: `${theme.spacing.unit * 8}px 0 ${theme.spacing.unit * 6}px`
    },
    titleButtons: {
        marginTop: theme.spacing.unit * 4
    },
    layout: {
        width: "auto",
        marginLeft: theme.spacing.unit * 3,
        marginRight: theme.spacing.unit * 3,
        [theme.breakpoints.up(1100 + (theme.spacing.unit * 3 * 2))]: {
            width: 1100,
            marginLeft: "auto",
            marginRight: "auto"
        }
    },
    orderPanelsContainer: {
        width: "100%"
    },
    orderId: {
        fontSize: theme.typography.pxToRem(15),
        flexBasis: "30%",
        flexShrink: 0
    },
    itemCount: {
        fontSize: theme.typography.pxToRem(15),
        color: theme.palette.text.secondary,
        flexBasis: "30%",
        flexShrink: 0
    },
    price: {
        fontSize: theme.typography.pxToRem(15),
        color: theme.palette.text.secondary,
        flexBasis: "30%",
        flexShrink: 0
    },
    orderDescriptionItem: {
        padding: theme.spacing.unit
    },
    itemsTable: {
        minWidth: 700,
        marginLeft: theme.spacing.unit * 4,
        marginRight: theme.spacing.unit * 4,
        marginTop: theme.spacing.unit,
        marginBottom: theme.spacing.unit
    }
});

class Orders extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            orders: [],
            expanded: null,
            notification: {
                open: false,
                message: ""
            }
        };
    }

    handleNotificationClose = () => {
        this.setState({
            notification: {
                open: false,
                message: ""
            }
        });
    };

    /**
     * Handle changes in the expansion panels expaneded status.
     *
     * @param {number} orderId The order id of the panel for which the on change should be handled for
     * @returns {Function} The function for handling the change for the panel
     */
    handlePanelExpansionChange = (orderId) => (event, expanded) => {
        this.setState({
            expanded: expanded ? orderId : false
        });
    };

    componentDidMount() {
        const self = this;
        const config = {
            url: "/orders",
            method: "GET"
        };
        utils.callApi(config)
            .then((response) => {
                self.setState({
                    orders: response.data.orders
                });
            })
            .catch(() => {
                self.setState({
                    notification: {
                        open: true,
                        message: "Failed to fetch orders"
                    }
                });
            });
    }

    render = () => {
        const {classes} = this.props;
        const {expanded, notification, orders} = this.state;
        return (
            <div className={classes.root}>
                <div className={classes.titleContent}>
                    <Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
                        Orders
                    </Typography>
                </div>
                <div className={classNames(classes.layout, classes.cardGrid)}>
                    {
                        orders.length > 0
                            ? (
                                <div className={classes.orderPanelsContainer}>
                                    {
                                        orders.map((orderDatum) => (
                                            <ExpansionPanel key={orderDatum.id} expanded={expanded === orderDatum.id}
                                                onChange={this.handlePanelExpansionChange(orderDatum.id)}>
                                                <ExpansionPanelSummary expandIcon={<ExpandMore/>}>
                                                    <Typography className={classes.orderId}>
                                                        Order {orderDatum.id}
                                                    </Typography>
                                                    <Typography className={classes.itemCount} align={"right"}>
                                                        {orderDatum.order.reduce(
                                                            (acc, orderItem) => acc + orderItem.amount, 0)} Items
                                                    </Typography>
                                                    <Typography className={classes.price} align={"right"}>
                                                        $ {
                                                            orderDatum.order.reduce(
                                                                (acc, orderItem) => acc
                                                                    + (orderItem.item.unitPrice * orderItem.amount),
                                                                0
                                                            ).toFixed(2)
                                                        }
                                                    </Typography>
                                                </ExpansionPanelSummary>
                                                <ExpansionPanelDetails>
                                                    <Grid container>
                                                        <Grid item sm={12}>
                                                            <table>
                                                                <tr>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textPrimary"}>
                                                                            Ordered Date
                                                                        </Typography>
                                                                    </td>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textSecondary"}>
                                                                            {orderDatum.orderDate}
                                                                        </Typography>
                                                                    </td>
                                                                </tr>
                                                                <tr>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textPrimary"}>
                                                                            Delivery Date
                                                                        </Typography>
                                                                    </td>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textSecondary"}>
                                                                            {
                                                                                orderDatum.deliveryDate
                                                                                    ? orderDatum.deliveryDate
                                                                                    : "Undelivered"
                                                                            }
                                                                        </Typography>
                                                                    </td>
                                                                </tr>
                                                                <tr>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textPrimary"}>
                                                                            Delivery Address
                                                                        </Typography>
                                                                    </td>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textSecondary"}>
                                                                            {orderDatum.deliveryAddress}
                                                                        </Typography>
                                                                    </td>
                                                                </tr>
                                                                <tr>
                                                                    <td className={classes.orderDescriptionItem}>
                                                                        <Typography color={"textPrimary"}>
                                                                            Items
                                                                        </Typography>
                                                                    </td>
                                                                </tr>
                                                            </table>
                                                        </Grid>
                                                        <Grid item sm={12}>
                                                            <div className={classes.itemsTable}>
                                                                <Table>
                                                                    <TableBody>
                                                                        {
                                                                            orderDatum.order.map((orderItem) => (
                                                                                <TableRow key={orderItem.item.id}>
                                                                                    <TableCell component={"th"}
                                                                                        scope={"row"}>
                                                                                        {orderItem.item.name} x {
                                                                                            orderItem.amount}
                                                                                    </TableCell>
                                                                                    <TableCell align={"right"}>
                                                                                        $ {orderItem.item.unitPrice
                                                                                            * orderItem.amount}
                                                                                    </TableCell>
                                                                                </TableRow>
                                                                            ))
                                                                        }
                                                                    </TableBody>
                                                                </Table>
                                                            </div>
                                                        </Grid>
                                                    </Grid>
                                                </ExpansionPanelDetails>
                                            </ExpansionPanel>
                                        ))
                                    }
                                </div>
                            )
                            : (
                                <Typography variant={"body1"} align={"center"} color={"textSecondary"}>
                                    No Orders Placed
                                </Typography>
                            )
                    }
                </div>
                <Notification open={notification.open} onClose={this.handleNotificationClose}
                    message={notification.message}/>
            </div>
        );
    }

}

Orders.propTypes = {
    classes: PropTypes.object.isRequired
};

export default withStyles(styles)(Orders);
