package firmware

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	firmwareML "github.com/mycontroller-org/server/v2/pkg/model/firmware"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]firmwareML.Firmware, 0)
	return store.STORAGE.Find(model.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgType.Filter) (firmwareML.Firmware, error) {
	result := firmwareML.Firmware{}
	err := store.STORAGE.FindOne(model.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (firmwareML.Firmware, error) {
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := firmwareML.Firmware{}
	err := store.STORAGE.FindOne(model.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func Save(firmware *firmwareML.Firmware, keepFile bool) error {
	eventType := eventML.TypeUpdated
	if firmware.ID == "" {
		firmware.ID = utils.RandID()
		eventType = eventML.TypeCreated
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: firmware.ID},
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

	err := store.STORAGE.Upsert(model.EntityFirmware, firmware, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventFirmware, eventType, model.EntityFirmware, firmware)
	return nil
}

// Delete firmwares
func Delete(ids []string) (int64, error) {
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: ids}}
	pagination := &stgType.Pagination{Limit: 100}

	// delete firmwares
	response, err := List(filters, pagination)
	if err != nil {
		return 0, err
	}
	firmwares := *response.Data.(*[]firmwareML.Firmware)
	for index := 0; index < len(firmwares); index++ {
		firmware := firmwares[index]
		firmwareDirectory := model.GetDataDirectoryFirmware()
		filename := fmt.Sprintf("%s/%s", firmwareDirectory, firmware.File.InternalName)
		err := os.Remove(filename)
		if err != nil {
			zap.L().Error("error on deleting firmware file", zap.Any("firmware", firmware), zap.String("filename", filename), zap.Error(err))
		}
	}

	// delete entries
	return store.STORAGE.Delete(model.EntityFirmware, filters)
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

	firmwareDirectory := model.GetDataDirectoryFirmware()
	err = utils.CreateDir(firmwareDirectory)
	if err != nil {
		return err
	}

	fullPath := fmt.Sprintf("%s/%s", firmwareDirectory, newFilename)
	targetFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	// Copy the file to the destination path
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	// get file details
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	// taking md5sum from sourceFile, returns wring md5 hash
	// load agin from disk
	savedFile, err := os.Open(fullPath)
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
