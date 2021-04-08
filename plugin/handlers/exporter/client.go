package exporter

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/handlers/exporter/disk"
	exporter "github.com/mycontroller-org/backend/v2/plugin/handlers/exporter/util"
)

// Config of email service
type Config struct {
	ExporterType string
	Spec         map[string]interface{}
}

// Client for email service
type Client interface {
	Start() error
	Post(variables map[string]interface{}) error
	Close() error
	State() *model.State
}

// Init exporter client
func Init(cfg *handlerML.Config) (Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}

	switch config.ExporterType {
	case exporter.TypeExporterDisk:
		return disk.Init(cfg, config.Spec)

	default:
		return nil, fmt.Errorf("Unknown exporter client:%s", cfg.Type)
	}
}
