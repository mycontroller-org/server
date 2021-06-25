package sunrise

import (
	"time"

	"github.com/btittelbach/astrotime"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
)

// GetSunriseTime returns sunrise time
func GetSunriseTime() (*time.Time, error) {
	// get location settings
	location, err := settingsAPI.GetGeoLocation()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sunrise := astrotime.CalcSunrise(now, location.Latitude, location.Longitude)
	return &sunrise, nil
}

// GetSunsetTime returns sunset time
func GetSunsetTime() (*time.Time, error) {
	// get location settings
	location, err := settingsAPI.GetGeoLocation()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sunset := astrotime.CalcSunset(now, location.Latitude, location.Longitude)
	return &sunset, nil
}
