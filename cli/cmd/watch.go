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
	"fmt"
	"time"

	"github.com/buxtronix/b6max"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch status of currently running program",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := open(cmd)
		if err != nil {
			fmt.Printf("Error opening device: %v\n", err)
			return
		}
		program, err := cmd.Flags().GetInt64("program")
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		defer dev.Close()
		d := &database{
			programId: program,
		}
		if err := d.Open(logFile); err != nil {
			fmt.Print(err.Error())
			return
		}
		defer d.Close()
		if err := doWatch(cmd, dev, d); err != nil {
			fmt.Print(err.Error())
			return
		}
	},
}

func doWatch(cmd *cobra.Command, d *b6max.B6Max, db *database) error {
	watch, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return err
	}
	if watch == 0 {
		return nil
	}
	fmt.Printf("%s\n", (&b6max.ProgramState{}).Header())
	c := time.Tick(watch)
	for range c {
		ps, err := d.ReadProgramState()
		if err != nil {
			return err
		}
		fmt.Printf("%s", ps)
		if err := db.writeInfo(ps, nil); err != nil {
			fmt.Print(err.Error())
		}

		if ps.WorkState == b6max.StateFinish {
			break
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().DurationP("interval", "i", 5*time.Second, "Interval to query charger")
	watchCmd.Flags().Int64("program", 0, "Append data to given program in db")
}
