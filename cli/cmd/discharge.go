/*
Copyright © 2019 Ben Buxton <bbuxton@gmail.com>

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
	"strings"

	"github.com/buxtronix/b6max"
	"github.com/spf13/cobra"
)

// dischargeCmd represents the discharge command
var dischargeCmd = &cobra.Command{
	Use:   "discharge",
	Short: "Run a battery discharge cycle",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		flagBattery, err := cmd.Flags().GetString("battery")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		flagCurrent, err := cmd.Flags().GetInt("current")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		if flagCurrent < 100 || flagCurrent > 2000 {
			fmt.Printf("current value must be between 100 and 2000 (mA)\n")
			return
		}
		flagVoltage, err := cmd.Flags().GetInt("voltage")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		flagCells, err := cmd.Flags().GetInt("cells")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		if flagCells < 1 || flagCells > 6 {
			fmt.Printf("current value must be between 1 and 6\n")
			return
		}
		flagTags, err := cmd.Flags().GetStringSlice("tags")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		p := &b6max.ProgramStart{
			PwmMode:          b6max.Discharge,
			Cells:            byte(flagCells),
			DischargeCurrent: uint16(flagCurrent),
			DischargeCutoff:  uint16(flagVoltage),
		}
		switch bt := strings.ToLower(flagBattery); bt {
		case "liio":
			p.BatteryType = b6max.LiIo
		case "lipo":
			p.BatteryType = b6max.LiPo
		case "life":
			p.BatteryType = b6max.LiFe
		case "nimh":
			p.BatteryType = b6max.NiMh
		case "nicd":
			p.BatteryType = b6max.NiCd
		case "pb":
			p.BatteryType = b6max.Pb
		default:
			fmt.Printf("Unknown battery type: '%s'\n", bt)
		}

		dev, err := open(cmd)
		if err != nil {
			fmt.Printf("Error opening device: %v\n", err)
			return
		}
		defer dev.Close()
		fmt.Printf("Starting discharge cycle..")
		if err := dev.Start(p); err != nil {
			fmt.Printf(err.Error())
			return
		}
		fmt.Printf("ok.\n")
		d := &database{}
		if err := d.Open(logFile); err != nil {
			fmt.Print(err.Error())
			return
		}
		defer d.Close()
		if err := d.writeProgram(p, flagTags); err != nil {
			fmt.Print(err.Error())
			return
		}
		if err := doWatch(cmd, dev, d); err != nil {
			fmt.Print(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(dischargeCmd)
	dischargeCmd.Flags().StringP("battery", "b", "", "Battery chemistry: LiIo , LiPo , LiFe , NiMh , NiCd , Pb")
	dischargeCmd.Flags().IntP("current", "c", 0, "Discharge current, mA")
	dischargeCmd.Flags().IntP("voltage", "v", 0, "Discharge cutoff voltage, mV")
	dischargeCmd.Flags().IntP("cells", "n", 0, "Number of series cells")
	dischargeCmd.PersistentFlags().StringSliceP("tags", "t", nil, "Tags for this program")
	dischargeCmd.Flags().DurationP("interval", "i", 0, "When >0, watch progress with this interval")
}
