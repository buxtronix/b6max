/*
Copyright © 2019 Ben Buxton

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
	"path/filepath"
	"fmt"
	"github.com/buxtronix/b6max"
	"github.com/google/logger"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var logFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "b6max",
	Short: "A utility for B6Max/ChargeMaster style battery chargers.",
	Long: `Control program for SkyRC IMAX battery chargers.

Allows for querying and controlling these battery chargers
including charging, discharging, repeaking, rebalancing and more.

Supports chargers with direct USB connection, tested with
SkyRC IMAX B6AC V2.
`,
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func open(cmd *cobra.Command) (*b6max.B6Max, error) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	logger.Init("default", verbose, false, ioutil.Discard)
	devFlag, err := cmd.Flags().GetString("device")
	if err != nil {
		return nil, err
	}
	devs, err := b6max.Find()
	if err != nil {
		return nil, err
	}
	for _, d := range devs {
		if devFlag == "" || devFlag == d.UsbInfo.Path {
			return d, d.Open()
		}
	}
	if devFlag != "" {
		return nil, fmt.Errorf("no B6Max devices found at %s", devFlag)
	}
	return nil, fmt.Errorf("no B6Max devices found")
}

func batteryType(cmd *cobra.Command) (b6max.BatteryType, error) {
	var bType b6max.BatteryType
	flagBattery, err := cmd.Flags().GetString("battery")
	if err != nil {
		return bType, nil
	}
	switch bt := strings.ToLower(flagBattery); bt {
	case "liio":
		bType = b6max.LiIo
	case "lipo":
		bType = b6max.LiPo
	case "life":
		bType = b6max.LiFe
	case "nimh":
		bType = b6max.NiMh
	case "nicd":
		bType = b6max.NiCd
	case "pb":
		bType = b6max.Pb
	default:
		return bType, fmt.Errorf("Unknown battery type: '%s'\n", bt)
	}
	return bType, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	homeDir, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose logging")
	rootCmd.PersistentFlags().StringP("device", "d", "", "Open specific USB device")
	rootCmd.PersistentFlags().StringVar(&logFile, "dblog", filepath.Join(homeDir, ".b6max.db"), "Database to log programs to")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
