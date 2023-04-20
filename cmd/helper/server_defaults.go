package helper

import (
	"time"

	systemJobs "github.com/mycontroller-org/server/v2/pkg/service/system_jobs"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/upgrade"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	loginMessage = "Default username and password to login: <b>admin</b> / <b>admin</b>"
)

// update initial system settings
// works on empty database
func (s *Server) updateInitialSystemSettings() {

	// update system settings data
	_, err := s.api.Settings().GetByID(settingsTY.KeySystemSettings)
	if err == nil {
		return
	}
	s.logger.Info("error on fetching system settings, assuming it is fresh install and populating default details. if not, report this issue", zap.Error(err))
	// update system settings
	systemSettings := settingsTY.SystemSettings{
		GeoLocation: settingsTY.GeoLocation{AutoUpdate: true},
		Login:       settingsTY.Login{Message: loginMessage},
		NodeStateJob: settingsTY.NodeStateJob{
			ExecutionInterval: systemJobs.DefaultExecutionInterval,
			InactiveDuration:  systemJobs.DefaultInactiveDuration,
		},
	}
	settings := &settingsTY.Settings{ID: settingsTY.KeySystemSettings}
	settings.Spec = utils.StructToMap(systemSettings)
	err = s.api.Settings().UpdateSystemSettings(settings)
	if err != nil {
		s.logger.Fatal("error on updating system settings", zap.Error(err))
	}

	// update version details
	// these details are used on upgrade to apply patches
	versionSettings := &settingsTY.Settings{ID: settingsTY.KeyVersion}
	ver := version.Get()
	versionData := settingsTY.VersionSettings{
		Version:     ver.Version,
		GitCommit:   ver.GitCommit,
		Database:    s.storage.Name(),
		InstalledOn: time.Now().Format(time.RFC3339),
		LastUpgrade: upgrade.GetLatestUpgradeVersion(),
	}
	versionSettings.Spec = utils.StructToMap(versionData)
	err = s.api.Settings().UpdateSettings(versionSettings)
	if err != nil {
		s.logger.Fatal("error on updating version detail", zap.Error(err))
	}

	// update system jobs
	systemJobsSettings := &settingsTY.Settings{ID: settingsTY.KeySystemJobs}
	systemJobs := settingsTY.SystemJobsSettings{
		Sunrise: "0 15 1 * * *", // everyday at 1:15 AM
	}
	systemJobsSettings.Spec = utils.StructToMap(systemJobs)
	err = s.api.Settings().UpdateSettings(systemJobsSettings)
	if err != nil {
		s.logger.Fatal("error on updating system jobs detail", zap.Error(err))
	}

	// update export locations
	settingsExportLocations := &settingsTY.Settings{ID: settingsTY.KeySystemBackupLocations}
	settingsExportLocations.Spec = utils.StructToMap(&settingsTY.BackupLocations{})
	err = s.api.Settings().UpdateSettings(settingsExportLocations)
	if err != nil {
		s.logger.Fatal("error on updating system export locations", zap.Error(err))
	}

	// update telemetry config
	settingsTelemetry := &settingsTY.Settings{ID: settingsTY.KeyTelemetry}
	settingsTelemetry.Spec = utils.StructToMap(&settingsTY.TelemetryConfig{AnonymousID: utils.RandUUID()})
	err = s.api.Settings().UpdateSettings(settingsTelemetry)
	if err != nil {
		s.logger.Fatal("error on updating telemetry config", zap.Error(err))
	}

	// update resource handler
	s.setupResourceHandler()
}

// setup initial user
// if there is no user found the database, this user will be created
func (s *Server) setupInitialUser() {
	pagination := &storageTY.Pagination{
		Limit: 1,
	}
	usersResult, err := s.api.User().List(nil, pagination)
	if err != nil {
		s.logger.Error("failed to list users", zap.Error(err))
	}
	if usersResult.Count == 0 {
		s.logger.Info("there is no user available in the storage database. creating default user 'admin'")
		hashedPassword, err := hashed.GenerateHash("admin")
		if err != nil {
			s.logger.Fatal("unable to get hashed password", zap.Error(err))
			return
		}
		adminUser := &userTY.User{
			Username: "admin",
			Password: hashedPassword,
			FullName: "Admin User",
			Email:    "admin@example.com",
		}
		err = s.api.User().Save(adminUser)
		if err != nil {
			s.logger.Error("failed to create default admin user", zap.Error(err))
		}
	}
}

// setup resource handler
func (s *Server) setupResourceHandler() {
	resourceHandler := handlerTY.Config{
		ID:          "resource_handler",
		Description: "Sends payload to resources",
		Enabled:     true,
		Labels: cmap.CustomStringMap{
			"location": "server",
		},
		Type: handlerTY.DataTypeResource,
	}

	err := s.api.Handler().Save(&resourceHandler)
	if err != nil {
		s.logger.Error("error on adding resource handler", zap.Error(err))
	}
}
