package backup

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/plugin/handler/backup/disk"
	backupUtil "github.com/mycontroller-org/server/v2/plugin/handler/backup/util"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
)

const PluginBackup = "backup"

// BackupConfig of backup service
type BackupConfig struct {
	ProviderType string
	Spec         map[string]interface{}
}

// Client for backup service
type Client interface {
	Name() string
	Start() error
	Post(variables map[string]interface{}) error
	Close() error
	State() *model.State
}

func NewBackupPlugin(cfg *handlerType.Config) (handlerType.Plugin, error) {
	config := &BackupConfig{}
	err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}

	switch config.ProviderType {
	case backupUtil.ProviderDisk:
		return disk.Init(cfg, config.Spec)

	default:
		return nil, fmt.Errorf("unknown backup provider:%s", config.ProviderType)
	}
}
