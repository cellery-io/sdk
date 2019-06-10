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

import Cart from "./orders/cart";
import CartView from "./orders/CartView";
import Catalog from "./catalog/Catalog";
import Orders from "./orders/Orders";
import React from "react";
import SignIn from "./user/SignIn";
import SignUp from "./user/SignUp";
import withState from "./common/state";
import {withStyles} from "@material-ui/core/styles";
import {AccountCircle, ArrowBack, Pets, ShoppingCart} from "@material-ui/icons";
import {AppBar, Avatar, Badge, Button, IconButton, Menu, MenuItem, Toolbar, Typography} from "@material-ui/core";
import {Redirect, Route, Switch, withRouter} from "react-router-dom";
import * as PropTypes from "prop-types";

const styles = (theme) => ({
    appBar: {
        position: "relative"
    },
    logo: {
        marginRight: theme.spacing.unit * 2,
        cursor: "pointer"
    },
    title: {
        flexGrow: 1,
        cursor: "pointer"
    },
    badge: {
        top: "15%",
        right: -25,
        marginRight: theme.spacing.unit * 3
    },
    cartButtonContainer: {
        display: "inline",
        marginRight: theme.spacing.unit * 3
    },
    userAvatarContainer: {
        marginBottom: theme.spacing.unit * 2,
        pointerEvents: "none"
    },
    userAvatar: {
        marginRight: theme.spacing.unit * 1.5,
        color: "#fff",
        backgroundColor: theme.palette.primary.main
    }
});

class App extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            accountPopoverElement: null,
            cartItemsCount: props.cart.getItems().length
        };
    }

    componentDidMount = () => {
        const {cart} = this.props;
        cart.addListener(this.handleCartUpdates);
    };

    componentWillUnmount = () => {
        const {cart} = this.props;
        cart.removeListener(this.handleCartUpdates);
    };

    handleCartUpdates = (items) => {
        this.setState({
            cartItemsCount: items.length
        });
    };

    handleAccountPopoverOpen = (event) => {
        this.setState({
            accountPopoverElement: event.currentTarget
        });
    };

    handleAccountPopoverClose = () => {
        this.setState({
            accountPopoverElement: null
        });
    };

    signIn = () => {
        window.location.href = `${window.__BASE_PATH__}/sign-in`;
    };

    signOut = () => {
        window.location.href = `${window.__BASE_PATH__}/_auth/logout`;
    };

    render() {
        const {classes, history, location, user} = this.props;
        const {accountPopoverElement, cartItemsCount} = this.state;

        const isAccountPopoverOpen = Boolean(accountPopoverElement);
        const backButtonHiddenRoutes = ["/", "/sign-up", "/sign-in"];
        return (
            <div className={classes.root}>
                <AppBar position={"static"} className={classes.appBar}>
                    <Toolbar>
                        {
                            history.length <= 1 || backButtonHiddenRoutes.includes(location.pathname)
                                ? null
                                : (
                                    <IconButton color={"inherit"} aria-label={"Back"}
                                        onClick={() => history.goBack()}>
                                        <ArrowBack/>
                                    </IconButton>
                                )
                        }
                        <Pets className={classes.logo} onClick={() => history.push("/")}/>
                        <Typography variant={"h6"} color={"inherit"} noWrap className={classes.title}
                            onClick={() => history.push("/")}>
                            Pet Store
                        </Typography>
                        {
                            user
                                ? (
                                    <div>
                                        <div className={classes.cartButtonContainer}>
                                            <Badge color={"secondary"} badgeContent={cartItemsCount}
                                                classes={{badge: classes.badge}}>
                                                <Button color={"inherit"} onClick={() => history.push("/cart")}>
                                                    <ShoppingCart/> Cart
                                                </Button>
                                            </Badge>
                                        </div>
                                        <IconButton
                                            aria-owns={isAccountPopoverOpen ? "user-info-appbar" : undefined}
                                            color={"inherit"} aria-haspopup={"true"}
                                            onClick={this.handleAccountPopoverOpen}>
                                            <AccountCircle/>
                                        </IconButton>
                                        <Menu id={"user-info-appbar"} anchorEl={accountPopoverElement}
                                            anchorOrigin={{
                                                vertical: "top",
                                                horizontal: "right"
                                            }}
                                            transformOrigin={{
                                                vertical: "top",
                                                horizontal: "right"
                                            }}
                                            open={isAccountPopoverOpen}
                                            onClose={this.handleAccountPopoverClose}>
                                            <MenuItem onClick={this.handleAccountPopoverClose}
                                                className={classes.userAvatarContainer}>
                                                <Avatar className={classes.userAvatar}>
                                                    {user.substr(0, 1).toUpperCase()}
                                                </Avatar>
                                                {user}
                                            </MenuItem>
                                            <MenuItem onClick={this.signOut}>
                                                Sign Out
                                            </MenuItem>
                                        </Menu>
                                    </div>
                                )
                                : (
                                    <Button style={{color: "#ffffff"}} onClick={this.signIn}>Sign In</Button>
                                )
                        }
                    </Toolbar>
                </AppBar>
                <main>
                    <Switch>
                        <Route exact path={"/"} component={Catalog}/>
                        <Route exact path={"/cart"} component={CartView}/>
                        <Route exact path={"/sign-in"} component={SignIn}/>
                        <Route exact path={"/sign-up"} component={SignUp}/>
                        {
                            user
                                ? <Route exact path={"/orders"} component={Orders}/>
                                : null
                        }
                        <Redirect from={"*"} to={"/"}/>
                    </Switch>
                </main>
            </div>
        );
    }

}

App.propTypes = {
    classes: PropTypes.string.isRequired,
    cart: PropTypes.instanceOf(Cart),
    location: PropTypes.shape({
        pathname: PropTypes.string.isRequired
    }).isRequired,
    history: PropTypes.shape({
        goBack: PropTypes.func.isRequired
    }).isRequired,
    user: PropTypes.string.isRequired
};

export default withStyles(styles)(withRouter(withState(App)));
