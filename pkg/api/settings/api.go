package settings

import (
	"context"
	"errors"
	"fmt"
	"time"

	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busutils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type SettingsAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
	enc     *encryptionAPI.Encryption
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) *SettingsAPI {
	return &SettingsAPI{
		ctx:     ctx,
		logger:  logger.Named("settings_api"),
		storage: storage,
		bus:     bus,
		enc:     enc,
	}
}

// List by filter and pagination
func (s *SettingsAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]settingsTY.Settings, 0)
	return s.storage.Find(types.EntitySettings, &result, filters, pagination)
}

// Save a setting details
func (s *SettingsAPI) Save(settings *settingsTY.Settings) error {
	if settings.ID == "" {
		return errors.New("id should not be nil")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: settings.ID},
	}
	return s.storage.Upsert(types.EntitySettings, settings, filters)
}

// GetByID returns a item
func (s *SettingsAPI) GetByID(ID string) (*settingsTY.Settings, error) {
	result := &settingsTY.Settings{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: ID},
	}
	err := s.storage.FindOne(types.EntitySettings, result, filters)
	if err != nil {
		return nil, err
	}

	// convert map data to struct
	var specStruct interface{}
	switch result.ID {
	case settingsTY.KeySystemSettings:
		settings := settingsTY.SystemSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsTY.KeySystemJobs:
		settings := settingsTY.SystemJobsSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsTY.KeySystemBackupLocations:
		exportLocations := settingsTY.BackupLocations{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &exportLocations)
		if err != nil {
			return nil, err
		}
		specStruct = exportLocations

	case settingsTY.KeySystemDynamicSecrets:
		settings := settingsTY.SystemDynamicSecrets{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	default:

	}

	// struct to json then json to map
	// this workaround to apply json tags
	if specStruct != nil {
		bytes, err := json.Marshal(specStruct)
		if err != nil {
			return nil, err
		}

		mapSpec := make(map[string]interface{})
		err = json.Unmarshal(bytes, &mapSpec)
		if err != nil {
			return nil, err
		}
		result.Spec = mapSpec
	}

	return result, nil
}

// UpdateSettings config into disk
func (s *SettingsAPI) UpdateSettings(settings *settingsTY.Settings) error {
	if settings.ID == "" {
		return errors.New("id cannot be empty")
	}

	switch settings.ID {
	case settingsTY.KeySystemSettings:
		return s.UpdateSystemSettings(settings)

	case settingsTY.KeySystemJobs,
		settingsTY.KeyVersion,
		settingsTY.KeySystemBackupLocations,
		settingsTY.KeyTelemetry,
		settingsTY.KeySystemDynamicSecrets:
		settings.ModifiedOn = time.Now()

		return s.update(settings)

	default:
		return fmt.Errorf("unknown settings id:%s", settings.ID)
	}

}

// GetSystemJobs details
func (s *SettingsAPI) GetSystemJobs() (*settingsTY.SystemJobsSettings, error) {
	settings, err := s.GetByID(settingsTY.KeySystemJobs)
	if err != nil {
		return nil, err
	}
	systemJobs := &settingsTY.SystemJobsSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemJobs)
	if err != nil {
		return nil, err
	}
	return systemJobs, nil
}

// UpdateSystemSettings config into disk
func (s *SettingsAPI) UpdateSystemSettings(settings *settingsTY.Settings) error {
	settings.ID = settingsTY.KeySystemSettings
	settings.ModifiedOn = time.Now()

	// TODO: verify required fields
	err := s.update(settings)
	if err != nil {
		return err
	}
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}
	if systemSettings.GeoLocation.AutoUpdate {
		err = s.AutoUpdateSystemGEOLocation()
		if err != nil {
			return err
		}
	}
	// send geo location updated event

	// post node state updater job change event
	busutils.PostServiceEvent(s.logger, s.bus, topic.TopicInternalSystemJobs, rsTY.TypeSystemJobs, rsTY.CommandReload, rsTY.SubCommandJobNodeStatusUpdater)

	return nil
}

