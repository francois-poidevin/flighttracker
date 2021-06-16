package config

import (
	"github.com/francois-poidevin/flighttracker/internal/app/sinkers/db"
	"github.com/francois-poidevin/flighttracker/internal/app/sinkers/file"
)

// Configuration contains conectivity settings
type Configuration struct {
	Log struct {
		Level string `toml:"level" default:"warn" comment:"Log level: debug, info, warn, error, dpanic, panic, and fatal"`
	} `toml:"Log" comment:"###############################\n Logs Settings \n##############################"`

	Flighttracker struct {
		Bbox       string             `toml:"bbox" default:"43.52,1.32^43.70,1.69" comment:"tracking bbox"`
		Refresh    int                `toml:"refresh" default:"5" comment:"refresh timing"`
		Sinkertype string             `toml:"sinkertype" default:"FILE" comment:"the sinker Type use"`
		File       file.Configuration `toml:"file" comment:"###############################\n file sinker configuration \n##############################"`
		Postgres   db.Configuration   `toml:"db" comment:"###############################\n db sinker configuration \n##############################"`
	} `toml:"Flighttracker" comment:"###############################\n Flighttracker Settings \n##############################"`
}
