package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.zenithar.org/pkg/log"
)

//FlightData - storage structure for flightRadar24 API response
type FlightData struct {
	FlightID         string  `json:"flightID"`
	ICAO24BITADDRESS string  `json:"ICAO24BITADDRESS"`
	Lat              float64 `json:"Lat"`
	Lon              float64 `json:"Lon"`
	Track            int64   `json:"Track"` //degree to the destination
	Altitude         int64   `json:"Altitude"`
	GroundSpeed      int64   `json:"GroundSpeed"` //kts 1kts => 1.852 kmh
	Unknown1         string  `json:"Unknown1"`    //not describe yet
	TranspondeurType string  `json:"TranspondeurType"`
	AircraftType     string  `json:"AircraftType"`
	Immatriculation1 string  `json:"Immatriculation1"`
	TimeStamp        float64 `json:"TimeStamp"`
	Origine          string  `json:"Origine"`
	Destination      string  `json:"Destination"`
	Unknown2         string  `json:"Unknown2"`
	VerticalSpeed    int64   `json:"VerticalSpeed"`
	Immatriculation2 string  `json:"Immatriculation2"`
	Unknown3         string  `json:"Unknown3"`
	Company          string  `json:"Company"`
}

// Bbox - a bounding box structure
type Bbox struct {
	latSW float64
	lonSW float64
	latNE float64
	lonNE float64
}

const (
	feetMeter = 0.3048
	ktsKmh    = 1.852
)

//Execute - start the worker
func Execute(ctx context.Context, bbox string, refreshTime int) error {
	log.For(ctx).Info("START with param: " + bbox)

	//interprete bbox parameter
	bboxStruct, errBbox := getBbox(bbox)
	if errBbox != nil {
		log.For(ctx).Error("Unable to interprate parameter bbox", zap.Error(errBbox))
		return errBbox
	}

	f, err := os.OpenFile("data.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.For(ctx).Error("Unable to Open file", zap.Error(err))
		return err
	}

	//Loop each <bbox parameter> secondes for working
	d := time.Duration(refreshTime) * time.Second
	ticker := time.NewTicker(d)

	for x := range ticker.C {
		errStore := storeDataOnFile(ctx, x, f, bboxStruct)
		if errStore != nil {
			log.For(ctx).Error("Unable to store data", zap.Error(errStore))
			return errStore
		}
	}

	defer func() {
		errCloseFile := f.Close()
		if errCloseFile != nil {
			log.For(ctx).Error("unable to close file", zap.Error(errCloseFile))
		}
		ticker.Stop()

		log.For(ctx).Info("END")
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

func storeDataOnFile(ctx context.Context, t time.Time, f *os.File, bbox Bbox) error {

	w := bufio.NewWriter(f)

	// Made the HTTP request - Test area 43.663712,1.570358,43.710510,1.700735
	// Toulouse and Airport Area - 43.515693,1.318359,43.702630,1.687775
	bounds := fmt.Sprintf("%.2f", bbox.latNE) + "," + fmt.Sprintf("%.2f", bbox.latSW) + "," + fmt.Sprintf("%.2f", bbox.lonSW) + "," + fmt.Sprintf("%.2f", bbox.lonNE)
	log.For(ctx).Info("bounds", zap.String("value", bounds))
	resp, errHTTPGet := http.Get("https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=" + bounds + "&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1")
	if errHTTPGet != nil {
		return errHTTPGet
	}
	defer func() {
		resp.Body.Close()
		w.Flush()
	}()

	body, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return errRead
	}

	data, errUnMarshal := unMarshalByte(ctx, body)

	if errUnMarshal != nil {
		return errUnMarshal
	}

	var buffer bytes.Buffer
	var IllegalFlight []FlightData
	for _, dataObj := range data {
		log.For(ctx).Info("========All Flights seen=============")
		log.For(ctx).Info("Aircraft",
			zap.String("aircraftType", dataObj.AircraftType),
			zap.String("immatriculation1", dataObj.Immatriculation1),
			zap.String("origine", dataObj.Origine),
			zap.String("destination", dataObj.Destination),
			zap.Int64("Altitude feets", dataObj.Altitude),
			zap.Float64("Altitude meters", float64(dataObj.Altitude)*feetMeter))
		//found flight above 500 meters
		if (float64(dataObj.Altitude)*feetMeter) < float64(500) &&
			(float64(dataObj.Altitude)*feetMeter) > float64(0) &&
			float64(dataObj.GroundSpeed)*ktsKmh > 0 {
			IllegalFlight = append(IllegalFlight, dataObj)
		}
	}
	if len(IllegalFlight) > 0 {
		Marshal, err := json.Marshal(IllegalFlight)
		if err != nil {
			return err
		}
		buffer.Write(Marshal)
		n4, errWS := w.WriteString(t.String() + "\n" + buffer.String() + "\n====================================\n")
		if errWS != nil {
			return errWS
		}
		log.For(ctx).Info("Wrote", zap.String("length", fmt.Sprintf("wrote %d bytes", n4)))
	} else {
		n4, errWS := w.WriteString(t.String() + "\n" + "No Illegal Flight" + "\n====================================\n")
		if errWS != nil {
			return errWS
		}
		log.For(ctx).Info("Wrote", zap.String("length", fmt.Sprintf("wrote %d bytes", n4)))
	}

	return nil
}

func unMarshalByte(ctx context.Context, byt []byte) ([]FlightData, error) {

	var data map[string]interface{}
	var result []FlightData
	if err := json.Unmarshal(byt, &data); err != nil {
		return nil, err
	}

	for k, v := range data {
		if k != "full_count" && k != "version" && k != "stats" {
			log.For(ctx).Info("Data key/value", zap.Any("value", v))
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				s := reflect.ValueOf(v)
				_lat, err := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(1)), 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _lat :", zap.Error(err))
				}
				_lon, err := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(2)), 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _lon :", zap.Error(err))
				}
				_track, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(3)), 10, 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _track :", zap.Error(err))
				}
				_altitude, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(4)), 10, 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _altitude :", zap.Error(err))
				}
				_groundSpeed, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(5)), 10, 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _groundSpeed :", zap.Error(err))
				}
				_timeStamp, err := strconv.ParseFloat(fmt.Sprintf("%f", s.Index(10)), 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _timeStamp :", zap.Error(err))
				}
				_verticalSpeed, err := strconv.ParseInt(fmt.Sprintf("%v", s.Index(14)), 10, 64)
				if err != nil {
					log.For(ctx).Error("Error in parsing _verticalSpeed :", zap.Error(err))
				}

				flightData := FlightData{
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
					Unknown3:         fmt.Sprintf("%v", s.Index(16)),
					Company:          fmt.Sprintf("%v", s.Index(17)),
				}
				result = append(result, flightData)
			}
		}
	}

	return result, nil
}
