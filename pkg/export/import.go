package export

import (
	"errors"
	"fmt"
	"strings"

	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	fpAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	notificationHandlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notify_handler"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	dashboardML "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	dataRepositoryML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	exportML "github.com/mycontroller-org/backend/v2/pkg/model/export"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	firmwareML "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	fpML "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	nhML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	sensorML "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	settingsML "github.com/mycontroller-org/backend/v2/pkg/model/settings"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var isImportJobRunning = false

// ExecuteImport update data into database
func ExecuteImport(targetDir, fileType string) error {
	if isImportJobRunning {
		return errors.New("There is an import job is progress")
	}
	isImportJobRunning = true
	defer func() { isImportJobRunning = false }()

	zap.L().Debug("Executing import job", zap.String("targetDir", targetDir), zap.String("fileType", fileType))
	// check directory availability
	if !utils.IsDirExists(targetDir) {
		return fmt.Errorf("Specified directory not available. targetDir:%s", targetDir)
	}
	files, err := utils.ListFiles(targetDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("No files found on:%s", targetDir)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name, fileType) {
			continue
		}
		entityName := getEntityName(file.Name)
		zap.L().Debug("Importing a file", zap.String("file", file.Name), zap.String("entityName", entityName))
		// read data from file
		fileBytes, err := utils.ReadFile(targetDir, file.Name)
		if err != nil {
			return err
		}
		err = updateEntities(fileBytes, entityName, fileType)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateEntities(fileBytes []byte, entityName, fileFormat string) error {
	switch entityName {
	case model.EntityGateway:
		entities := make([]gatewayML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = gwAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityNode:
		entities := make([]nodeML.Node, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = nodeAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntitySensor:
		entities := make([]sensorML.Sensor, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = sensorAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntitySensorField:
		entities := make([]fieldML.Field, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fieldAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityFirmware:
		entities := make([]firmwareML.Firmware, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fwAPI.Save(&entities[index], false)
			if err != nil {
				return err
			}
		}

	case model.EntityUser:
		entities := make([]userML.User, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = userAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityDashboard:
		entities := make([]dashboardML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = dashboardAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityNotifyHandler:
		entities := make([]nhML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = notificationHandlerAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityForwardPayload:
		entities := make([]fpML.Mapping, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fpAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityTask:
		entities := make([]taskML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = taskAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityScheduler:
		entities := make([]schedulerML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = schedulerAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntitySettings:
		entities := make([]settingsML.Settings, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = settingsAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case model.EntityDataRepository:
		entities := make([]dataRepositoryML.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = dataRepositoryAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Unknown entity type:%s", entityName)
	}

	return nil
}

func unmarshal(provider string, fileBytes []byte, entities interface{}) error {
	switch provider {
	case exportML.TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case exportML.TypeYAML:
		err := yaml.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown provider:%s", provider)
	}
	return nil
}

func getEntityName(filename string) string {
	entity := strings.Split(filename, exportML.EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}
