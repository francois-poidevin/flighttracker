package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/francois-poidevin/flighttracker/config"
	"github.com/francois-poidevin/flighttracker/internal/app"
	pgSinker "github.com/francois-poidevin/flighttracker/internal/app/sinkers/db"
	fileSinker "github.com/francois-poidevin/flighttracker/internal/app/sinkers/file"
	stdoutSinker "github.com/francois-poidevin/flighttracker/internal/app/sinkers/stdout"
	"github.com/sirupsen/logrus"
)

// Bbox - a bounding box structure
type Bbox struct {
	latSW float64
	lonSW float64
	latNE float64
	lonNE float64
}

//Execute - start the worker
func Execute(ctx context.Context,
	log *logrus.Logger,
	conf config.Configuration) error {

	log.WithContext(ctx).WithFields(logrus.Fields{
		"bbox":                 conf.Flighttracker.Bbox,
		"refreshTime (sec)":    conf.Flighttracker.Refresh,
		"outputRawFileName":    conf.Flighttracker.File.Outputraw,    //TODO: use it in file sinker
		"outputReportFileName": conf.Flighttracker.File.Outputreport, //TODO: use it in file sinker
		"sinkerType":           conf.Flighttracker.Sinkertype,
		"dbHost":               conf.Flighttracker.Postgres.Host,
		"dbPort":               conf.Flighttracker.Postgres.Port,
		"dbPassword":           conf.Flighttracker.Postgres.Password,
		"dbUser":               conf.Flighttracker.Postgres.User,
		"dbName":               conf.Flighttracker.Postgres.Dbname,
	}).Info("START with Configuration params: ")

	//interprete bbox parameter
	bboxStruct, errBbox := getBbox(conf.Flighttracker.Bbox)
	if errBbox != nil {
		log.WithContext(ctx).WithFields(logrus.Fields{
			"Error": errBbox,
		}).Error("Unable to interpret parameter bbox")
		return errBbox
	}

	if conf.Flighttracker.Sinkertype == "FILE" {
		log.WithContext(ctx).Info("Initiate File Sinker")
		sinker := fileSinker.New(log)
		//init sinker object (files)
		errInit := sinker.Init(ctx)
		if errInit != nil {
			log.WithContext(ctx).Error(errInit)
			return errInit
		}
		//launch the ticking
		errFileSink := ticking(ctx, conf.Flighttracker.Refresh, bboxStruct, sinker, log)
		if errFileSink != nil {
			log.WithContext(ctx).Error(errFileSink)
			return errFileSink
		}
	} else if conf.Flighttracker.Sinkertype == "STDOUT" {
		log.WithContext(ctx).Info("Initiate stdOut Sinker")
		sinker := stdoutSinker.New(log)
		//launch the ticking
		errStdOutSink := ticking(ctx, conf.Flighttracker.Refresh, bboxStruct, sinker, log)
		if errStdOutSink != nil {
			log.WithContext(ctx).Error(errStdOutSink)
			return errStdOutSink
		}
	} else if conf.Flighttracker.Sinkertype == "DB" {
		log.WithContext(ctx).Info("Initiate DB Sinker")
		sinker := pgSinker.New(log)
		//init sinker object (files)
		errInit := sinker.Init(ctx)
		if errInit != nil {
			log.WithContext(ctx).Error(errInit)
			return errInit
		}
		errDBSink := ticking(ctx, conf.Flighttracker.Refresh, bboxStruct, sinker, log)
		if errDBSink != nil {
			log.WithContext(ctx).Error(errDBSink)
			return errDBSink
		}
	} else {
		return errors.New("Wrong sinker specified")
	}

	return nil
}

//TODO: handle how to handle signals
func sigCatch(ctx context.Context, sinker app.Sinker, log *logrus.Logger) {

	sigc := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		log.WithContext(ctx).Info("Signal: " + s.String())
		done <- true
	}()

	<-done
}

func ticking(ctx context.Context, refreshTime int, bbox Bbox, sinker app.Sinker, log *logrus.Logger) error {
	//Loop each <bbox parameter> secondes for working
	d := time.Duration(refreshTime) * time.Second
	ticker := time.NewTicker(d)

	for x := range ticker.C {
		//get Raw datas
		rawData, errRaw := getRawData(ctx, bbox, log)
		if errRaw != nil {
			log.WithContext(ctx).WithFields(logrus.Fields{
				"Error": errRaw,
			}).Error("Unable to get Raw data")
			return errRaw
		}
		errSink := sinker.Sink(ctx, x, rawData)
		if errSink != nil {
			log.WithContext(ctx).Error(errSink)
		}
	}

	defer func() {
		ticker.Stop()
	}()

	return nil
}

