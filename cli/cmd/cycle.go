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
	"errors"
	"fmt"
	"strings"

	"github.com/buxtronix/b6max"
	"github.com/spf13/cobra"
)

// cycleCmd represents the cycle command
var cycleCmd = &cobra.Command{
	Use:   "cycle",
	Short: "Run a series of charge/discharge cycles (NiMh/NiCd only)",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		flagBattery, err := cmd.Flags().GetString("battery")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		var bType b6max.BatteryType
		switch bt := strings.ToLower(flagBattery); bt {
		case "nimh":
			bType = b6max.NiMh
		case "nicd":
			bType = b6max.NiCd
		default:
			fmt.Printf("Unsupported battery type for cycle '%s', only NiMh and NiCd\n", bt)
			return
		}
		flagCycles, err := cmd.Flags().GetInt("cycles")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
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
		flagDCurrent, err := cmd.Flags().GetInt("dcurrent")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		flagDVoltage, err := cmd.Flags().GetInt("dvoltage")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		var checkError error
		switch {
		case flagCurrent < 100 || flagCurrent > 6000:
			checkError = errors.New("'current' flag must be >=100 and <= 6000")
		case flagDCurrent < 100 || flagDCurrent > 2000:
			checkError = errors.New("'dcurrent' flag must be >=100 and <= 2000")
		case flagDVoltage < 800:
			checkError = errors.New("'dvoltage' flag must be >=800")
		case flagCycles < 1 || flagCycles > 3:
			checkError = errors.New("'cycles' flag must be between 1 and 3")
		}
		if checkError != nil {
			fmt.Print(err.Error())
			return
		}
		p := &b6max.ProgramStart{
			BatteryType:      bType,
			PwmMode:          b6max.Cycle,
			Cells:            1,
			RePeakCycleInfo:  2,
			ChargeCurrent:    uint16(flagCurrent),
			ChargeCutoff:     uint16(flagVoltage),
			DischargeCurrent: uint16(flagDCurrent),
			DischargeCutoff:  uint16(flagDVoltage),
			CycleCount:       byte(flagCycles),
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
		if err := d.writeProgram(p, nil); err != nil {
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
	chargeCmd.AddCommand(cycleCmd)

	cycleCmd.Flags().StringP("battery", "b", "", "Battery chemistry: LiIo , LiPo , LiFe , NiMh , NiCd , Pb")
	cycleCmd.Flags().Int("cycles", 3, "Number of cycles to run")
	cycleCmd.Flags().DurationP("interval", "i", 0, "When >0, watch progress with this interval")
	cycleCmd.Flags().Int("dcurrent", 0, "Discharge current, mA")
	cycleCmd.Flags().Int("dvoltage", 0, "Discharge cutoff voltage, mV")
}
