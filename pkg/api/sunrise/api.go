package sunrise

import (
	"context"
	"time"

	"github.com/btittelbach/astrotime"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type SunriseAPI struct {
	ctx         context.Context
	logger      *zap.Logger
	settingsAPI *settingsAPI.SettingsAPI
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) types.Sunrise {
	_settingsAPI := settingsAPI.New(ctx, logger, storage, enc, bus)
	return &SunriseAPI{
		ctx:         ctx,
		logger:      logger.Named("sunrise_api"),
		settingsAPI: _settingsAPI,
	}
}

// returns sunrise time
func (sr *SunriseAPI) SunriseTime() (*time.Time, error) {
	// get location settings
	location, err := sr.settingsAPI.GetGeoLocation()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sunrise := astrotime.CalcSunrise(now, location.Latitude, location.Longitude)
	return &sunrise, nil
}

// returns sunset time
func (sr *SunriseAPI) SunsetTime() (*time.Time, error) {
	// get location settings
	location, err := sr.settingsAPI.GetGeoLocation()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sunset := astrotime.CalcSunset(now, location.Latitude, location.Longitude)
	return &sunset, nil
}
