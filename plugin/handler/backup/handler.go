package backup

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/plugin/handler/backup/disk"
	backupUtils "github.com/mycontroller-org/server/v2/plugin/handler/backup/util"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
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
	State() *types.State
}

func New(ctx context.Context, cfg *handlerTY.Config) (handlerTY.Plugin, error) {
	providerType := cfg.Spec.GetString(backupUtils.KeyProviderType)
	switch providerType {
	case backupUtils.ProviderDisk:
		return disk.New(ctx, cfg)

	default:
		return nil, fmt.Errorf("unknown backup provider:%s", providerType)
	}
}
