package tools

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Bbox - a bounding box structure
type Bbox struct {
	LatSW float64
	LonSW float64
	LatNE float64
	LonNE float64
}

func GetBbox(data string) (Bbox, error) {
	sWnE := strings.Split(data, "^")
	result := Bbox{}
	if len(sWnE) != 2 {
		return result, errors.New("Bounding Box malformed - need ^ for separating SW and NE coordinate")
	}

	for idx, latlonRec := range sWnE {
		latlon := strings.Split(latlonRec, ",")
		if len(latlon) != 2 {
			return result, errors.New("Bounding Box malformed - need , for separating lat and lon coordinate")
		}
		lat, errLat := strconv.ParseFloat(latlon[0], 64)
		if errLat != nil {
			return result, errLat
		}
		lon, errLon := strconv.ParseFloat(latlon[1], 64)
		if errLon != nil {
			return result, errLon
		}
		if idx == 0 {
			result.LatSW = lat
			result.LonSW = lon
		} else {

			result.LatNE = lat
			result.LonNE = lon
		}
	}
	return result, nil
}

func BboxToWKT(bbox Bbox) string {
	sw := fmt.Sprintf("%f %f", bbox.LonSW, bbox.LatSW)
	nw := fmt.Sprintf("%f %f", bbox.LonSW, bbox.LatNE)
	ne := fmt.Sprintf("%f %f", bbox.LonNE, bbox.LatNE)
	se := fmt.Sprintf("%f %f", bbox.LonNE, bbox.LatSW)
	result := fmt.Sprintf("POLYGON((%s, %s, %s, %s, %s))", sw, nw, ne, se, sw)
	return result
}
