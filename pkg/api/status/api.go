package status

import (
	"context"
	"fmt"
	"os"
	"time"

	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	EnvironmentDocker     = "docker"
	EnvironmentKubernetes = "kubernetes"
	EnvironmentBareMetal  = "bare_metal"
)

var (
	startTime time.Time
)

func init() {
	startTime = time.Now()
}

type StatusAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	enc     *encryptionAPI.Encryption
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) *StatusAPI {
	return &StatusAPI{
		ctx:     ctx,
		logger:  logger.Named("status_api"),
		storage: storage,
		bus:     bus,
		enc:     enc,
	}
}

type Status struct {
	Hostname          string           `json:"hostname"`
	DocumentationURL  string           `json:"documentationUrl"`
	Login             settingsTY.Login `json:"login"`
	StartTime         time.Time        `json:"startTime"`
	ServerTime        time.Time        `json:"serverTime"`
	Uptime            uint64           `json:"uptime"` // in milliseconds
	MetricsDBDisabled bool             `json:"metricsDBDisabled"`
	Language          string           `json:"language"`
}

func (s *StatusAPI) get(minimal bool) Status {
	status := Status{
		DocumentationURL: types.GetEnvString(types.ENV_DOCUMENTATION_URL),
	}
	status.MetricsDBDisabled = types.GetEnvBool(types.ENV_METRIC_DB_DISABLED)

	if !minimal {
		hostname, err := os.Hostname()
		if err != nil {
			s.logger.Error("error on getting hostname", zap.Error(err))
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		status.Hostname = hostname
		status.ServerTime = time.Now()
		status.StartTime = startTime
		status.Uptime = uint64(time.Since(startTime).Milliseconds())
	}

	// include login message
	login := settingsTY.Login{}
	_settingsAPI := settingsAPI.New(s.ctx, s.logger, s.storage, s.enc, s.bus)
	sysSettings, err := _settingsAPI.GetSystemSettings()
	if err != nil {
		s.logger.Error("error on getting system settings", zap.Error(err))
		login.Message = fmt.Sprintf("error on getting login message: %s", err.Error())
	} else {
		login = sysSettings.Login
		status.Language = sysSettings.Language
	}
	status.Login = login

	return status
}

// Get returns status with all fields
func (s *StatusAPI) Get() Status {
	return s.get(false)
}

// GetMinimal returns limited fields, can be used under status rest api (login not required)
func (s *StatusAPI) GetMinimal() Status {
	return s.get(true)
}

// docker creates a .dockerenv file at the root of the directory tree inside the container.
// if this file exists then the viewer is running from inside a container so return true
// With the default configuration, Kubernetes will mount the serviceaccount secrets into pods.
func (s *StatusAPI) RunningIn() string {
	if utils.IsFileExists("/.dockerenv") {
		return EnvironmentDocker
	} else if utils.IsDirExists("/var/run/secrets/kubernetes.io") {
		return EnvironmentKubernetes
	}
	return EnvironmentBareMetal
}
