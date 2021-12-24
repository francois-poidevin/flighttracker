# Flight Tracker

This app is to track all aircrafts that passing onto a specified Bouding Box, store it, and alert an user if the flight doesn't respect flight rules

## Use cases
- You think to buy a house, and you want to prevently check if your investissement will be under hight flight pressure
- You want to know if a specific area have illegal flights (under altitude rules)

## Informations
### French reglementation
[Guide_autorisation_survol_basses_hauteurs](https://www.ecologie.gouv.fr/sites/default/files/Guide_autorisation_survol_basses_hauteurs.pdf)
[Definition - Unité urbaine / Agglomeration](https://www.insee.fr/fr/metadonnees/definition/c1501)

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
./bin/flighttracker config new > /configlocal/config_flighttracker.toml
```
You can change the configuration file for specifying configuration

```toml
###############################
# Flighttracker Settings 
###############################
[Flighttracker]

  # tracking bbox (Lat/Lon)
  bbox = "43.52,1.32^43.70,1.69"

  # refresh timing
  refresh = 5

  # the sinker Type use
  sinkertype = "DB"

  ###############################
  # postgres sinker configuration 
  ###############################
  [Flighttracker.postgres]

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
| Flighttracker.postgres.dbName				| Postgres Database Name	|
| Flighttracker.postgres.host			        | Postgres Database host	|
| Flighttracker.postgres.password				| Postgres Database password	|
| Flighttracker.postgres.port				    | Postgres Database port	|
| Flighttracker.postgres.user				    | Postgres Database user	|
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

###### Docker
```
docker run -p 5432:5432 --name some-postgis -e POSTGRES_PASSWORD=mysecretpassword -d postgis/postgis
```

Database is accessible at <YOUR_LOCAL_DOCKER_IP>:5432

###### Local installation (Ubuntu 20.04)
 * Step 1: Update system
  ```shell
  sudo apt update
  sudo apt -y install vim bash-completion wget
  sudo apt -y upgrade
  sudo reboot
  ```
 * Step 2: Add PostgreSQL 12 repository
  ```shell
  wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
  echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" |sudo tee  /etc/apt/sources.list.d/pgdg.list
  ```
 * Step 3: Install PostgreSQL 12
  ```shell
  sudo apt update
  sudo apt -y install postgresql-12 postgresql-client-12
  ```
 * Step 4: Install Postgis extension
  ```shell
  sudo su - postgres
  psql -d postgres
  postgres=# CREATE EXTENSION postgis;
  ```

Database is accessible at 127.0.0.1:5432

##### Informations
This sinker will create a database structure in postgres database (schema and table)
This sinker will store inbound data to postgres database

## Run
### start service

The _start_ CLI service allow to sink data in differents sinkers has: STDOUT, FILE, DB (POSTGRES)

Running flightTracker with default information of configuration file
```bash
./bin/flighttracker start --config ./configlocal/config_flighttracker.toml
```
or

```bash
docker run -v /home/poidevin/Projects/flighttracker/configlocal:/tmp flighttracker:latest start --config /tmp/config_flighttracker.toml
```
then call the endpoint on docker container IP (i.e. 172.17.0.1:8080)

### startHttp service
The _startHttp_ CLI service allow start a REST server on 8080 port to sink data in database (POSTGRES) and allow to request to the database

Running flightTracker with default information of configuration file
```bash
./bin/flighttracker startHttp --config ./configlocal/config_flighttracker.toml
```
or

```bash
docker run -v /home/poidevin/Projects/flighttracker/configlocal:/tmp -p 8080:8080 flighttracker:latest startHttp --config /tmp/config_flighttracker.toml
```
then call the endpoint on docker container IP (i.e. 172.17.0.1:8080)

#### endpoints

| endpoint        	| HTTP Methods           			| example           			|signification           			|
| ------------- 	|---------------   |---------------   |   ---------------|
| /start | GET | localhost:8080/api/v1/start | to start the sinking service on database |
| /stop | GET | localhost:8080/api/v1/stop | to stop the sinking service on database |
| /search | GET | localhost:8080/api/v1/search?bbox=43.52,1.32^43.70,1.69&altThresholdFeet=500&fromTimeStamp=2021-07-22T09:00:00&toTimeStamp=2021-07-24T12:00:00 | to search data from database on several criteria as path parameters |

##### start
To start the sinking service on database

##### stop
To stop the sinking service on database

##### search
| path parameters        	| signification           			|
|-----------------------  |------------------------------|
|         bbox                |  BoundingBox where analyse is done (Bottom Left-Top Right)                             |
|       altThresholdFeet      |   floor threshold for research in Feet unit. Only above and equals data will be returned                         |
|     fromTimeStamp           |  from time windows for search                             |
|     toTimeStamp             |  to time windows for search                             |

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



