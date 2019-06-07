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
	"time"
)

//FlightData - storage structure for flightRadar24 API response
type FlightData struct {
	flightID         int
	iCAO24BITADDRESS string
	lat              float32
	lon              float32
	track            int //degree to the destination
	altitude         int
	groundSpeed      int    //kts
	unknown1         string //not describe yet
	transpondeurType string
	aircraftType     string
	immatriculation1 string
	timeStamp        uint
	origine          string
	destination      string
	unknown2         string
	verticalSpeed    int
	immatriculation2 string
	unknown3         int
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

	unMarshalByte(body)

	// fmt.Println(string(body))

	n4, errWS := w.WriteString(t.String() + "\n" + string(body) + "\n====================================\n")
	if errWS != nil {
		panic(errWS)
	}
	w.Flush()
	fmt.Printf("wrote %d bytes\n", n4)
}

func unMarshalByte(byt []byte) []FlightData {

	var data map[string]interface{}

	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}

	for k, v := range data {
		if k != "full_count" && k != "version" && k != "stats" {
			fmt.Println("key: " + k)
			fmt.Println(fmt.Sprintf("value: %v", v))
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				fmt.Println("it's a slice !!!")
			}
		}
	}

	// fmt.Println(data)

	return nil
}
