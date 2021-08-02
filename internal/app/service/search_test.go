package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/francois-poidevin/flighttracker/internal/app/sinkers/db"
	"github.com/francois-poidevin/flighttracker/internal/app/tools"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

//TODO: do real test
func TestService(t *testing.T) {
	searchSvc := New(log)
	cxt := context.Background()
	conf := db.Configuration{
		Host:     "172.17.0.2",
		Port:     5432,
		User:     "postgres",
		Password: "mysecretpassword",
		Dbname:   "postgres",
	}

	bbox := tools.Bbox{
		LatSW: 43.52,
		LonSW: 1.32,
		LatNE: 43.70,
		LonNE: 1.69,
	}

	from := time.Date(2021, 07, 22, 9, 00, 00, 651387237, time.UTC)
	to := time.Date(2021, 07, 22, 12, 00, 00, 651387237, time.UTC)

	_, errSearch := searchSvc.Search(cxt, conf, bbox, 500, from, to)

	if errSearch != nil {
		log.Error(errSearch)
	}
}

func init() {

	//log handling
	log = logrus.New()
	// log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)                     //default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	// startCmd.Flags().StringVar(&cfgFile, "config", "config_flighttracker.toml", "config file")
}
