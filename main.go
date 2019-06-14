package main

import (
	"fmt"

	worker "github.com/francois-poidevin/flighttracker/src"
)

func main() {

	fmt.Println("Start")

	worker.Execute()

	fmt.Println("End")
}
