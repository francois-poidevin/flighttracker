# Flight Tracker

This app (under construction ...) is to track all aircrafts that passing onto a specified Bouding Box, store it, and alert an user if the flight doesn't respect flight rules

## Use cases
- You think to buy a house, and you want to prevently check if your investissement will be under hight flight pressure
- You want to know if a specific area have illegal flights (under altitude rules)

## Informations
### API feed

- [current use] https://www.flightradar24.com (https://data-live.flightradar24.com/zones/fcgi/feed.js?bounds=43.79,43.53,1.23,2.03&faa=1&satellite=1&mlat=1&flarm=1&adsb=1&gnd=1&air=1&vehicles=1&estimated=1&maxage=14400&gliders=1&stats=1)
- [in study] https://www.adsbexchange.com/data
- [in study] https://opensky-network.org/

### Bbox construction
If you need to construct a bbox that fit with FlightTracker requierement, take a look in [bboxfinder.com](http://bboxfinder.com)

## Build
```bash
go build -o bin/flighttracker
```

or

```bash
docker build -t flighttracker:latest .
```

## Configuration file
### Generate
Generate a configuration file on local folder with default values
```bash
./bin/flighttracker config new > config_flighttracker.toml
```
You can change the configuration file for specifying configuration

```toml
###############################
# Flighttracker Settings 
###############################
[Flighttracker]

  # tracking bbox
  bbox = "43.52,1.32^43.70,1.69"

  # refresh timing
  refresh = 5

  # the sinker Type use
  sinkertype = "DB"

  ###############################
  # db sinker configuration 
  ###############################
  [Flighttracker.db]

    # Postgres dbName
    dbName = "postgres"

    # Postgres host
    host = "172.17.0.2"

    # Postgres password
    password = "mysecretpassword"

    # Postgres port
    port = 5432

    # Postgres user
    user = "postgres"

  ###############################
  # file sinker configuration 
  ###############################
  [Flighttracker.file]

    # output raw file name
    outputraw = "rawData.log"

    # output report file name
    outputreport = "report.log"

###############################
# Logs Settings 
###############################
[Log]

  # Log level: debug, info, warn, error, dpanic, panic, and fatal
  level = "warn"
```
### Explaination

| Parameter        	| Signification           			|
| ------------- 	|---------------|
| Flighttracker.refresh			| Refresh timer (every n seconds)	|
| Flighttracker.bbox				| BoundingBox where analyse is done (Bottom Left-Top Right)	|
| Flighttracker.sinkerType				| Sinker type (STDOUT or FILE or DB)	|
| Flighttracker.db.dbName				| Postgres Database Name	|
| Flighttracker.db.host			        | Postgres Database host	|
| Flighttracker.db.password				| Postgres Database password	|
| Flighttracker.db.port				    | Postgres Database port	|
| Flighttracker.db.user				    | Postgres Database user	|
| Flighttracker.file.outputraw			| File name for output raw for sinker type 'FILE' 	|
| Flighttracker.file.outputreport		| File name for output report for sinker type 'FILE'	|
| Log		| Log level used	|

### sinkerType
#### STDOUT
The sinker will display on Standard Output the raw data unmarshalled and unmarshalled data with criteria (flight under ...) 
#### FILE
This sinker will create files on local folder where the application is running under 'log' folder. The _rawData.log_ file store all unmarshalled data from raw json. The _report.log_ file store only unmarshalled data with criteria (flight under ...)
#### DB
##### Pre-requisite

```
docker run -p 5432:5432 --name some-postgis -e POSTGRES_PASSWORD=mysecretpassword -d postgis/postgis
```

This sinker will create a database structure in postgres database (schema and table)
This sinker will store inbound data to postgres database

## Run
Running flightTracker with default information of configuration file
```bash
./bin/flighttracker start --config config_flighttracker.toml
```
or

```bash
docker run flighttracker start --config config_flighttracker.toml
```

## Docker images

- Storing data

[Postgis dockerHub image](https://hub.docker.com/r/postgis/postgis/)
```
FROM postgis/postgis
```

- worker

[Golang dockerHub image](https://hub.docker.com/_/golang)
```
FROM golang:alpine
```

## Docker compose - not yet implemented

## Airport area
For excluding illegal flights onto airport area, we will use overpass QL to request OSM API to generate polygon of each airport in France
### Overpass
Play the overpass QL script on [Overpass-turbo](https://overpass-turbo.eu/)
### Overpass QL
```json
[out:json][timeout:600];
// gather results
(
  // query part for: “aeroway=aerodrome”
  node["aeroway"="aerodrome"]({{bbox}});
  way["aeroway"="aerodrome"]({{bbox}});
  relation["aeroway"="aerodrome"]({{bbox}});
);
// print results
out body;
>;
out skel qt;
```

### osmtogeojson
To generate OSM data to geojson
[osmtogeojson](https://tyrasd.github.io/osmtogeojson/)