func getBbox(data string) (Bbox, error) {
	sWnE := strings.Split(data, "^")
	result := Bbox{}
	if len(sWnE) != 2 {
		return result, errors.New("Bounding Box malformed - need ^ for separating SW and NE coordinate")
	}

	for idx, latlonRec := range sWnE {
		latlon := strings.Split(latlonRec, ",")
		if len(latlon) != 2 {
			return result, errors.New("Bounding Box malformed - need , for separating lat and lon coordinate")
		}
		lat, errLat := strconv.ParseFloat(latlon[0], 64)
		if errLat != nil {
			return result, errLat
		}
		lon, errLon := strconv.ParseFloat(latlon[1], 64)
		if errLon != nil {
			return result, errLon
		}
		if idx == 0 {
			result.latSW = lat
			result.lonSW = lon
		} else {

			result.latNE = lat
			result.lonNE = lon
		}
	}
	return result, nil
}

func getRawData(ctx context.Context, bbox Bbox, log *logrus.Logger) ([]app.FlightData, error) {
	// Made the HTTP request - Test area 43.663712,1.570358,43.710510,1.700735
	// Toulouse and Airport Area - 43.515693,1.318359,43.702630,1.687775
	bounds := fmt.Sprintf("%.2f", bbox.latNE) + "," + fmt.Sprintf("%.2f", bbox.latSW) + "," + fmt.Sprintf("%.2f", bbox.lonSW) + "," + fmt.Sprintf("%.2f", bbox.lonNE)
	resp, errHTTPGet := http.Get("https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=" + bounds + "&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1")
	if errHTTPGet != nil {
		return nil, errHTTPGet
	}
	defer func() {
		resp.Body.Close()
	}()

	body, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}

	return unMarshalByte(ctx, body, log)
}

func unMarshalByte(ctx context.Context, byt []byte, log *logrus.Logger) ([]app.FlightData, error) {

	var data map[string]interface{}
	var result []app.FlightData
	if err := json.Unmarshal(byt, &data); err != nil {
		return nil, err
	}

	for k, v := range data {
		if k != "full_count" && k != "version" && k != "stats" {
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				s := reflect.ValueOf(v)
				_lat, err := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(1)), 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _lat :": err,
					}).Error()
				}
				_lon, err := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(2)), 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _lon :": err,
					}).Error()
				}
				_track, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(3)), 10, 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _track :": err,
					}).Error()
				}
				_altitude, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(4)), 10, 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _altitude :": err,
					}).Error()
				}
				_groundSpeed, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(5)), 10, 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _groundSpeed :": err,
					}).Error()
				}
				_timeStamp, err := strconv.ParseFloat(fmt.Sprintf("%f", s.Index(10)), 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _timeStamp :": err,
					}).Error()
				}
				_verticalSpeed, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(14)), 10, 64)
				if err != nil {
					log.WithContext(ctx).WithFields(logrus.Fields{
						"Error in parsing _verticalSpeed :": err,
					}).Error()
				}

				flightData := app.FlightData{
					FlightID:         k,
					ICAO24BITADDRESS: fmt.Sprintf("%v", s.Index(0)),
					Lat:              _lat,
					Lon:              _lon,
					Track:            _track,
					Altitude:         _altitude,
					GroundSpeed:      _groundSpeed,
					Unknown1:         fmt.Sprintf("%v", s.Index(6)),
					TranspondeurType: fmt.Sprintf("%v", s.Index(7)),
					AircraftType:     fmt.Sprintf("%v", s.Index(8)),
					Immatriculation1: fmt.Sprintf("%v", s.Index(9)),
					TimeStamp:        _timeStamp,
					Origine:          fmt.Sprintf("%v", s.Index(11)),
					Destination:      fmt.Sprintf("%v", s.Index(12)),
					Unknown2:         fmt.Sprintf("%v", s.Index(13)),
					VerticalSpeed:    _verticalSpeed,
					Immatriculation2: fmt.Sprintf("%v", s.Index(15)),
					Hint:             fmt.Sprintf("%v", s.Index(16)),
					Company:          fmt.Sprintf("%v", s.Index(17)),
				}
				result = append(result, flightData)
			}
		}
	}

	return result, nil
}
