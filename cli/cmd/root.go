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
	"fmt"
	"os"
	"strings"

	"github.com/francois-poidevin/flighttracker/config"
	defaults "github.com/mcuadros/go-defaults"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flighttracker",
	Short: "Flighttracker application allow to track all flights in a Bounding Box",
	Long: `Flighttracker application allow to track all flights in a Bounding Box 
	and detect illegal flights that not respect flight altitude rules.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	log     *logrus.Logger
	cfgFile string
	conf    = &config.Configuration{}
)

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(startHttpCmd)
	rootCmd.AddCommand(configCmd)
}
func initConfig() {
	//TODO: refactor this code for better handling env variable in case of docker (env. var. pass to docker image)
	for k := range asEnvVariables(conf, "", false) {
		err := viper.BindEnv(strings.ToLower(strings.Replace(k, "_", ".", -1)), "FT_"+k)
		if err != nil {
			log.WithFields(logrus.Fields{
				"var": "EH_" + k,
			}).Error("Unable to bind environment variable")
		}
	}

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

	if err := viper.Unmarshal(conf); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Unable to parse config")
	}
}
