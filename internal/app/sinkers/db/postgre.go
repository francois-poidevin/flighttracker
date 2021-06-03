package db

import (
	"context"
	"time"

	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/sirupsen/logrus"
)

//TODO add database connection
type PostGreSinker struct {
	Log *logrus.Logger
}

func New(log *logrus.Logger) app.Sinker {
	//init the logger here
	return &PostGreSinker{Log: log}
}

func (s *PostGreSinker) Init(ctx context.Context) error {
	//TODO init the connection to the database
	// create database : schema, table
	s.Log.WithContext(ctx).Info("Init DB ...")
	return nil
}

func (s *PostGreSinker) Sink(ctx context.Context, t time.Time, data []app.FlightData) error {

	//TODO do the job here
	s.Log.WithContext(ctx).Info("Insert in DB ...")
	return nil
}
