package firmware

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type FirmwareAPI struct {
	ctx               context.Context
	logger            *zap.Logger
	storage           storageTY.Plugin
	bus               busTY.Plugin
	firmwareDirectory string
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *FirmwareAPI {
	return &FirmwareAPI{
		ctx:               ctx,
		logger:            logger.Named("firmware_api"),
		storage:           storage,
		bus:               bus,
		firmwareDirectory: types.GetEnvString(types.ENV_DIR_DATA_FIRMWARE),
	}
}

// List by filter and pagination
func (fw *FirmwareAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]firmwareTY.Firmware, 0)
	return fw.storage.Find(types.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func (fw *FirmwareAPI) Get(filters []storageTY.Filter) (firmwareTY.Firmware, error) {
	result := firmwareTY.Firmware{}
	err := fw.storage.FindOne(types.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func (fw *FirmwareAPI) GetByID(id string) (firmwareTY.Firmware, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := firmwareTY.Firmware{}
	err := fw.storage.FindOne(types.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func (fw *FirmwareAPI) Save(firmware *firmwareTY.Firmware, keepFile bool) error {
	eventType := eventTY.TypeUpdated
	if firmware.ID == "" {
		firmware.ID = utils.RandID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: firmware.ID},
	}

	firmware.ModifiedOn = time.Now()

	if keepFile {
		firmwareOld, err := fw.GetByID(firmware.ID)
		if err == nil {
			firmware.File = firmwareOld.File
		}
	}

	err := fw.storage.Upsert(types.EntityFirmware, firmware, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(fw.logger, fw.bus, topic.TopicEventFirmware, eventType, types.EntityFirmware, firmware)
	return nil
}

// Delete firmwares
func (fw *FirmwareAPI) Delete(ids []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}

	// delete firmwares
	response, err := fw.List(filters, pagination)
	if err != nil {
		return 0, err
	}
	firmwares := *response.Data.(*[]firmwareTY.Firmware)
	for index := 0; index < len(firmwares); index++ {
		firmware := firmwares[index]
		filename := fmt.Sprintf("%s/%s", fw.firmwareDirectory, firmware.File.InternalName)
		err := os.Remove(filename)
		if err != nil {
			fw.logger.Error("error on deleting firmware file", zap.Any("firmware", firmware), zap.String("filename", filename), zap.Error(err))
		}
	}

	// delete entries
	return fw.storage.Delete(types.EntityFirmware, filters)
}

// Upload a firmware file
func (fw *FirmwareAPI) Upload(sourceFile multipart.File, id, filename string) error {
	// get firmware
	firmware, err := fw.GetByID(id)
	if err != nil {
		return err
	}

	oldFile := firmware.File.InternalName

	extension := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", id, extension)

	err = utils.CreateDir(fw.firmwareDirectory)
	if err != nil {
		return err
	}

	fileFullPath := fmt.Sprintf("%s/%s", fw.firmwareDirectory, newFilename)
	// delete the existing file if any
	if utils.IsFileExists(fileFullPath) {
		err = utils.RemoveFileOrEmptyDir(fileFullPath)
		if err != nil {
			fw.logger.Error("error on deleting existing file", zap.String("filename", fileFullPath), zap.Error(err))
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

	err = fw.Save(&firmware, false)
	if err != nil {
		return err
	}

	// remove old file, if the extension different
	if oldFile != "" && oldFile != newFilename {
		oldFileFullPath := fmt.Sprintf("%s/%s", fw.firmwareDirectory, oldFile)
		err = os.Remove(oldFileFullPath)
		if err != nil {
			fw.logger.Error("error on removing old file", zap.String("file", oldFileFullPath), zap.Error(err))
			return err
		}
	}
	return nil
}

func (fw *FirmwareAPI) Import(data interface{}) error {
	input, ok := data.(firmwareTY.Firmware)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}

	return fw.storage.Upsert(types.EntityFirmware, &input, filters)
}

func (fw *FirmwareAPI) GetEntityInterface() interface{} {
	return firmwareTY.Firmware{}
}
