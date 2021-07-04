package settings

import (
	"errors"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/model"
	settingsML "github.com/mycontroller-org/server/v2/pkg/model/settings"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	stgSVC "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]settingsML.Settings, 0)
	return stgSVC.SVC.Find(model.EntitySettings, &result, filters, pagination)
}

// Save a setting details
func Save(settings *settingsML.Settings) error {
	if settings.ID == "" {
		return errors.New("id should not be nil")
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: settings.ID},
	}
	return stgSVC.SVC.Upsert(model.EntitySettings, settings, filters)
}

// GetByID returns a item
func GetByID(ID string) (*settingsML.Settings, error) {
	result := &settingsML.Settings{}
	filters := []stgType.Filter{
		{Key: model.KeyID, Operator: stgType.OperatorEqual, Value: ID},
	}
	err := stgSVC.SVC.FindOne(model.EntitySettings, result, filters)
	if err != nil {
		return nil, err
	}

	// convert map data to struct
	var specStruct interface{}
	switch result.ID {
	case settingsML.KeySystemSettings:
		settings := settingsML.SystemSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsML.KeySystemJobs:
		settings := settingsML.SystemJobsSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsML.KeySystemBackupLocations:
		exportLocations := settingsML.BackupLocations{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &exportLocations)
		if err != nil {
			return nil, err
		}
		specStruct = exportLocations

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
func UpdateSettings(settings *settingsML.Settings) error {
	if settings.ID == "" {
		return errors.New("id cannot be empty")
	}

	switch settings.ID {
	case settingsML.KeySystemSettings:
		return UpdateSystemSettings(settings)

	case settingsML.KeySystemJobs, settingsML.KeyVersion, settingsML.KeySystemBackupLocations, settingsML.KeyAnalytics:
		if !configuration.PauseModifiedOnUpdate.IsSet() {
			settings.ModifiedOn = time.Now()
		}
		return update(settings)

	default:
		return fmt.Errorf("unknown settings id:%s", settings.ID)
	}

}

// GetSystemJobs details
func GetSystemJobs() (*settingsML.SystemJobsSettings, error) {
	settings, err := GetByID(settingsML.KeySystemJobs)
	if err != nil {
		return nil, err
	}
	systemJobs := &settingsML.SystemJobsSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemJobs)
	if err != nil {
		return nil, err
	}
	return systemJobs, nil
}

// UpdateSystemSettings config into disk
func UpdateSystemSettings(settings *settingsML.Settings) error {
	settings.ID = settingsML.KeySystemSettings
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		settings.ModifiedOn = time.Now()
	}

	// TODO: verify required fields
	err := update(settings)
	if err != nil {
		return err
	}
	systemSettings := &settingsML.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}
	if systemSettings.GeoLocation.AutoUpdate {
		err = AutoUpdateSystemGEOLocation()
		if err != nil {
			return err
		}
	}
	// send geo location updated event

	return nil
}

// UpdateGeoLocation updates the location details
func UpdateGeoLocation(location *settingsML.GeoLocation) error {
	settings, err := GetByID(settingsML.KeySystemSettings)
	if err != nil {
		return err
	}

	// convert spec to system settings
	systemSettings := &settingsML.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}

	// update location
	systemSettings.GeoLocation = *location
	settings.Spec = utils.StructToMap(systemSettings)
	err = update(settings)
	if err != nil {
		return err
	}

	// send geo location updated event
	return nil
}

// GetGeoLocation returns configured latitude and longitude settings to calculate sunrise and sunset
func GetGeoLocation() (*settingsML.GeoLocation, error) {
	settings, err := GetByID(settingsML.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to system settings
	systemSettings := &settingsML.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	sunrise := systemSettings.GeoLocation
	return &sunrise, nil
}

// GetBackupLocations returns locations set by user
func GetBackupLocations() (*settingsML.BackupLocations, error) {
	settings, err := GetByID(settingsML.KeySystemBackupLocations)
	if err != nil {
		return nil, err
	}

	// convert spec to BackupLocations
	systemSettings := &settingsML.BackupLocations{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}

// update is a common function to update a document in settings entity
func update(settings *settingsML.Settings) error {
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: settings.ID},
	}
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		settings.ModifiedOn = time.Now()
	}
	return stgSVC.SVC.Upsert(model.EntitySettings, settings, filters)
}

// GetAnalytics returns analytics data
func GetAnalytics() (*settingsML.AnalyticsConfig, error) {
	settings, err := GetByID(settingsML.KeyAnalytics)
	if err != nil {
		return nil, err
	}

	// convert spec to analytics data
	analyticsData := &settingsML.AnalyticsConfig{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, analyticsData)
	if err != nil {
		return nil, err
	}

	return analyticsData, nil
}

// GetSystemSettings returns system settings data
func GetSystemSettings() (*settingsML.SystemSettings, error) {
	settings, err := GetByID(settingsML.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to analytics data
	systemSettings := &settingsML.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}
