package file

// Configuration settings for file sinking
type Configuration struct {
	Outputraw    string `toml:"outputraw" default:"rawData.log" comment:"output raw file name"`
	Outputreport string `toml:"outputreport" default:"report.log" comment:"output report file name"`
}
