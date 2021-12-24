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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/francois-poidevin/flighttracker/internal"
	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/francois-poidevin/flighttracker/internal/app/service"
	"github.com/francois-poidevin/flighttracker/internal/app/tools"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ctx, cancel         = context.WithCancel(context.Background())
	processStarted bool = false
)

type parameters struct {
	Bbox               tools.Bbox `json:"bbox"`
	AltThreshold       int        `json:"altThreshold"`
	FromTimeStampParam time.Time  `json:"fromTimeStampParam"`
	ToTimeStampParam   time.Time  `json:"toTimeStampParam"`
}

type response struct {
	Parameters parameters       `json:"parameters"`
	NbFlight   int              `json:"nbFlight"`
	Data       []app.FlightData `json:"data"`
}

// startCmd represents the start command
// see https://dev.to/moficodes/build-your-first-rest-api-with-go-2gcj
var startHttpCmd = &cobra.Command{
	Use:   "startHttp",
	Short: "Allow to start REST API service around tracking of all flights",
	Long:  `The HTTP Rest API service start with config parameters. Several endpoints are available `,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize config
		initConfig()

		if conf.Flighttracker.Sinkertype != "DB" {
			log.Fatal("Service can't be started without a Database sinker, please change config file")
		}

		r := mux.NewRouter()

		api := r.PathPrefix("/api/v1").Subrouter()
		api.HandleFunc("/start", startService).Methods(http.MethodGet)
		api.HandleFunc("/stop", stopService).Methods(http.MethodGet)
		api.HandleFunc("/search", searchService).Methods(http.MethodGet)

		//Start http server here
		log.Fatal(http.ListenAndServe(":8080", r))

	},
}

//Start collecting service
func startService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !processStarted {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "start Sinker service called"}`))

		processStarted = true

		ctx, cancel = context.WithCancel(context.Background())
		go func() {
			errExec := internal.Execute(ctx, log, *conf)
			if errExec != nil {
				log.WithContext(ctx).WithFields(logrus.Fields{
					"Error": errExec,
				}).Error("Error in Execute processing")
				os.Exit(1)
			}
		}()

	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "start Sinker service already processing"}`))
	}
}

//Stop collecting service
func stopService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if processStarted {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "stop Sinker service called and done"}`))
		cancel()
		processStarted = false
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Sinker service is not processing currently"}`))
	}
}

//Search on collecting data
// params : BBox, altitude threshold, time windows (from, to)
// return : json
func searchService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	//Check bbox parameter
	bboxParam := query.Get("bbox")
	bbox, errBBox := tools.GetBbox(bboxParam)
	if errBBox != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"message": "bbox have to be well formatted (%s)"}`, errBBox.Error())))
		return
	}
	//Check threshold parameter
	altThresholdParam := query.Get("altThresholdFeet")
	altThreshold, errAltThreshold := strconv.Atoi(altThresholdParam)
	if errAltThreshold != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"message": "need a number (%s)"}`, errAltThreshold.Error())))
		return
	}

	//Check time windows parameters
	layout := "2006-01-02T15:04:05"
	fromTimeStampParam := query.Get("fromTimeStamp")
	fromTimeStamp, errFromTimeStamp := time.Parse(layout, fromTimeStampParam)

	if errFromTimeStamp != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"message": "need a time with layout (%s) - error: %s"}`, layout, errFromTimeStamp.Error())))
		return
	}
	toTimeStampParam := query.Get("toTimeStamp")
	toTimeStamp, errToTimeStamp := time.Parse(layout, toTimeStampParam)

	if errToTimeStamp != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"message": "need a time with layout (%s) - error: %s"}`, layout, errToTimeStamp.Error())))
		return
	}

	//call logical for searching in DB
	searchSvc := service.New(log)
	//TODO: remove db connection at the starting of the startHttp service, and then pass to search service
	data, errSearch := searchSvc.Search(ctx, *&conf.Flighttracker.Postgres, bbox, altThreshold, fromTimeStamp, toTimeStamp)

	if errSearch != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"message": "internal server error (%s)"}`, errAltThreshold.Error())))
		return
	}

	parameters := parameters{
		Bbox:               bbox,
		AltThreshold:       altThreshold,
		FromTimeStampParam: fromTimeStamp,
		ToTimeStampParam:   toTimeStamp,
	}

	response := response{
		Parameters: parameters,
		NbFlight:   len(data),
		Data:       data,
	}

	result, errJsonMarshal := json.Marshal(response)
	if errJsonMarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"message": "internal server error (%s)"}`, errJsonMarshal.Error())))
		return
	}

	w.Write(result)
}

func init() {
	startHttpCmd.Flags().StringVar(&cfgFile, "config", "config_flighttracker.toml", "config file")
}
