package config

// Configuration contains conectivity settings
type Configuration struct {
	Log struct {
		Level string `toml:"level" default:"warn" comment:"Log level: debug, info, warn, error, dpanic, panic, and fatal"`
	} `toml:"Log" comment:"###############################\n Logs Settings \n##############################"`

	Flighttracker struct {
		Bbox         string `toml:"bbox" default:"43.52,1.32^43.70,1.69" comment:"tracking bbox"`
		Refresh      int    `toml:"refresh" default:"5" comment:"refresh timing"`
		Outputraw    string `toml:"outputraw" default:"rawData.log" comment:"output raw file name"`
		Outputreport string `toml:"outputreport" default:"report.log" comment:"output report file name"`
	} `toml:"Flighttracker" comment:"###############################\n Flighttracker Settings \n##############################"`
}
