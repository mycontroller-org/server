package export

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	isRunning = concurrency.SafeBool{}
)

// ExecuteCopyFirmware copies firmware files
func ExecuteCopyFirmware(targetDir string) error {
	targetDirFullPath := fmt.Sprintf("%s%s", targetDir, types.DirectoryDataFirmware)
	err := utils.CreateDir(targetDirFullPath)
	if err != nil {
		return err
	}

	files, err := utils.ListFiles(types.GetDataDirectoryFirmware())
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
	backupDetails := &backupTY.BackupDetails{
		Filename:          path.Base(targetDir),
		StorageExportType: storageExportType,
		CreatedOn:         time.Now(),
		Version:           version.Get(),
	}

	dataBytes, err := yaml.Marshal(backupDetails)
	if err != nil {
		return err
	}

	err = utils.WriteFile(targetDir, backupTY.BackupDetailsFilename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", targetDir), zap.String("filename", backupTY.BackupDetailsFilename), zap.Error(err))
		return err
	}
	return nil
}

// ExecuteExportStorage exports data from database to disk
func ExecuteExportStorage(exportMap map[string]backupTY.ListFunc, targetDir, storageExportType string) error {
	if isRunning.IsSet() {
		return errors.New("there is a exporter job in progress")
	}
	isRunning.Set()
	defer isRunning.Reset()

	// include version details
	err := addBackupInformation(targetDir, storageExportType)
	if err != nil {
		return err
	}

	targetDirFullPath := fmt.Sprintf("%s%s", targetDir, types.DirectoryDataStorage)

	for entityName := range exportMap {
		listFn := exportMap[entityName]
		p := &storageTY.Pagination{
			Limit: backupTY.LimitPerFile, SortBy: []storageTY.Sort{{Field: types.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
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

			offset += backupTY.LimitPerFile
			if result.Count < offset {
				break
			}
		}
	}
	return nil
}

func dump(targetDir, entityName string, index int, data interface{}, storageExportType string) {
	// update user to userPassword to keep the password on the json export
	if entityName == types.EntityUser {
		if users, ok := data.(*[]userTY.User); ok {
			usersWithPasswd := make([]userTY.UserWithPassword, 0)
			for _, user := range *users {
				usersWithPasswd = append(usersWithPasswd, userTY.UserWithPassword(user))
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
	case backupTY.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return
		}
	case backupTY.TypeYAML:
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
