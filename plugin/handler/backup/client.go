// +build server,!standalone

package exporter

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/handler/backup/disk"
	backupUtil "github.com/mycontroller-org/backend/v2/plugin/handler/backup/util"
)

// Config of backup service
type Config struct {
	ProviderType string
	Spec         map[string]interface{}
}

// Client for backup service
type Client interface {
	Start() error
	Post(variables map[string]interface{}) error
	Close() error
	State() *model.State
}

// Init backup client
func Init(cfg *handlerML.Config) (Client, error) {
	config := &Config{}
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
