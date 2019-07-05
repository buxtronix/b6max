/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the currently running program",
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := open(cmd)
		if err != nil {
			fmt.Printf("Error opening device: %v\n", err)
			return
		}
		defer dev.Close()
		if _, err := dev.Stop(); err != nil {
			fmt.Printf("Error sending 'stop' command: %v", err)
			return
		}
		fmt.Printf("Successfully issued 'stop' command.\n")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
