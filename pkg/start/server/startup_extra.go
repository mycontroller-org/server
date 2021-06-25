package server

import (
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	settingsML "github.com/mycontroller-org/server/v2/pkg/model/settings"
	userML "github.com/mycontroller-org/server/v2/pkg/model/user"
	systemJobs "github.com/mycontroller-org/server/v2/pkg/service/system_jobs"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	"github.com/mycontroller-org/server/v2/pkg/version"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
	"go.uber.org/zap"
)

const (
	loginMessage = "Default username and password to login: <b>admin</b> / <b>admin</b>"
)

// StartupJobsExtra func
func StartupJobsExtra() {
	CreateSettingsData()
	UpdateInitialUser()
	UpdateGeoLocation()
	systemJobs.ReloadSystemJobs()
}

// UpdateInitialUser func
func UpdateInitialUser() {
	pagination := &stgML.Pagination{
		Limit: 1,
	}
	users, err := userAPI.List(nil, pagination)
	if err != nil {
		zap.L().Error("failed to list users", zap.Error(err))
	}
	if users.Count == 0 {
		hashedPassword, err := hashed.GenerateHash("admin")
		if err != nil {
			zap.L().Fatal("unable to get hashed password", zap.Error(err))
			return
		}
		adminUser := &userML.User{
			Username: "admin",
			Password: hashedPassword,
			FullName: "Admin User",
			Email:    "admin@example.com",
		}
		err = userAPI.Save(adminUser)
		if err != nil {
			zap.L().Error("failed to create default admin user", zap.Error(err))
		}
	}
}

// CreateSettingsData if non available
func CreateSettingsData() {
	// update system settings data
	_, err := settingsAPI.GetByID(settingsML.KeySystemSettings)
	if err == nil {
		return
	}
	zap.L().Info("error on fetching system settings, assuming it is fresh install and populating default details. if not, report this issue", zap.String("error", err.Error()))

	// update system settings
	systemSettings := settingsML.SystemSettings{
		GeoLocation: settingsML.GeoLocation{AutoUpdate: true},
		Login:       settingsML.Login{Message: loginMessage},
	}
	settings := &settingsML.Settings{ID: settingsML.KeySystemSettings}
	settings.Spec = utils.StructToMap(systemSettings)
	err = settingsAPI.UpdateSystemSettings(settings)
	if err != nil {
		zap.L().Fatal("error on updating system settings", zap.Error(err))
	}

	// update version details
	versionSettings := &settingsML.Settings{ID: settingsML.KeyVersion}
	versionData := settingsML.VersionSettings{Version: version.Get().Version}
	versionSettings.Spec = utils.StructToMap(versionData)
	err = settingsAPI.UpdateSettings(versionSettings)
	if err != nil {
		zap.L().Fatal("error on updating version detail", zap.Error(err))
	}

	// update system jobs
	systemJobsSettings := &settingsML.Settings{ID: settingsML.KeySystemJobs}
	systemJobs := settingsML.SystemJobsSettings{
		Sunrise: "0 15 1 * * *", // everyday at 1:15 AM
	}
	systemJobsSettings.Spec = utils.StructToMap(systemJobs)
	err = settingsAPI.UpdateSettings(systemJobsSettings)
	if err != nil {
		zap.L().Fatal("error on updating system jobs detail", zap.Error(err))
	}

	// update export locations
	settingsExportLocations := &settingsML.Settings{ID: settingsML.KeySystemBackupLocations}
	settingsExportLocations.Spec = utils.StructToMap(&settingsML.BackupLocations{})
	err = settingsAPI.UpdateSettings(settingsExportLocations)
	if err != nil {
		zap.L().Fatal("error on updating system export locations", zap.Error(err))
	}

	// update analytics config
	settingsAnalytics := &settingsML.Settings{ID: settingsML.KeyAnalytics}
	settingsAnalytics.Spec = utils.StructToMap(&settingsML.AnalyticsConfig{AnonymousID: utils.RandUUID()})
	err = settingsAPI.UpdateSettings(settingsAnalytics)
	if err != nil {
		zap.L().Fatal("error on updating analytics config", zap.Error(err))
	}

}

// UpdateGeoLocation updates geo location if autoUpdate enabled
func UpdateGeoLocation() {
	err := settingsAPI.AutoUpdateSystemGEOLocation()
	if err != nil {
		zap.L().Error("error on updating geo location", zap.Error(err))
	}
}
