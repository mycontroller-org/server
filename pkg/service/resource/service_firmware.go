package resource

import (
	"errors"

	firmwareAPI "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	"github.com/mycontroller-org/server/v2/pkg/model"
	firmwareML "github.com/mycontroller-org/server/v2/pkg/model/firmware"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

func firmwareService(reqEvent *rsML.ServiceEvent) error {
	resEvent := &rsML.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsML.CommandGet:
		data, err := getFirmware(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsML.CommandBlocks:
		sendFirmwareBlocks(reqEvent)
		return nil

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getFirmware(request *rsML.ServiceEvent) (interface{}, error) {
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

func sendFirmwareBlocks(reqEvent *rsML.ServiceEvent) {
	if reqEvent.ID == "" || reqEvent.ReplyTopic == "" {
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

	blockNumber := 0
	totalBytes := len(fwBytes)
	for {
		positionStart := blockNumber * firmwareML.BlockSize
		positionEnd := positionStart + firmwareML.BlockSize

		reachedEnd := false
		var bytes []byte
		if positionEnd < len(fwBytes) {
			bytes = fwBytes[positionStart:positionEnd]
		} else {
			bytes = fwBytes[positionStart:]
			reachedEnd = true
		}

		err := postFirmwareBlock(reqEvent.ReplyTopic, fw.ID, bytes, blockNumber, totalBytes, reachedEnd)
		if err != nil {
			zap.L().Error("error on posting firmware blocks", zap.String("firmwareId", fw.ID), zap.Error(err))
		}

		if reachedEnd {
			return
		}
		blockNumber++
	}

}

func postFirmwareBlock(replyTopic, id string, bytes []byte, blockNumber, totalBytes int, isFinal bool) error {
	resEvent := &rsML.ServiceEvent{
		Type:    rsML.TypeFirmware,
		Command: rsML.CommandBlocks,
		ID:      id,
	}

	fwBlock := firmwareML.FirmwareBlock{
		ID:          id,
		BlockNumber: blockNumber,
		TotalBytes:  totalBytes,
		Data:        bytes,
		IsFinal:     isFinal,
	}

	resEvent.SetData(fwBlock)
	return postResponse(replyTopic, resEvent)
}
