package resource

import (
	"errors"

	firmwareAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	firmwareML "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

func firmwareService(reqEvent *rsModel.Event) error {
	resEvent := &rsModel.Event{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getFirmware(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}

	case rsModel.CommandBlocks:
		sendFirmwareBlocks(reqEvent)
		return nil

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getFirmware(request *rsModel.Event) (interface{}, error) {
	if request.ID != "" {
		cfg, err := firmwareAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := firmwareAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func sendFirmwareBlocks(reqEvent *rsModel.Event) {
	if reqEvent.ID == "" {
		return
	}
	fw, err := firmwareAPI.GetByID(reqEvent.ID)
	if err != nil {
		zap.L().Error("error fetching firmware", zap.String("id", reqEvent.ID), zap.Error(err))
		return
	}

	fwBytes, err := utils.ReadFile(model.GetDataDirectoryFirmware(), fw.File.InternalName)
	if err != nil {
		zap.L().Error("error on reading a firmware file", zap.String("directory", model.GetDataDirectoryFirmware()), zap.String("fileName", fw.File.InternalName), zap.Error(err))
		return
	}

	block := 0
	totalBlocks := len(fwBytes) / firmwareML.BlockSize
	if len(fwBytes)%firmwareML.BlockSize > 0 {
		totalBlocks++
	}
	for {
		positionStart := block * firmwareML.BlockSize
		positionEnd := (block + 1) * firmwareML.BlockSize

		reachedEnd := false
		var bytes []byte
		if positionEnd < len(fwBytes) {
			bytes = fwBytes[positionStart:positionEnd]
		} else {
			bytes = fwBytes[positionStart:]
			reachedEnd = true
		}

		err := postFirmwareBlock(fw.ID, bytes, int64(block), int64(totalBlocks))
		if err != nil {
			zap.L().Error("error on posting firmware blocks", zap.String("firmwareId", fw.ID), zap.Error(err))
		}

		if reachedEnd {
			return
		}
		block++
	}

}

func postFirmwareBlock(id string, bytes []byte, blockNumber, totalBlocks int64) error {
	resEvent := &rsModel.Event{
		Type:    rsModel.TypeFirmware,
		Command: rsModel.CommandBlocks,
		ID:      id,
	}

	fwBlock := firmwareML.FirmwareBlock{
		ID:    id,
		Block: blockNumber,
		Total: totalBlocks,
		Data:  bytes,
	}

	resEvent.SetData(fwBlock)
	return postResponse(mcbus.FormatTopic(mcbus.TopicFirmwareBlocks), resEvent)
}
