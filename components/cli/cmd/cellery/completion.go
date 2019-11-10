/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
	"os"

	"github.com/spf13/cobra"
)

// newCompletionCommand creates a cobra command which can be invoked to generate completion scripts.
// To properly use this command it should be sourced.
func newCompletionCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <type>",
		Short: "Generate bash/zsh completion scripts",
		Long: "This generates bash/zsh completion scripts. " +
			"To load completion scripts run\n\n" +
			". <(cellery completion bash)\n\n" +
			"To configure your bash shell to load completions for each session add to your bashrc\n\n" +
			"# ~/.bashrc or ~/.profile\n" +
			". <(cellery completion bash)\n",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			err = cobra.OnlyValidArgs(cmd, args)
			if err != nil {
				return err
			}
			return nil
		},
		ValidArgs: []string{"bash", "zsh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "bash" {
				return root.GenBashCompletion(os.Stdout)
			} else {
				return root.GenZshCompletion(os.Stdout)
			}
		},
		Example: "  cellery completion bash" +
			"  cellery completion zsh",
	}
	return cmd
}
