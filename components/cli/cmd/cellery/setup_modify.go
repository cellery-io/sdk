/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

var apimEnabled = false
var observabilityEnabled = false
var scaleToZeroEnabled = false
var hpaEnabled = false
var apimgtChange = runtime.NoChange
var observabilityChange = runtime.NoChange
var scaleToZeroChange = runtime.NoChange
var hpaChange = runtime.NoChange

func newSetupModifyCommand() *cobra.Command {
	var apimgt = ""
	var observability = ""
	var scaleToZero = ""
	var hpa = ""
	cmd := &cobra.Command{
		Use:   "modify <command>",
		Short: "Modify Cellery runtime",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if runtime.IsGcpRuntime() {
				if strings.TrimSpace(hpa) != "" {
					fmt.Printf("HPA cannot be changed in gcp runtime")
					hpa = ""
				}
			}
			invalidInputMessage := "invalid input for"
			acceptedValuesMessage := "expected values are enable or disable"
			if !isValidInput(apimgt) {
				return fmt.Errorf("%s apim: %s. %s", invalidInputMessage, apimgt, acceptedValuesMessage)
			}
			if !isValidInput(observability) {
				return fmt.Errorf("%s observability: %s. %s", invalidInputMessage, observability, acceptedValuesMessage)
			}
			if !isValidInput(scaleToZero) {
				return fmt.Errorf("%s scale-to-zero: %s. %s", invalidInputMessage, scaleToZero, acceptedValuesMessage)
			}
			if !isValidInput(hpa) {
				return fmt.Errorf("%s hpa: %s. %s", invalidInputMessage, hpa, acceptedValuesMessage)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			apimgtChange = convertToSelection(apimgt)
			observabilityChange = convertToSelection(observability)
			scaleToZeroChange = convertToSelection(scaleToZero)
			hpaChange = convertToSelection(hpa)
			// Check the user inputs against the current runtime
			validateRuntime()
			setup.RunSetupModify(apimgtChange, observabilityChange, scaleToZeroChange, hpaChange)
		},
		Example: "  cellery setup modify --apim=enable --observability=disable",
	}
	cmd.Flags().StringVar(&apimgt, "apim", "", "enable or disable API Management in the runtime")
	cmd.Flags().StringVar(&observability, "observability", "", "enable or disable observability in the runtime")
	cmd.Flags().StringVar(&scaleToZero, "scale-to-zero", "", "enable or disable scale to zero in the runtime")
	cmd.Flags().StringVar(&hpa, "hpa", "", "enable or disable hpa in the runtime")
	return cmd
}

func convertToSelection(change string) runtime.Selection {
	if change == "enable" {
		return runtime.Enable
	} else if change == "disable" {
		return runtime.Disable
	}
	return runtime.NoChange
}

func validateRuntime() {
	var err error
	// If user's desired apim change already exists in the runtime do not change apim behavior
	if apimgtChange != runtime.NoChange {
		apimEnabled, err = runtime.IsApimEnabled()
		if err != nil {
			util.ExitWithErrorMessage("Failed to set flag for apim", err)
		}
		if (apimEnabled && apimgtChange == runtime.Enable) || (!apimEnabled && apimgtChange == runtime.Disable) {
			apimgtChange = runtime.NoChange
		}
	}
	// If user's desired observability change already exists in the runtime do not change observability behavior
	if observabilityChange != runtime.NoChange {
		observabilityEnabled, err = runtime.IsObservabilityEnabled()
		if err != nil {
			util.ExitWithErrorMessage("Failed to set flag for observability", err)
		}
		if (observabilityEnabled && observabilityChange == runtime.Enable) || (!observabilityEnabled &&
			observabilityChange == runtime.Disable) {
			observabilityChange = runtime.NoChange
		}
	}
	// If user's desired scale to zero change already exists in the runtime do not change scale to zero behavior
	if scaleToZeroChange != runtime.NoChange {
		scaleToZeroEnabled, err = runtime.IsKnativeEnabled()
		if err != nil {
			util.ExitWithErrorMessage("Failed to set flag for scale to zero", err)
		}
		if (scaleToZeroEnabled && scaleToZeroChange == runtime.Enable) || (!scaleToZeroEnabled &&
			scaleToZeroChange == runtime.Disable) {
			scaleToZeroChange = runtime.NoChange
		}
	}
	// If user's desired hpa change already exists in the runtime do not change hpa behavior
	if hpaChange != runtime.NoChange {
		hpaEnabled, err = runtime.IsHpaEnabled()
		if err != nil {
			util.ExitWithErrorMessage("Failed to set flag for hpa", err)
		}
		if (hpaEnabled && hpaChange == runtime.Enable) || (!hpaEnabled &&
			hpaChange == runtime.Disable) {
			hpaChange = runtime.NoChange
		}
	}
}

func isValidInput(change string) bool {
	if change == "enable" || change == "disable" || change == "" {
		return true
	}
	return false
}
