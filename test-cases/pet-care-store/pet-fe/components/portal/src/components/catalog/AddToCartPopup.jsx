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

import Cart from "../orders/cart";
import {Check} from "@material-ui/icons";
import Notification from "../common/Notification";
import React from "react";
import withState from "../common/state";
import {withStyles} from "@material-ui/core/styles";
import {
    Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField, Typography, Zoom
} from "@material-ui/core";
import * as PropTypes from "prop-types";

const styles = (theme) => ({
    description: {
        paddingTop: theme.spacing.unit * 3,
        paddingBottom: theme.spacing.unit * 3
    },
    totalAmountContainer: {
        justifyContent: "flex-end"
    },
    totalAmountItems: {
        display: "inline"
    }
});

class AddToCartPopup extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            addToCartEnabled: true,
            amount: 1,
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

    handleAmountChange = (event) => {
        const {item} = this.props;
        const elementValue = event.currentTarget.value;
        if (elementValue === "") {
            this.setState({
                amount: "",
                addToCartEnabled: false
            });
        } else {
            const value = parseInt(elementValue, 10);
            if (value > 0 && value <= item.inStock) {
                this.setState({
                    amount: value,
                    addToCartEnabled: true
                });
            }
        }
    };

    handleAddToCart = () => {
        const {cart, item, onClose} = this.props;
        const {amount} = this.state;
        cart.addItem(item.id, amount);
        onClose();
        this.setState({
            notification: {
                open: true,
                message: `Added ${item.name} to the Cart`
            }
        });
    };

    render() {
        const {classes, item, onClose, open} = this.props;
        const {addToCartEnabled, amount, notification} = this.state;
        return (
            <React.Fragment>
                <Dialog maxWidth={"sm"} open={open} onClose={onClose} TransitionComponent={Zoom}
                    aria-labelledby={"add-to-cart-dialog"}>
                    <DialogTitle id={"add-to-cart-dialog"}>Add To Cart</DialogTitle>
                    <DialogContent>
                        <div>
                            <Typography variant={"h5"} color={"textPrimary"}>{item.name}</Typography>
                            <Typography className={classes.description} variant={"body1"} color={"textSecondary"}>
                                {item.description}
                            </Typography>
                            <table>
                                <tr>
                                    <td><Typography variant={"body2"} color={"textPrimary"}>In Stock</Typography></td>
                                    <td><Typography variant={"body2"} color={"textPrimary"}>:</Typography></td>
                                    <td><Typography color={"textSecondary"}>{item.inStock}</Typography></td>
                                </tr>
                                <tr>
                                    <td><Typography variant={"body2"} color={"textPrimary"}>Unit Price</Typography></td>
                                    <td><Typography variant={"body2"} color={"textPrimary"}>:</Typography></td>
                                    <td><Typography color={"textSecondary"}>$ {item.unitPrice}</Typography></td>
                                </tr>
                            </table>
                            <TextField id={"amount"} label={"Amount"} type={"number"} value={amount} margin={"normal"}
                                onChange={this.handleAmountChange}
                                InputLabelProps={{
                                    shrink: true
                                }}/>
                            <div className={classes.totalAmountContainer}>
                                <Typography className={classes.totalAmountItems} variant={"body2"}
                                    color={"textPrimary"}>
                                    Total:
                                </Typography>
                                <Typography className={classes.totalAmountItems} color={"textSecondary"}>
                                    $ {item.unitPrice * amount}
                                </Typography>
                            </div>
                        </div>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={onClose} size={"small"}>Cancel</Button>
                        <Button disabled={amount === 0 || !addToCartEnabled} variant={"contained"}
                            color={"primary"} size={"small"} onClick={this.handleAddToCart}>
                            <Check/> Add to Cart
                        </Button>
                    </DialogActions>
                </Dialog>
                <Notification open={notification.open} onClose={this.handleNotificationClose}
                    message={notification.message}/>
            </React.Fragment>
        );
    }

}

AddToCartPopup.propTypes = {
    classes: PropTypes.object.isRequired,
    cart: PropTypes.instanceOf(Cart).isRequired,
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
    item: {
        id: PropTypes.number.isRequired,
        name: PropTypes.string.isRequired,
        description: PropTypes.string.isRequired,
        unitPrice: PropTypes.number.isRequired,
        inStock: PropTypes.number.isRequired
    }.inStock
};

export default withStyles(styles)(withState(AddToCartPopup));
