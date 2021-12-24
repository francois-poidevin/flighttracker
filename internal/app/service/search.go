package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/francois-poidevin/flighttracker/internal/app/sinkers/db"
	"github.com/francois-poidevin/flighttracker/internal/app/tools"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Log *logrus.Logger
	db  *sql.DB
}

const (
	schemaname = "flighttracker"
	tablename  = "flight"
)

func New(log *logrus.Logger) app.Service {
	//init the logger here
	return &Service{Log: log}
}

func (s *Service) Search(ctx context.Context, params interface{}, bbox tools.Bbox, altThresholdFeet int, fromTimeStamp, toTimeStamp time.Time) ([]app.FlightData, error) {
	//Do the search logical here
	s.Log.WithContext(ctx).Info("Search service called")

	//check if service have a db connection
	if s.db == nil {
		s.Log.WithContext(ctx).Info("Search service - init DB")
		s.init(ctx, params)
	}

	//search SQL statement
	selectSQLstmt := "SELECT flightID, iCAO24BITADDRESS, lat, lon, track, altitude, groundSpeed, unknown1, transpondeurType, aircraftType, immatriculation1, timeStamp, origine, destination, unknown2, verticalSpeed, immatriculation2, hint, company FROM " + schemaname + "." + tablename + " WHERE ST_WITHIN(geom, ST_GEOMFROMTEXT($1, 4326)) AND Altitude <= $2 AND TimeStamp BETWEEN $3 AND $4"

	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"SQL": selectSQLstmt,
	}).Info("Select statement")

	rows, errQuery := s.db.Query(selectSQLstmt,
		tools.BboxToWKT(bbox),
		altThresholdFeet,
		fromTimeStamp,
		toTimeStamp,
	)

	if errQuery != nil {
		return nil, errQuery
	}

	result := make([]app.FlightData, 0)

	var (
		flight           app.FlightData
		flightID         string
		iCAO24BITADDRESS string
		lat              float64
		lon              float64
		track            int64
		altitude         int64
		groundSpeed      int64
		unknown1         string
		transpondeurType string
		aircraftType     string
		immatriculation1 string
		timeStamp        time.Time
		origine          string
		destination      string
		unknown2         string
		verticalSpeed    int64
		immatriculation2 string
		hint             string
		company          string
	)

	defer rows.Close()

	for rows.Next() {
		if errScan := rows.Scan(&flightID, &iCAO24BITADDRESS, &lat, &lon, &track, &altitude, &groundSpeed, &unknown1, &transpondeurType, &aircraftType, &immatriculation1, &timeStamp, &origine, &destination, &unknown2, &verticalSpeed, &immatriculation2, &hint, &company); errScan != nil {
			return nil, errScan
		}

		flight.FlightID = flightID
		flight.ICAO24BITADDRESS = iCAO24BITADDRESS
		flight.Lat = lat
		flight.Lon = lon
		flight.Track = track
		flight.Altitude = altitude
		flight.GroundSpeed = groundSpeed
		flight.Unknown1 = unknown1
		flight.TranspondeurType = transpondeurType
		flight.AircraftType = aircraftType
		flight.Immatriculation1 = immatriculation1
		flight.TimeStamp = float64(timeStamp.Unix())
		flight.Origine = origine
		flight.Destination = destination
		flight.Unknown2 = unknown2
		flight.VerticalSpeed = verticalSpeed
		flight.Immatriculation2 = immatriculation2
		flight.Hint = hint
		flight.Company = company

		result = append(result, flight)
	}

	errRow := rows.Err()
	if errRow != nil {
		return nil, errRow
	}

	return result, nil
}

func (s *Service) init(ctx context.Context, params interface{}) error {
	parameters := params.(db.Configuration)

	// Init the connection to the database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		parameters.Host, parameters.Port, parameters.User, parameters.Password, parameters.Dbname)
	s.Log.WithContext(ctx).Info("Init DB ... : " + psqlInfo)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	s.Log.WithContext(ctx).Info("Successfully connected : " + parameters.Host)

	s.db = db

	return nil
}
