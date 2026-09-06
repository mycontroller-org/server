package resource

import (
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/types"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

func (svc *ResourceService) firmwareService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getFirmware(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandBlocks:
		svc.sendFirmwareBlocks(reqEvent)
		return nil

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getFirmware(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := svc.api.Firmware().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Firmware().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) sendFirmwareBlocks(reqEvent *rsTY.ServiceEvent) {
	if reqEvent.ID == "" || reqEvent.ReplyTopic == "" {
		return
	}
	fw, err := svc.api.Firmware().GetByID(reqEvent.ID)
	if err != nil {
		svc.logger.Error("error fetching firmware", zap.String("id", reqEvent.ID), zap.Error(err))
		return
	}
	firmwareBaseDir := types.GetEnvString(types.ENV_DIR_DATA_FIRMWARE)
	fwBytes, err := utils.ReadFile(firmwareBaseDir, fw.File.InternalName)
	if err != nil {
		svc.logger.Error("error on reading a firmware file", zap.String("directory", firmwareBaseDir), zap.String("fileName", fw.File.InternalName), zap.Error(err))
		return
	}

	// One message with the whole file. Chunking into 512-byte bus replies races:
	// IsFinal can be handled while earlier chunks are still in flight, and a
	// nested CommandGet from the gateway mutates the shared decode buffer.
	if err := svc.postFirmwareBlock(reqEvent.ReplyTopic, fw.ID, fwBytes, 0, len(fwBytes), true); err != nil {
		svc.logger.Error("error on posting firmware blocks", zap.String("firmwareId", fw.ID), zap.Error(err))
	}
}

func (svc *ResourceService) postFirmwareBlock(replyTopic, id string, bytes []byte, blockNumber, totalBytes int, isFinal bool) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    rsTY.TypeFirmware,
		Command: rsTY.CommandBlocks,
		ID:      id,
	}

	fwBlock := firmwareTY.FirmwareBlock{
		ID:          id,
		BlockNumber: blockNumber,
		TotalBytes:  totalBytes,
		Data:        bytes,
		IsFinal:     isFinal,
	}

	resEvent.SetData(fwBlock)
	return svc.postResponse(replyTopic, resEvent)
}
