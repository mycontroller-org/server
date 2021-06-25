package settings

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"go.uber.org/zap"
)

const geoLocationURL = "https://ipinfo.io/json"

type GeoLocationAPIResponse struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Location string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

func GetLocation() (*GeoLocationAPIResponse, error) {
	response, err := http.Get(geoLocationURL)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var geoData GeoLocationAPIResponse
	err = json.Unmarshal(body, &geoData)
	if err != nil {
		return nil, err
	}

	return &geoData, nil
}

// AutoUpdateSystemGEOLocation updates system location based on the ip
func AutoUpdateSystemGEOLocation() error {
	// get location details
	location, err := GetGeoLocation()
	if err != nil {
		return err
	}

	if !location.AutoUpdate {
		return nil
	}

	geoData, err := GetLocation()
	if err != nil {
		return err
	}

	zap.L().Debug("detected geo details", zap.Any("geoData", geoData))

	tokens := strings.Split(geoData.Location, ",")
	if len(geoData.Location) == 0 || len(tokens) != 2 {
		return errors.New("error on detecting geo location")
	}

	location.LocationName = geoData.City
	location.Latitude, _ = strconv.ParseFloat(tokens[0], 64)
	location.Longitude, _ = strconv.ParseFloat(tokens[1], 64)

	zap.L().Info("location data to be updated", zap.Any("location", location))

	return UpdateGeoLocation(location)
}
