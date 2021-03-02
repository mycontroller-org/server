package settings

import (
	"errors"
	"fmt"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/settings"
	settingsML "github.com/mycontroller-org/backend/v2/pkg/model/settings"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]settingsML.Settings, 0)
	return stg.SVC.Find(ml.EntitySettings, &result, filters, pagination)
}

// Save a setting details
func Save(settings *settingsML.Settings) error {
	if settings.ID == "" {
		return errors.New("ID should not be nil")
	}
	filters := []stgML.Filter{
		{Key: ml.KeyID, Value: settings.ID},
	}
	return stg.SVC.Upsert(ml.EntitySettings, settings, filters)
}

// GetByID returns a item
func GetByID(ID string) (*settingsML.Settings, error) {
	result := &settingsML.Settings{}
	filters := []stgML.Filter{
		{Key: ml.KeyID, Operator: stgML.OperatorEqual, Value: ID},
	}
	err := stg.SVC.FindOne(ml.EntitySettings, result, filters)
	return result, err
}

// UpdateSettings config into disk
func UpdateSettings(settings *settingsML.Settings) error {
	if settings.ID == "" {
		return errors.New("id cannot be empty")
	}

	switch settings.ID {
	case settingsML.KeySystemSettings:
		return UpdateSystemSettings(settings)

	case settingsML.KeySystemJobs, settingsML.KeyVersion:
		settings.Modified = time.Now()
		return update(settings)

	default:
		return fmt.Errorf("unknown settings id:%s", settings.ID)
	}

}

// GetSystemJobs details
func GetSystemJobs() (*settingsML.SystemJobsSettings, error) {
	settings, err := GetByID(settings.KeySystemJobs)
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
	settings.Modified = time.Now()
	// TODO: verify required fields
	return update(settings)
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
	return update(settings)
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

// update is a common function to update a document in settings entity
func update(settings *settingsML.Settings) error {
	filters := []stgML.Filter{
		{Key: ml.KeyID, Value: settings.ID},
	}
	settings.Modified = time.Now()
	return stg.SVC.Upsert(ml.EntitySettings, settings, filters)
}
