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
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.zenithar.org/pkg/log"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allow to start troacking of all flights",
	Long: `Search in a Bounding Box (parameter) for all flights and check altitude rules.
	The application generate an output data.log file in the execution folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		bbox, errBBox := cmd.Flags().GetString("bbox")
		if errBBox != nil {
			log.For(ctx).Error("Error in fetching flag bbox", zap.Error(errBBox))
			os.Exit(1)
		}
		refreshTime, errRefresh := cmd.Flags().GetInt("refresh")
		if errRefresh != nil {
			log.For(ctx).Error("Error in fetching flag refresh", zap.Error(errRefresh))
			os.Exit(1)
		}
		errExec := internal.Execute(ctx, bbox, refreshTime)
		if errExec != nil {
			log.For(ctx).Error("Error in Execute processing", zap.Error(errExec))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().String("bbox", "",
		"Searching Bounding Box (SW^NE) 'lat,lon^lat,lon'")
	startCmd.Flags().Int("refresh", 5, "refresh time for scanning flight")
}
