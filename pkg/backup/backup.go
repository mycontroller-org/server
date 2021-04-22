package export

import (
	"errors"
	"fmt"
	"path"
	"time"

	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	firmwareAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	forwardPayloadAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	notificationHandlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	backupML "github.com/mycontroller-org/backend/v2/pkg/model/backup"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/backend/v2/pkg/version"
	pml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	isRunning = concurrency.SafeBool{}
)

var (
	entitiesList = map[string]func(f []pml.Filter, p *pml.Pagination) (*pml.Result, error){
		model.EntityGateway:        gatewayAPI.List,
		model.EntityNode:           nodeAPI.List,
		model.EntitySource:         sourceAPI.List,
		model.EntityField:          fieldAPI.List,
		model.EntityFirmware:       firmwareAPI.List,
		model.EntityUser:           userAPI.List,
		model.EntityDashboard:      dashboardAPI.List,
		model.EntityForwardPayload: forwardPayloadAPI.List,
		model.EntityHandler:        notificationHandlerAPI.List,
		model.EntityTask:           taskAPI.List,
		model.EntityScheduler:      schedulerAPI.List,
		model.EntitySettings:       settingsAPI.List,
		model.EntityDataRepository: dataRepositoryAPI.List,
	}
)

// ExecuteCopyFirmware copies firmware files
func ExecuteCopyFirmware(targetDir string) error {
	targetDirFullPath := fmt.Sprintf("%s%s", targetDir, model.DirectoryDataFirmware)
	err := utils.CreateDir(targetDirFullPath)
	if err != nil {
		return err
	}

	files, err := utils.ListFiles(model.GetDataDirectoryFirmware())
	if err != nil {
		return err
	}

	for _, file := range files {
		destPath := fmt.Sprintf("%s/%s", targetDirFullPath, file.Name)
		err = utils.CopyFile(file.FullPath, destPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func addBackupInformation(targetDir, storageExportType string) error {
	backupDetails := &backupML.BackupDetails{
		Filename:          path.Base(targetDir),
		StorageExportType: storageExportType,
		CreatedOn:         time.Now(),
		Version:           version.Get(),
	}

	dataBytes, err := yaml.Marshal(backupDetails)
	if err != nil {
		return err
	}

	err = utils.WriteFile(targetDir, backupML.BackupDetailsFilename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", targetDir), zap.String("filename", backupML.BackupDetailsFilename), zap.Error(err))
		return err
	}
	return nil
}

// ExecuteExportStorage exports data from database to disk
func ExecuteExportStorage(targetDir, storageExportType string) error {
	if isRunning.IsSet() {
		return errors.New("there is a exporter job in progress")
	}
	isRunning.Set()
	defer isRunning.Reset()

	// include version details
	addBackupInformation(targetDir, storageExportType)

	targetDirFullPath := fmt.Sprintf("%s%s", targetDir, model.DirectoryDataStorage)

	for entityName := range entitiesList {
		listFn := entitiesList[entityName]
		p := &pml.Pagination{
			Limit: backupML.LimitPerFile, SortBy: []pml.Sort{{Field: model.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
		}
		offset := int64(0)
		for {
			p.Offset = offset
			result, err := listFn(nil, p)
			if err != nil {
				zap.L().Error("failed to get entities", zap.String("entityName", entityName), zap.Error(err))
				return err
			}

			dump(targetDirFullPath, entityName, int(result.Offset), result.Data, storageExportType)

			offset += backupML.LimitPerFile
			if result.Count < offset {
				break
			}
		}
	}
	return nil
}

func dump(targetDir, entityName string, index int, data interface{}, storageExportType string) {
	// update user to userPassword to keep the password on the json export
	if entityName == model.EntityUser {
		if users, ok := data.(*[]userML.User); ok {
			usersWithPasswd := make([]userML.UserWithPassword, 0)
			for _, user := range *users {
				usersWithPasswd = append(usersWithPasswd, userML.UserWithPassword(user))
			}
			if len(usersWithPasswd) > 0 {
				data = usersWithPasswd
			}
		} else {
			zap.L().Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", data)))
		}
	}
	var dataBytes []byte
	var err error
	switch storageExportType {
	case backupML.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return
		}
	case backupML.TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return
		}
	default:
		zap.L().Error("This format not supported", zap.String("format", storageExportType), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, "__", index, storageExportType)
	dir := fmt.Sprintf("%s/%s", targetDir, storageExportType)
	err = utils.WriteFile(targetDir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}
