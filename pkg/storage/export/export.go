package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/api/field"
	"github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	"github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/api/kind"
	"github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"github.com/mycontroller-org/backend/v2/pkg/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// global export/import contants
const (
	TypeJSON = "json"
	TypeYAML = "yaml"

	ExporterNone = "none"

	LimitPerFile         = 50
	EntityNameIndexSplit = "__"

	DateSuffixLayout = "20060102_150405"
)

// Config for export job
type Config struct {
	Enabled       bool     `json:"enabled"`
	Interval      string   `json:"interval"`
	TargetDir     string   `json:"targetDir"`
	Clean         bool     `json:"clean"`
	UseDateSuffix bool     `json:"useDateSuffix"`
	ExportType    []string `json:"exportType"`
	Exporter      []string `json:"exporter"`
}

var (
	isRunning    = false
	entitiesList = map[string]func(f []pml.Filter, p *pml.Pagination) (*pml.Result, error){
		ml.EntityGateway:     gateway.List,
		ml.EntityNode:        node.List,
		ml.EntitySensor:      sensor.List,
		ml.EntitySensorField: field.List,
		ml.EntityFirmware:    firmware.List,
		ml.EntityKind:        kind.List,
	}
)

// ExporterFuncCall implementation
func ExporterFuncCall(cfg Config) func() {
	return func() {
		if len(cfg.ExportType) == 0 {
			zap.L().Error("No export type defined", zap.Any("config", cfg))
			return
		}
		// generate targetDirname
		targetDir := cfg.TargetDir
		if cfg.UseDateSuffix {
			suffix := time.Now().Format(DateSuffixLayout)
			targetDir = fmt.Sprintf("%s_%s", cfg.TargetDir, suffix)
		}
		// export data in different export type
		for _, exportType := range cfg.ExportType {
			ExecuteExport(targetDir, exportType)
		}
		// execute exporter plugins
		// TODO...
	}
}

// ExecuteExport exports data from database to disk
func ExecuteExport(targetDir, exportType string) error {
	if isRunning {
		return errors.New("There is a exporter job in progress")
	}
	isRunning = true
	defer func() { isRunning = false }()

	for entityName, listFn := range entitiesList {
		p := &pml.Pagination{
			Limit: LimitPerFile, SortBy: []pml.Sort{{Field: model.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
		}
		offset := int64(0)
		for {
			p.Offset = offset
			result, err := listFn(nil, p)
			if err != nil {
				zap.L().Error("Failed to get entities", zap.String("entityName", entityName), zap.Error(err))
				return err
			}
			if result.Count == 0 {
				break
			}
			offset++

		}
	}
	return nil
}

func dump(targetDir, entityName string, index int, data interface{}, exportType string) {
	var dataBytes []byte
	var err error
	switch exportType {
	case TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", exportType), zap.Error(err))
			return
		}
	case TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", exportType), zap.Error(err))
			return
		}
	default:
		zap.L().Error("This format not supported", zap.String("format", exportType), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, "__", index, exportType)
	dir := fmt.Sprintf("%s/%s", targetDir, exportType)
	err = util.WriteFile(targetDir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}
