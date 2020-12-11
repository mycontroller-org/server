package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	kindAPI "github.com/mycontroller-org/backend/v2/pkg/api/kind"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	mlfl "github.com/mycontroller-org/backend/v2/pkg/model/field"
	mlfw "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	mlgw "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	mlkd "github.com/mycontroller-org/backend/v2/pkg/model/kind"
	mlnd "github.com/mycontroller-org/backend/v2/pkg/model/node"
	mlsr "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
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

	case ml.EntityKind:
		entities := make([]mlkd.Kind, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = kindAPI.Save(&entities[index])
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
	case TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case TypeYAML:
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
	entity := strings.Split(filename, EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}