// UpdateGeoLocation updates the location details
func (s *SettingsAPI) UpdateGeoLocation(location *settingsTY.GeoLocation) error {
	settings, err := s.GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return err
	}

	// convert spec to system settings
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}

	// update location
	systemSettings.GeoLocation = *location
	settings.Spec = utils.StructToMap(systemSettings)
	err = s.update(settings)
	if err != nil {
		return err
	}

	// send geo location updated event
	return nil
}

// GetGeoLocation returns configured latitude and longitude settings to calculate sunrise and sunset
func (s *SettingsAPI) GetGeoLocation() (*settingsTY.GeoLocation, error) {
	settings, err := s.GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to system settings
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	sunrise := systemSettings.GeoLocation
	return &sunrise, nil
}

// GetBackupLocations returns locations set by user
func (s *SettingsAPI) GetBackupLocations() (*settingsTY.BackupLocations, error) {
	settings, err := s.GetByID(settingsTY.KeySystemBackupLocations)
	if err != nil {
		return nil, err
	}

	// convert spec to BackupLocations
	systemSettings := &settingsTY.BackupLocations{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}

// update is a common function to update a document in settings entity
func (s *SettingsAPI) update(settings *settingsTY.Settings) error {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: settings.ID},
	}
	settings.ModifiedOn = time.Now()

	// encrypt passwords, tokens, etc
	err := s.enc.EncryptSecrets(settings)
	if err != nil {
		s.logger.Error("error on encryption", zap.Error(err))
		return err
	}

	return s.storage.Upsert(types.EntitySettings, settings, filters)
}

// returns telemetry config data
func (s *SettingsAPI) GetTelemetry() (*settingsTY.TelemetryConfig, error) {
	settings, err := s.GetByID(settingsTY.KeyTelemetry)
	if err != nil {
		return nil, err
	}

	// convert spec to telemetry config data
	telemetryConfig := &settingsTY.TelemetryConfig{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, telemetryConfig)
	if err != nil {
		return nil, err
	}

	return telemetryConfig, nil
}

// GetSystemSettings returns system settings data
func (s *SettingsAPI) GetSystemSettings() (*settingsTY.SystemSettings, error) {
	settings, err := s.GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to telemetry config data
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}

func (s *SettingsAPI) ResetJwtAccessSecret(newSecret string) error {
	if newSecret == "" {
		newSecret = utils.RandUUID()
	}

	systemSecrets := &settingsTY.SystemDynamicSecrets{
		JwtAccessSecret: newSecret,
	}
	spec := utils.StructToMap(systemSecrets)

	settings := &settingsTY.Settings{
		ID:   settingsTY.KeySystemDynamicSecrets,
		Spec: spec,
	}
	err := s.UpdateSettings(settings)
	if err != nil {
		return err
	}
	return s.UpdateJwtAccessSecret()
}

func (s *SettingsAPI) UpdateJwtAccessSecret() error {
	settings, err := s.GetByID(settingsTY.KeySystemDynamicSecrets)
	if err != nil {
		if err != storageTY.ErrNoDocuments {
			return err
		}
		settings = &settingsTY.Settings{}
	}

	// decrypt passwords, tokens, etc
	err = s.enc.DecryptSecrets(settings)
	if err != nil {
		s.logger.Error("error on decryption", zap.Error(err))
		return err
	}

	systemSecret := &settingsTY.SystemDynamicSecrets{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSecret)
	if err != nil {
		return err
	}

	if systemSecret.JwtAccessSecret == "" {
		systemSecret.JwtAccessSecret = utils.RandUUID()
		err = s.ResetJwtAccessSecret(systemSecret.JwtAccessSecret)
		if err != nil {
			return err
		}
	}

	return types.SetEnv(types.ENV_JWT_ACCESS_SECRET, systemSecret.JwtAccessSecret)
}

func (s *SettingsAPI) Import(data interface{}) error {
	input, ok := data.(settingsTY.Settings)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return s.storage.Upsert(types.EntitySettings, &input, filters)
}

func (s *SettingsAPI) GetEntityInterface() interface{} {
	return settingsTY.Settings{}
}
