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
	Short: "Allow to start tracking of all flights",
	Long: `Search in a Bounding Box (parameter) for all flights and check altitude rules.
	The application generate an output data.log file in the execution folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Initialize config
		initConfig()

		errExec := internal.Execute(ctx, log, *conf)
		if errExec != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errExec,
			}).Error("Error in Execute processing")
			os.Exit(1)
		}
	},
}

func init() {

	//log handling
	log = logrus.New()
	// log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)                     //default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	startCmd.Flags().StringVar(&cfgFile, "config", "config_flighttracker.toml", "config file")
}
