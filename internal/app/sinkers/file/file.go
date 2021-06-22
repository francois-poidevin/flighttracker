package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/francois-poidevin/flighttracker/internal/app"
	"github.com/sirupsen/logrus"
)

//TODO: close the files when app close
type FileSinker struct {
	Log             *logrus.Logger
	fIllegalFlights *os.File
	fAllFlights     *os.File
}

func New(log *logrus.Logger) app.Sinker {
	//init the logger here
	return &FileSinker{Log: log}
}

func (s *FileSinker) Init(ctx context.Context, params interface{}) error {
	parameters := params.(Configuration)
	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"Outputraw":    parameters.Outputraw,
		"Outputreport": parameters.Outputreport,
	}).Info("Initialisation File sinker Parameters")

	logFolder := "log"
	errCreateFolder := makeDirectoryIfNotExists(logFolder)
	if errCreateFolder != nil {
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"Error":  errCreateFolder,
			"folder": logFolder,
		}).Error("Unable to create folder path")
		return errCreateFolder
	}

	//create folder for files
	timestampFolderName := time.Now().Unix()
	errCreateFolder = makeDirectoryIfNotExists(filepath.Join(logFolder, strconv.FormatInt(timestampFolderName, 10)))
	if errCreateFolder != nil {
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"Error":  errCreateFolder,
			"folder": timestampFolderName,
		}).Error("Unable to create folder path")
		return errCreateFolder
	}

	fIllegalFlights, err := os.OpenFile(filepath.Join(logFolder, strconv.FormatInt(timestampFolderName, 10), parameters.Outputreport),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"Error": err,
		}).Error("Unable to Open file")
		return err
	}
	s.fIllegalFlights = fIllegalFlights
	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"Illegal Flights file": s.fIllegalFlights.Name(),
	}).Info("File successfully created")

	fAllFlights, err := os.OpenFile(filepath.Join("log", strconv.FormatInt(timestampFolderName, 10), parameters.Outputraw),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"Error": err,
		}).Error("Unable to Open file")
		return err
	}
	s.fAllFlights = fAllFlights
	s.Log.WithContext(ctx).WithFields(logrus.Fields{
		"All Flights file": s.fAllFlights.Name(),
	}).Info("File successfully created")

	return nil
}

func (s *FileSinker) Sink(ctx context.Context, t time.Time, data []app.FlightData) error {
	errAllFlights := s.storeAllFlightsOnFile(ctx, t, data)
	if errAllFlights != nil {
		return errAllFlights
	}

	errIllegalFlight := s.storeIllegalFlightOnFile(ctx, t, data)
	if errIllegalFlight != nil {
		return errIllegalFlight
	}

	return nil
}

func (s *FileSinker) storeAllFlightsOnFile(ctx context.Context, t time.Time, data []app.FlightData) error {

	if s.fAllFlights != nil {
		w := bufio.NewWriter(s.fAllFlights)

		defer func() {
			w.Flush()
		}()

		var buffer bytes.Buffer

		if len(data) > 0 {
			Marshal, err := json.Marshal(data)
			if err != nil {
				return err
			}
			s.Log.WithContext(ctx).WithFields(logrus.Fields{
				"number of Flights": len(data),
			}).Debug("========All Flights seen=============")

			buffer.Write(Marshal)
			n4, errWS := w.WriteString(t.String() + " Raw Datas\n" + buffer.String() + "\n====================================\n")
			if errWS != nil {
				return errWS
			}
			s.Log.WithContext(ctx).WithFields(logrus.Fields{
				"length": fmt.Sprintf("wrote %d bytes", n4),
			}).Debug("Wrote")

		} else {
			n4, errWS := w.WriteString(t.String() + "\n" + "No Raw data" + "\n====================================\n")
			if errWS != nil {
				return errWS
			}
			s.Log.WithContext(ctx).WithFields(logrus.Fields{
				"length": fmt.Sprintf("wrote %d bytes", n4),
			}).Debug("Wrote")
		}
	} else {
		return errors.New("No AllFlights file for storing data")
	}

	return nil
}

func (s *FileSinker) storeIllegalFlightOnFile(ctx context.Context, t time.Time, data []app.FlightData) error {

	if s.fIllegalFlights != nil {
		w := bufio.NewWriter(s.fIllegalFlights)

		defer func() {
			w.Flush()
		}()

		var buffer bytes.Buffer
		var IllegalFlight []app.FlightData
		for _, dataObj := range data {
			//found flight above 500 meters that moving
			if (float64(dataObj.Altitude)*app.FEETTOMETER) < float64(500) &&
				(float64(dataObj.Altitude)*app.FEETTOMETER) > float64(25) &&
				float64(dataObj.GroundSpeed)*app.KTSKMH > 0 {
				IllegalFlight = append(IllegalFlight, dataObj)
			}
		}

		s.Log.WithContext(ctx).WithFields(logrus.Fields{
			"number of Flights": len(IllegalFlight),
		}).Debug("========IllegalFlight Flights seen=============")

		if len(IllegalFlight) > 0 {
			Marshal, err := json.Marshal(IllegalFlight)
			if err != nil {
				return err
			}
			buffer.Write(Marshal)
			n4, errWS := w.WriteString(t.String() + " Illegal Flights\n" + buffer.String() + "\n====================================\n")
			if errWS != nil {
				return errWS
			}
			s.Log.WithContext(ctx).WithFields(logrus.Fields{
				"length": fmt.Sprintf("wrote %d bytes", n4),
			}).Debug("Wrote")
		} else {
			n4, errWS := w.WriteString(t.String() + "\n" + "No Illegal Flight" + "\n====================================\n")
			if errWS != nil {
				return errWS
			}
			s.Log.WithContext(ctx).WithFields(logrus.Fields{
				"length": fmt.Sprintf("wrote %d bytes", n4),
			}).Debug("Wrote")
		}
	} else {
		return errors.New("No IllegalFlight file for storing data")
	}

	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}
