package app

import (
	"context"
	"time"
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
	Hint             string  `json:"Hint"`
	Company          string  `json:"Company"`
}

const (
	FEETTOMETER = 0.3048
	KTSKMH      = 1.852
)

type Sinker interface {
	Init(ctx context.Context, params interface{}) error
	Sink(ctx context.Context, t time.Time, data []FlightData) error
}
