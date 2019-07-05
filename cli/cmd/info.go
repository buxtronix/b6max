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

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Query and display info about the attached charger",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := open(cmd)
		if err != nil {
			fmt.Printf("Error opening device: %v\n", err)
			return
		}
		defer dev.Close()

		di, err := dev.ReadDevInfo()
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fmt.Printf("%s\n", di)

		si, err := dev.ReadSysInfo()
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fmt.Printf("%s\n", si)

		ps, err := dev.ReadProgramState()
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fmt.Printf("%s\n%s\n", ps.Header(), ps)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
