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
	"strings"

	"github.com/francois-poidevin/flighttracker/config"
	"github.com/francois-poidevin/flighttracker/internal"
	defaults "github.com/mcuadros/go-defaults"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log     *logrus.Logger
	cfgFile string
	conf    = &config.Configuration{}
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allow to start troacking of all flights",
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
	startCmd.Flags().String("bbox", "43.52,1.32^43.70,1.69",
		"Searching Bounding Box (SW^NE) 'lat,lon^lat,lon'")
	startCmd.Flags().Int("refresh", 5, "refresh time for scanning flight")
	startCmd.Flags().String("outputraw", "rawData.log", "set the output file name for raw data")
	startCmd.Flags().String("outputreport", "report.log", "set the output file name for illegal flight report")
	startCmd.Flags().String("sinkerType", "FILE", "set the sinker type (STDOUT|FILE|DB)")
}

func initConfig() {
	// tag::docConfigEnvVariable[]
	//TODO: refactor this code for better handling env variable in case of docker (env. var. pass to docker image)
	for k := range asEnvVariables(conf, "", false) {
		err := viper.BindEnv(strings.ToLower(strings.Replace(k, "_", ".", -1)), "EH_"+k)
		if err != nil {
			log.WithFields(logrus.Fields{
				"var": "EH_" + k,
			}).Error("Unable to bind environment variable")
		}
	}
	// end::docConfigEnvVariable[]

	switch {
	case cfgFile != "":
		// If the config file doesn't exists, let's exit
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("File doesn't exists")
		}

		log.WithFields(logrus.Fields{
			"File": cfgFile,
		}).Info("Reading configuration file")

		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Unable to read config")
		}
	default:
		defaults.SetDefaults(conf)
	}

	// tag::docConfigEnvVariable[]
	if err := viper.Unmarshal(conf); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Unable to parse config")
	}
	// end::docConfigEnvVariable[]
}
