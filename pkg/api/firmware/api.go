package firmware

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fml.Firmware, 0)
	return stg.SVC.Find(ml.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (fml.Firmware, error) {
	result := fml.Firmware{}
	err := stg.SVC.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (fml.Firmware, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	result := fml.Firmware{}
	err := stg.SVC.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func Save(firmware *fml.Firmware, keepFile bool) error {
	if firmware.ID == "" {
		firmware.ID = ut.RandID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: firmware.ID},
	}
	firmware.LastModifiedOn = time.Now()

	if keepFile {
		firmwareOld, err := GetByID(firmware.ID)
		if err == nil {
			firmware.File = firmwareOld.File
		}
	}

	return stg.SVC.Upsert(ml.EntityFirmware, firmware, filters)
}

// Delete firmwares
func Delete(ids []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}

	// delete firmwares
	response, err := List(filters, pagination)
	if err != nil {
		return 0, err
	}
	firmwares := *response.Data.(*[]fml.Firmware)
	for index := 0; index < len(firmwares); index++ {
		firmware := firmwares[index]
		firmwareDirectory := ml.GetDirectoryFirmware()
		filename := fmt.Sprintf("%s/%s", firmwareDirectory, firmware.File.InternalName)
		err := os.Remove(filename)
		if err != nil {
			zap.L().Error("error on deleting firmware file", zap.Any("firmware", firmware), zap.String("filename", filename), zap.Error(err))
		}
	}

	// delete entries
	return stg.SVC.Delete(ml.EntityFirmware, filters)
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

	firmwareDirectory := ml.GetDirectoryFirmware()
	utils.CreateDir(firmwareDirectory)

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

	hash := md5.New()
	if _, err := io.Copy(hash, savedFile); err != nil {
		return err
	}
	checkSum := hash.Sum(nil)

	// update file object in firmware
	firmware.File.Name = filename
	firmware.File.InternalName = newFilename
	firmware.File.Size = int(fileInfo.Size())
	firmware.File.Checksum = fmt.Sprintf("md5:%x", checkSum)
	firmware.File.ModifiedTime = fileInfo.ModTime()

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
