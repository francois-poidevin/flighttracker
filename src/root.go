/*
 * Copyright (C) Continental Automotive GmbH 2019
 * Alle Rechte vorbehalten. All Rights Reserved.
 * The reproduction, transmission or use of this document or its contents is not
 * permitted without express written authority. Offenders will be liable for
 * damages. All rights, including rights created by patent grant or registration of
 * a utility model or design, are reserved.
 */

/*
 * Copyright (C) Continental Automotive GmbH 2019
 * Alle Rechte vorbehalten. All Rights Reserved.
 * The reproduction, transmission or use of this document or its contents is not
 * permitted without express written authority. Offenders will be liable for
 * damages. All rights, including rights created by patent grant or registration of
 * a utility model or design, are reserved.
 */

package src

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.zenithar.org/pkg/log"
)

//FlightData - storage structure for flightRadar24 API response
type FlightData struct {
	flightID         string
	iCAO24BITADDRESS string
	lat              float64
	lon              float64
	track            int64 //degree to the destination
	altitude         int64
	groundSpeed      int64  //kts
	unknown1         string //not describe yet
	transpondeurType string
	aircraftType     string
	immatriculation1 string
	timeStamp        float64
	origine          string
	destination      string
	unknown2         string
	verticalSpeed    int64
	immatriculation2 string
	unknown3         string
	company          string
}

const (
	feetMeter = 0.3048
)

//Execute - start the worker
func Execute() {
	ctx := context.Background()

	log.For(ctx).Info("START")

	//Loop each 5 secondes for working
	d := 5 * time.Second
	f, err := os.OpenFile("data.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.For(ctx).Error("Unable to Open file", zap.Error(err))
	}
	for x := range time.Tick(d) {
		errStore := storeDataOnFile(ctx, x, f)
		if errStore != nil {
			log.For(ctx).Error("Unable to store data", zap.Error(errStore))
		}
	}

	defer func() {
		errCloseFile := f.Close()
		if errCloseFile != nil {
			log.For(ctx).Error("unable to close file", zap.Error(errCloseFile))
		}

		log.For(ctx).Info("END")
	}()
}

func storeDataOnFile(ctx context.Context, t time.Time, f *os.File) error {

	w := bufio.NewWriter(f)

	// Made the HTTP request - Test area 43.663712,1.570358,43.710510,1.700735
	// Toulouse and Airport Area - 43.515693,1.318359,43.702630,1.687775
	resp, errHTTPGet := http.Get("https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=43.70,43.51,1.31,1.68&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1")
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

	for i := 0; i < len(data); i++ {
		log.For(ctx).Info("aircraftType", zap.String("Type", data[i].aircraftType))
		log.For(ctx).Info("immatriculation1", zap.String("immat", data[i].immatriculation1))
		log.For(ctx).Info("origine", zap.String("ORIGIN", data[i].origine))
		log.For(ctx).Info("destination", zap.String("DEST", data[i].destination))
		log.For(ctx).Info("Altitude", zap.Int64("Feet", data[i].altitude))
		log.For(ctx).Info("Altitude", zap.Float64("meters", float64(data[i].altitude)*feetMeter))
	}

	n4, errWS := w.WriteString(t.String() + "\n" + string(body) + "\n====================================\n")
	if errWS != nil {
		return errWS
	}
	log.For(ctx).Info("Wrote", zap.String("length", fmt.Sprintf("wrote %d bytes", n4)))
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
					flightID:         k,
					iCAO24BITADDRESS: fmt.Sprintf("%v", s.Index(0)),
					lat:              _lat,
					lon:              _lon,
					track:            _track,
					altitude:         _altitude,
					groundSpeed:      _groundSpeed,
					unknown1:         fmt.Sprintf("%v", s.Index(6)),
					transpondeurType: fmt.Sprintf("%v", s.Index(7)),
					aircraftType:     fmt.Sprintf("%v", s.Index(8)),
					immatriculation1: fmt.Sprintf("%v", s.Index(9)),
					timeStamp:        _timeStamp,
					origine:          fmt.Sprintf("%v", s.Index(11)),
					destination:      fmt.Sprintf("%v", s.Index(12)),
					unknown2:         fmt.Sprintf("%v", s.Index(13)),
					verticalSpeed:    _verticalSpeed,
					immatriculation2: fmt.Sprintf("%v", s.Index(15)),
					unknown3:         fmt.Sprintf("%v", s.Index(16)),
					company:          fmt.Sprintf("%v", s.Index(17)),
				}
				result = append(result, flightData)
			}
		}
	}

	return result, nil
}
