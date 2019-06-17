package src

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
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
	timeStamp        uint64
	origine          string
	destination      string
	unknown2         string
	verticalSpeed    int64
	immatriculation2 string
	unknown3         int64
	company          string
}

//Execute - start the worker
func Execute() {
	//Loop each 5 secondes for working
	d := 5 * time.Second
	for x := range time.Tick(d) {
		storeDataOnFile(x)
	}
}

func storeDataOnFile(t time.Time) {
	f, err := os.OpenFile("data.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	// Made the HTTP request - Test area 43.663712,1.570358,43.710510,1.700735
	// Toulouse and Airport Aera - 43.515693,1.318359,43.702630,1.687775
	resp, errHTTPGet := http.Get("https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=43.70,43.51,1.31,1.68&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1")
	if errHTTPGet != nil {
		panic(errHTTPGet)
	}
	defer resp.Body.Close()

	body, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		panic(errRead)
	}

	data := unMarshalByte(body)

	for i := 0; i < len(data); i++ {
		fmt.Println("aircraftType " + data[i].aircraftType)
		fmt.Println("immatriculation1 " + data[i].immatriculation1)
		fmt.Println("origine " + data[i].origine)
		fmt.Println("destination " + data[i].destination)
		fmt.Println("Altitude " + fmt.Sprintf("%v", data[i].altitude))
	}

	n4, errWS := w.WriteString(t.String() + "\n" + string(body) + "\n====================================\n")
	if errWS != nil {
		panic(errWS)
	}
	w.Flush()
	fmt.Printf("wrote %d bytes\n", n4)
}

func unMarshalByte(byt []byte) []FlightData {

	var data map[string]interface{}
	var result []FlightData
	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}

	for k, v := range data {
		if k != "full_count" && k != "version" && k != "stats" {
			fmt.Println("key: " + k)
			fmt.Println(fmt.Sprintf("value: %v", v))
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				s := reflect.ValueOf(v)
				_lat, _ := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(1)), 64)
				_lon, _ := strconv.ParseFloat(fmt.Sprintf("%v", s.Index(2)), 64)
				_track, _ := strconv.ParseInt(fmt.Sprintf("%v", s.Index(3)), 10, 64)
				_altitude, _ := strconv.ParseInt(fmt.Sprintf("%v", s.Index(4)), 10, 64)
				_groundSpeed, _ := strconv.ParseInt(fmt.Sprintf("%v", s.Index(5)), 10, 64)

				_timeStamp, _ := strconv.ParseUint(fmt.Sprintf("%v", s.Index(10)), 10, 64)

				_verticalSpeed, _ := strconv.ParseInt(fmt.Sprintf("%v", s.Index(14)), 10, 64)

				_unknown3, _ := strconv.ParseInt(fmt.Sprintf("%v", s.Index(16)), 10, 64)

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
					unknown3:         _unknown3,
					company:          fmt.Sprintf("%v", s.Index(17)),
				}
				result = append(result, flightData)
			}
		}
	}

	// fmt.Println(result)

	return result
}
