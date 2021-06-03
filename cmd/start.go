package cmd

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

import (
	"context"
	"os"

	"github.com/francois-poidevin/flighttracker/internal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allow to start troacking of all flights",
	Long: `Search in a Bounding Box (parameter) for all flights and check altitude rules.
	The application generate an output data.log file in the execution folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		//log handling
		var log = logrus.New()
		// log.Formatter = new(logrus.JSONFormatter)
		log.Formatter = new(logrus.TextFormatter)                     //default
		log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
		log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
		log.Level = logrus.TraceLevel
		log.Out = os.Stdout

		// Initialize config
		// initConfig()

		//handle parameters
		bbox, errBBox := cmd.Flags().GetString("bbox")
		if errBBox != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errBBox,
			}).Error("Error in fetching flag 'bbox'")
			os.Exit(1)
		}
		refreshTime, errRefresh := cmd.Flags().GetInt("refresh")
		if errRefresh != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errRefresh,
			}).Error("Error in fetching flag 'refresh'")
			os.Exit(1)
		}
		outputRawFileName, errOutputRaw := cmd.Flags().GetString("outputraw")
		if errOutputRaw != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errOutputRaw,
			}).Error("Error in fetching flag 'outputraw'")
			os.Exit(1)
		}
		outputReportFileName, errOutputReport := cmd.Flags().GetString("outputreport")
		if errOutputReport != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errOutputReport,
			}).Error("Error in fetching flag 'outputreport'")
			os.Exit(1)
		}
		sinkerType, errSinkerType := cmd.Flags().GetString("sinkerType")
		if errSinkerType != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errSinkerType,
			}).Error("Error in fetching flag 'sinkerType'")
			os.Exit(1)
		}
		errExec := internal.Execute(ctx, bbox, refreshTime, outputRawFileName, outputReportFileName, sinkerType, log)
		if errExec != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errExec,
			}).Error("Error in Execute processing")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().String("bbox", "",
		"Searching Bounding Box (SW^NE) 'lat,lon^lat,lon'")
	startCmd.Flags().Int("refresh", 5, "refresh time for scanning flight")
	startCmd.Flags().String("outputraw", "rawData.log", "set the output file name for raw data")
	startCmd.Flags().String("outputreport", "report.log", "set the output file name for illegal flight report")
	startCmd.Flags().String("sinkerType", "FILE", "set the sinker type (STDOUT|FILE|DB)")
}
