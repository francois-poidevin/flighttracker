package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/sirupsen/logrus"
)

const (
	schemaname = "flighttracker"
	tablename  = "flight"
)

type PostGreSinker struct {
	Log *logrus.Logger
	db  *sql.DB
}

func New(log *logrus.Logger) app.Sinker {
	//init the logger here
	return &PostGreSinker{Log: log}
}

func (s *PostGreSinker) Init(ctx context.Context, params interface{}) error {
	parameters := params.(Configuration)

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

	// create database :
	// schema
	//TODO: reduce SQL injection
	createSchemaSQL := "CREATE SCHEMA IF NOT EXISTS " + schemaname
	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"SQL": createSchemaSQL,
	}).Info("create shema")
	_, err = s.db.Exec(createSchemaSQL)
	if err != nil {
		return err
	}

	// create database :
	// table
	createTableSQL := "CREATE TABLE IF NOT EXISTS " + schemaname + "." + tablename + " (FlightID varchar(40) NOT NULL, ICAO24BITADDRESS varchar(40), Lat decimal, Lon decimal, Track integer, Altitude integer, GroundSpeed integer, Unknown1 varchar(40), TranspondeurType varchar(40), AircraftType varchar(40), Immatriculation1 varchar(40), TimeStamp timestamp, Origine varchar(40), Destination varchar(40), Unknown2 varchar(40), VerticalSpeed integer, Immatriculation2 varchar(40), Hint varchar(40), Company varchar(40), geom geometry(Geometry,4326))"
	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"SQL": createTableSQL,
	}).Info("create table")

	_, err = s.db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostGreSinker) Sink(ctx context.Context, t time.Time, data []app.FlightData) error {

	if len(data) > 0 {
		insertSQL := "INSERT INTO " + schemaname + "." + tablename + " VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, ST_GeomFromText($20, 4326))"

		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"SQL": insertSQL,
		}).Info("Insert statement")
		nbRow := int64(0)
		for _, flight := range data {

			result, err := s.db.Exec(insertSQL,
				flight.FlightID,
				flight.ICAO24BITADDRESS,
				flight.Lat,
				flight.Lon,
				flight.Track,
				flight.Altitude,
				flight.GroundSpeed,
				flight.Unknown1,
				flight.TranspondeurType,
				flight.AircraftType,
				flight.Immatriculation1,
				time.Unix(int64(flight.TimeStamp), 0),
				flight.Origine,
				flight.Destination,
				flight.Unknown2,
				flight.VerticalSpeed,
				flight.Immatriculation2,
				flight.Hint,
				flight.Company,
				"POINT("+fmt.Sprintf("%f", flight.Lon)+" "+fmt.Sprintf("%f", flight.Lat)+")",
			)

			if err != nil {
				return err
			}

			nb, _ := result.RowsAffected()
			nbRow = nbRow + nb
		}
		s.Log.WithContext(ctx).WithFields(logrus.Fields{"Rows Affected": nbRow}).Info("Insert in DB ...")
	}

	return nil
}
