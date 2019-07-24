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
	"log"

	"github.com/francois-poidevin/flighttracker/internal"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allow to start troacking of all flights",
	Long: `Search in a Bounding Box (parameter) for all flights and check altitude rules.
	The application generate an output data.log file in the execution folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		bbox, errBBox := cmd.Flags().GetString("bbox")
		if errBBox != nil {
			log.Fatal(errBBox.Error())
		}
		refreshTime, errRefresh := cmd.Flags().GetInt("refresh")
		if errRefresh != nil {
			log.Fatal(errRefresh.Error())
		}
		internal.Execute(bbox, refreshTime)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().String("bbox", "",
		"Searching Bounding Box (SW^NE) 'lat,lon^lat,lon'")
	startCmd.Flags().Int("refresh", 5, "refresh time for scanning flight")
}
