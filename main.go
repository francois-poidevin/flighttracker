package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("Start")
	d := 5 * time.Second
	for x := range time.Tick(d) {
		storeDataOnFile(x)
	}
	fmt.Println("End")
}

func storeDataOnFile(t time.Time) {
	f, err := os.OpenFile("data.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	//Made the HTTP request
	resp, errHTTPGet := http.Get("https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=43.79,43.53,1.23,2.03&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1")
	if errHTTPGet != nil {
		panic(errHTTPGet)
	}
	defer resp.Body.Close()

	body, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		panic(errRead)
	}

	fmt.Println(string(body))

	n4, errWS := w.WriteString(t.String() + "\n" + string(body) + "\n====================================\n")
	if errWS != nil {
		panic(errWS)
	}
	w.Flush()
	fmt.Printf("wrote %d bytes\n", n4)
}
