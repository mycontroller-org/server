package firmware

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]firmwareTY.Firmware, 0)
	return store.STORAGE.Find(types.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func Get(filters []storageTY.Filter) (firmwareTY.Firmware, error) {
	result := firmwareTY.Firmware{}
	err := store.STORAGE.FindOne(types.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (firmwareTY.Firmware, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := firmwareTY.Firmware{}
	err := store.STORAGE.FindOne(types.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func Save(firmware *firmwareTY.Firmware, keepFile bool) error {
	eventType := eventTY.TypeUpdated
	if firmware.ID == "" {
		firmware.ID = utils.RandID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: firmware.ID},
	}

	if !configuration.PauseModifiedOnUpdate.IsSet() {
		firmware.ModifiedOn = time.Now()
	}

	if keepFile {
		firmwareOld, err := GetByID(firmware.ID)
		if err == nil {
			firmware.File = firmwareOld.File
		}
	}

	err := store.STORAGE.Upsert(types.EntityFirmware, firmware, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventFirmware, eventType, types.EntityFirmware, firmware)
	return nil
}

// Delete firmwares
func Delete(ids []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}

	// delete firmwares
	response, err := List(filters, pagination)
	if err != nil {
		return 0, err
	}
	firmwares := *response.Data.(*[]firmwareTY.Firmware)
	for index := 0; index < len(firmwares); index++ {
		firmware := firmwares[index]
		firmwareDirectory := types.GetDataDirectoryFirmware()
		filename := fmt.Sprintf("%s/%s", firmwareDirectory, firmware.File.InternalName)
		err := os.Remove(filename)
		if err != nil {
			zap.L().Error("error on deleting firmware file", zap.Any("firmware", firmware), zap.String("filename", filename), zap.Error(err))
		}
	}

	// delete entries
	return store.STORAGE.Delete(types.EntityFirmware, filters)
}

// Upload a firmware file
func Upload(sourceFile multipart.File, id, filename string) error {
	// get firmware
	firmware, err := GetByID(id)
	if err != nil {
		return err
	}

	oldFile := firmware.File.InternalName

	extension := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", id, extension)

	firmwareDirectory := types.GetDataDirectoryFirmware()
	err = utils.CreateDir(firmwareDirectory)
	if err != nil {
		return err
	}

	fileFullPath := fmt.Sprintf("%s/%s", firmwareDirectory, newFilename)
	// delete the existing file if any
	if utils.IsFileExists(fileFullPath) {
		err = utils.RemoveFileOrEmptyDir(fileFullPath)
		if err != nil {
			zap.L().Error("error on deleting existing file", zap.String("filename", fileFullPath), zap.Error(err))
			return err
		}
	}

	targetFile, err := os.OpenFile(fileFullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	// Copy the file to the destination path
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	// get file details
	fileInfo, err := os.Stat(fileFullPath)
	if err != nil {
		return err
	}

	// taking md5sum from sourceFile, returns wring md5 hash
	// load agin from disk
	savedFile, err := os.Open(fileFullPath)
	if err != nil {
		return err
	}
	defer savedFile.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, savedFile); err != nil {
		return err
	}
	checkSum := hash.Sum(nil)

	// update file object in firmware
	firmware.File.Name = filename
	firmware.File.InternalName = newFilename
	firmware.File.Size = int(fileInfo.Size())
	firmware.File.Checksum = fmt.Sprintf("sha256:%x", checkSum)
	firmware.File.ModifiedOn = fileInfo.ModTime()

	err = Save(&firmware, false)
	if err != nil {
		return err
	}

	// remove old file, if the extension different
	if oldFile != "" && oldFile != newFilename {
		oldFileFullPath := fmt.Sprintf("%s/%s", firmwareDirectory, oldFile)
		err = os.Remove(oldFileFullPath)
		if err != nil {
			zap.L().Error("error on removing old file", zap.String("file", oldFileFullPath), zap.Error(err))
			return err
		}
	}
	return nil
}
