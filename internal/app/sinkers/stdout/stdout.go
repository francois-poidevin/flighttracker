package stdout

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/sirupsen/logrus"
)

type StdOutSinker struct {
	Log *logrus.Logger
}

func New(log *logrus.Logger) app.Sinker {
	//init the logger here
	return &StdOutSinker{Log: log}
}

func (s *StdOutSinker) Init(ctx context.Context, params interface{}) error {
	//Nothing to do here
	return nil
}

func (s *StdOutSinker) Sink(ctx context.Context, t time.Time, data []app.FlightData) error {
	if len(data) > 0 {
		var buffer bytes.Buffer
		var bufferIllegalFlight bytes.Buffer
		var IllegalFlight []app.FlightData

		//found flights above 500 meters that moving
		for _, dataObj := range data {
			if (float64(dataObj.Altitude)*app.FEETTOMETER) < float64(500) &&
				(float64(dataObj.Altitude)*app.FEETTOMETER) > float64(25) &&
				float64(dataObj.GroundSpeed)*app.KTSKMH > 0 {
				IllegalFlight = append(IllegalFlight, dataObj)
			}
		}

		//All flights
		Marshal, err := json.Marshal(data)
		if err != nil {
			return err
		}
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"number of Flights": len(data),
		}).Info("========All Flights seen=============")

		buffer.Write(Marshal)
		s.Log.WithContext(ctx).Debug(" Raw Datas" + buffer.String())

		//Illegal flight
		MarshalIllegalFlight, err := json.Marshal(IllegalFlight)
		if err != nil {
			return err
		}
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"number of Flights": len(IllegalFlight),
		}).Debug("========IllegalFlight Flights seen=============")

		bufferIllegalFlight.Write(MarshalIllegalFlight)
		s.Log.WithContext(ctx).Debug(" Illegal Flights" + buffer.String())

	} else {
		s.Log.WithContext(ctx).Info("No Raw data")
	}
	return nil
}
