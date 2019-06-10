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

/* eslint react/prefer-stateless-function: ["off"] */

import Cart from "../orders/cart";
import React from "react";
import * as PropTypes from "prop-types";

// Creating a context that can be accessed
const StateContext = React.createContext({
    cart: null,
    catalog: null,
    user: null
});

const StateProvider = ({catalog, children, user}) => (
    <StateContext.Provider value={{
        cart: new Cart(),
        catalog: catalog,
        user: user
    }}>
        {children}
    </StateContext.Provider>
);

StateProvider.propTypes = {
    children: PropTypes.any.isRequired,
    catalog: PropTypes.shape({
        accessories: PropTypes.arrayOf(PropTypes.shape({
            id: PropTypes.number.isRequired,
            name: PropTypes.string.isRequired,
            description: PropTypes.string.isRequired,
            unitPrice: PropTypes.number.isRequired,
            inStock: PropTypes.number.isRequired
        })).isRequired
    }),
    user: PropTypes.string.isRequired
};

/**
 * Higher Order Component for accessing the Cart.
 *
 * @param {React.ComponentType} Component component which needs access to the cart.
 * @returns {React.ComponentType} The new HOC with access to the cart.
 */
const withState = (Component) => {
    class StateConsumer extends React.Component {

        render = () => {
            const {forwardedRef, ...otherProps} = this.props;

            return (
                <StateContext.Consumer>
                    {(state) => <Component ref={forwardedRef} {...otherProps} {...state}/>}
                </StateContext.Consumer>
            );
        };

    }

    StateConsumer.propTypes = {
        forwardedRef: PropTypes.any
    };

    return React.forwardRef((props, ref) => <StateConsumer {...props} forwardedRef={ref} />);
};

export default withState;
export {StateProvider};
