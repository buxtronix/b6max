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

	"github.com/buxtronix/b6max"
	"github.com/spf13/cobra"
)

// chargeCmd represents the charge command
var chargeCmd = &cobra.Command{
	Use:   "charge",
	Short: "Run a battery charge program",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		flagCurrent, err := cmd.Flags().GetInt("current")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		flagVoltage, err := cmd.Flags().GetInt("voltage")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		flagCells, err := cmd.Flags().GetInt("cells")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		flagTags, err := cmd.Flags().GetStringSlice("tags")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		bType, err := batteryType(cmd)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		cType := b6max.Charge
		switch {
		case flagCurrent < 100 || flagCurrent > 6000:
			fmt.Printf("current value must be between 100 and 6000 (mA)\n")
			return
		case flagCells < 1 || flagCells > 6:
			fmt.Printf("current value must be between 1 and 6\n")
			return
		}
		p := &b6max.ProgramStart{
			BatteryType:   bType,
			PwmMode:       cType,
			Cells:         byte(flagCells),
			ChargeCurrent: uint16(flagCurrent),
			ChargeCutoff:  uint16(flagVoltage),
		}

		dev, err := open(cmd)
		if err != nil {
			fmt.Printf("Error opening device: %v\n", err)
			return
		}
		defer dev.Close()
		fmt.Printf("Starting charge cycle..")
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
	rootCmd.AddCommand(chargeCmd)
	chargeCmd.PersistentFlags().StringP("battery", "b", "", "Battery chemistry: LiIo , LiPo , LiFe , NiMh , NiCd , Pb")
	chargeCmd.Flags().BoolP("auto", "a", false, "Auto detect current/voltage (Ni* only)")
	chargeCmd.PersistentFlags().IntP("current", "c", 0, "Charge current, mA")
	chargeCmd.PersistentFlags().IntP("voltage", "v", 0, "Charge cutoff voltage, mV. For NI*, delta peak voltage.")
	chargeCmd.PersistentFlags().IntP("cells", "n", 0, "Number of series cells")
	chargeCmd.PersistentFlags().StringSliceP("tags", "t", nil, "Tags for this program")
	chargeCmd.Flags().DurationP("interval", "i", 0, "When >0, watch progress with this interval")
}
