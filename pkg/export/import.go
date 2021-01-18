package export

import (
	"errors"
	"fmt"
	"strings"

	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	notificationHandlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notification_handler"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	mldb "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	exportml "github.com/mycontroller-org/backend/v2/pkg/model/export"
	mlfl "github.com/mycontroller-org/backend/v2/pkg/model/field"
	mlfw "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	mlgw "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	mlnd "github.com/mycontroller-org/backend/v2/pkg/model/node"
	mlnh "github.com/mycontroller-org/backend/v2/pkg/model/notification_handler"
	mlsr "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	mlus "github.com/mycontroller-org/backend/v2/pkg/model/user"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
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
	if !ut.IsDirExists(targetDir) {
		return fmt.Errorf("Specified directory not available. targetDir:%s", targetDir)
	}
	files, err := ut.ListFiles(targetDir)
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
		fileBytes, err := ut.ReadFile(targetDir, file.Name)
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
	case ml.EntityGateway:
		entities := make([]mlgw.Config, 0)
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

	case ml.EntityNode:
		entities := make([]mlnd.Node, 0)
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

	case ml.EntitySensor:
		entities := make([]mlsr.Sensor, 0)
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

	case ml.EntitySensorField:
		entities := make([]mlfl.Field, 0)
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

	case ml.EntityFirmware:
		entities := make([]mlfw.Firmware, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fwAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case ml.EntityUser:
		entities := make([]mlus.User, 0)
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

	case ml.EntityDashboard:
		entities := make([]mldb.Config, 0)
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

	case ml.EntityNotificationHandlers:
		entities := make([]mlnh.Config, 0)
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

	default:
		return fmt.Errorf("Unknown entity type:%s", entityName)
	}

	return nil
}

func unmarshal(provider string, fileBytes []byte, entities interface{}) error {
	switch provider {
	case exportml.TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case exportml.TypeYAML:
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
	entity := strings.Split(filename, exportml.EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}
