package server

import (
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	systemJobs "github.com/mycontroller-org/server/v2/pkg/service/system_jobs"
	nodeJobs "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/node_job"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
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
	pagination := &storageTY.Pagination{
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
		adminUser := &userTY.User{
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
	_, err := settingsAPI.GetByID(settingsTY.KeySystemSettings)
	if err == nil {
		return
	}
	zap.L().Info("error on fetching system settings, assuming it is fresh install and populating default details. if not, report this issue", zap.String("error", err.Error()))

	// update system settings
	systemSettings := settingsTY.SystemSettings{
		GeoLocation: settingsTY.GeoLocation{AutoUpdate: true},
		Login:       settingsTY.Login{Message: loginMessage},
		NodeStateJob: settingsTY.NodeStateJob{
			ExecutionInterval: nodeJobs.DefaultExecutionInterval,
			InactiveDuration:  nodeJobs.DefaultInactiveDuration,
		},
	}
	settings := &settingsTY.Settings{ID: settingsTY.KeySystemSettings}
	settings.Spec = utils.StructToMap(systemSettings)
	err = settingsAPI.UpdateSystemSettings(settings)
	if err != nil {
		zap.L().Fatal("error on updating system settings", zap.Error(err))
	}

	// update version details
	versionSettings := &settingsTY.Settings{ID: settingsTY.KeyVersion}
	versionData := settingsTY.VersionSettings{Version: version.Get().Version}
	versionSettings.Spec = utils.StructToMap(versionData)
	err = settingsAPI.UpdateSettings(versionSettings)
	if err != nil {
		zap.L().Fatal("error on updating version detail", zap.Error(err))
	}

	// update system jobs
	systemJobsSettings := &settingsTY.Settings{ID: settingsTY.KeySystemJobs}
	systemJobs := settingsTY.SystemJobsSettings{
		Sunrise: "0 15 1 * * *", // everyday at 1:15 AM
	}
	systemJobsSettings.Spec = utils.StructToMap(systemJobs)
	err = settingsAPI.UpdateSettings(systemJobsSettings)
	if err != nil {
		zap.L().Fatal("error on updating system jobs detail", zap.Error(err))
	}

	// update export locations
	settingsExportLocations := &settingsTY.Settings{ID: settingsTY.KeySystemBackupLocations}
	settingsExportLocations.Spec = utils.StructToMap(&settingsTY.BackupLocations{})
	err = settingsAPI.UpdateSettings(settingsExportLocations)
	if err != nil {
		zap.L().Fatal("error on updating system export locations", zap.Error(err))
	}

	// update analytics config
	settingsAnalytics := &settingsTY.Settings{ID: settingsTY.KeyAnalytics}
	settingsAnalytics.Spec = utils.StructToMap(&settingsTY.AnalyticsConfig{AnonymousID: utils.RandUUID()})
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
