package db

// Configuration settings for Postgres DB sinking
type Configuration struct {
	Host     string `toml:"host" default:"172.17.0.2" comment:"Postgres host"`
	Port     int    `toml:"port" default:"5432" comment:"Postgres port"`
	User     string `toml:"user" default:"postgres" comment:"Postgres user"`
	Password string `toml:"password" default:"mysecretpassword" comment:"Postgres password"`
	Dbname   string `toml:"dbName" default:"postgres" comment:"Postgres dbName"`
}
