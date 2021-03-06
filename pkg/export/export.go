package export

import (
	"errors"
	"fmt"

	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	firmwareAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	forwardPayloadAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	notificationHandlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notify_handler"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	exportml "github.com/mycontroller-org/backend/v2/pkg/model/export"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	pml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	isRunning = concurrency.SafeBool{}
)

var (
	entitiesList = map[string]func(f []pml.Filter, p *pml.Pagination) (*pml.Result, error){
		ml.EntityGateway:        gatewayAPI.List,
		ml.EntityNode:           nodeAPI.List,
		ml.EntitySensor:         sensorAPI.List,
		ml.EntitySensorField:    fieldAPI.List,
		ml.EntityFirmware:       firmwareAPI.List,
		ml.EntityUser:           userAPI.List,
		ml.EntityDashboard:      dashboardAPI.List,
		ml.EntityForwardPayload: forwardPayloadAPI.List,
		ml.EntityNotifyHandler:  notificationHandlerAPI.List,
		ml.EntityTask:           taskAPI.List,
		ml.EntityScheduler:      schedulerAPI.List,
		ml.EntitySettings:       settingsAPI.List,
		ml.EntityDataRepository: dataRepositoryAPI.List,
	}
)

// ExecuteExport exports data from database to disk
func ExecuteExport(targetDir, exportType string) error {
	if isRunning.IsSet() {
		return errors.New("There is a exporter job in progress")
	}
	isRunning.Set()
	defer isRunning.Reset()

	for entityName := range entitiesList {
		listFn := entitiesList[entityName]
		p := &pml.Pagination{
			Limit: exportml.LimitPerFile, SortBy: []pml.Sort{{Field: model.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
		}
		offset := int64(0)
		for {
			p.Offset = offset
			result, err := listFn(nil, p)
			if err != nil {
				zap.L().Error("Failed to get entities", zap.String("entityName", entityName), zap.Error(err))
				return err
			}

			dump(targetDir, entityName, int(result.Offset), result.Data, exportType)

			offset += exportml.LimitPerFile
			if result.Count < offset {
				break
			}
		}
	}
	return nil
}

func dump(targetDir, entityName string, index int, data interface{}, exportType string) {
	var dataBytes []byte
	var err error
	switch exportType {
	case exportml.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", exportType), zap.Error(err))
			return
		}
	case exportml.TypeYAML:
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
	err = ut.WriteFile(targetDir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}
