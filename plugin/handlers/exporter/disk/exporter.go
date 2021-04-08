package disk

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	variableUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	exporter "github.com/mycontroller-org/backend/v2/plugin/handlers/exporter/util"
	"go.uber.org/zap"
)

// Config of disk exporter
type Config struct {
	ExportType      string
	TargetDirectory string
}

// Client struct
type Client struct {
	handlerCfg *handlerML.Config
	cfg        *Config
}

// Init disk exporter
func Init(cfg *handlerML.Config, spec map[string]interface{}) (*Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, spec, config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		handlerCfg: cfg,
		cfg:        config,
	}

	return client, nil
}

// Start func
func (c *Client) Start() error {
	return nil
}

// Close Func
func (c *Client) Close() error {
	return nil
}

// State func
func (c *Client) State() *model.State {
	if c.handlerCfg != nil {
		if c.handlerCfg.State == nil {
			c.handlerCfg.State = &model.State{}
		}
		return c.handlerCfg.State
	}
	return &model.State{}
}

// Post func
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		zap.L().Debug("processing a request", zap.String("name", name), zap.Any("value", value))
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerML.DataTypeExporter {
			continue
		}

		exporterData := handlerML.ExporterData{}
		err = variableUtils.UnmarshalBase64Yaml(genericData.Data, &exporterData)
		if err != nil {
			zap.L().Error("error on converting exporter data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
			continue
		}

		if exporterData.ExporterType != exporter.TypeExporterDisk {
			continue
		}

		err = c.triggerExport(exporterData.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) triggerExport(spec map[string]interface{}) error {
	newConfig := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, spec, newConfig)
	if err != nil {
		return err
	}

	zap.L().Debug("data", zap.Any("config", newConfig))

	targetExportType := c.cfg.ExportType
	targetDirectory := c.cfg.TargetDirectory

	if newConfig.ExportType != "" {
		targetExportType = newConfig.ExportType
	}

	if newConfig.TargetDirectory != "" {
		targetDirectory = newConfig.TargetDirectory
	}

	if strings.HasSuffix(targetDirectory, "/") {
		targetDirectory = targetDirectory[:len(targetDirectory)-1]
	}

	start := time.Now()
	zap.L().Info("Export job triggered")
	// start export
	filename, err := exporter.Export(targetExportType)
	if err != nil {
		return err
	}

	// move the file to target location
	// get final file name
	zipFilename := filepath.Base(filename)
	targetLocation := fmt.Sprintf("%s/%s", targetDirectory, zipFilename)
	err = utils.CopyFile(filename, targetLocation)
	if err != nil {
		return err
	}
	err = utils.RemoveFileOrEmptyDir(filename)
	if err != nil {
		return err
	}

	zap.L().Info("Export job completed", zap.String("timeTaken", time.Since(start).String()))

	return nil
}
