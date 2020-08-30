package mysensors

import (
	"errors"

	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
	"go.uber.org/zap"
)

// executeFirmwareConfigRequest executes firmware config request and response with hex payload
func executeFirmwareConfigRequest(msg *msgml.Message) (string, error) {
	// get the assigned firmware details
	// if non assigned, throw error

	// convert the received hex to matching struct format
	fwCfgReq := &firmwareConfigRequest{}
	err := toStruct(msg.Payload, fwCfgReq)
	if err != nil {
		return "", err
	}
	zap.L().Debug("Received a firmware config request", zap.Any("firmwareConfigRequest", fwCfgReq))
	// get the node details
	node, err := nodeAPI.GetByIDs(msg.GatewayID, msg.NodeID)
	if err != nil {
		return "", err
	}

	// get mapped firmware by id
	fwID := node.Labels.Get(nml.LabelAssignedFirmware)
	if fwID == "" {
		return "", errors.New("Firmware not assigned for this node")
	}

	// get firmware details
	firmware, err := fwAPI.GetByID(fwID)
	if err != nil {
		return "", err
	}

	// get mysensor specific ids
	if firmware.Labels.Get(LabelFirmwareTypeID) == "" || firmware.Labels.Get(LabelFirmwareVersionID) == "" {
		return "", errors.New("Firmware type id or version id not set")
	}
	fwTypeID := firmware.Labels.GetInt(LabelFirmwareTypeID)
	fwVersionID := firmware.Labels.GetInt(LabelFirmwareVersionID)

	// calculate firmware CRC and blocks
	// get firmware hex file
	hexFile, err := ut.ReadFile(ml.DirectoryFullPath(ml.DirectoryFirmware), firmware.File.Name)
	if err != nil {
		return "", err
	}

	fwFile, err := hexByteToLocalFormat(hexFile)
	if err != nil {
		return "", err
	}

	// create firmware config response struct and update required values
	fwCfgRes := &firmwareConfigResponse{}

	// if erase eeprom set for this node, update erase eeprom command and clear the label on the node detail
	if node.Labels.GetBool(LabelEraseEEPROM) {
		// set erase command
		fwCfgRes.SetEraseEEPROM()
		// remove erase config data from node
		node.Labels.Set(LabelEraseEEPROM, "false")
		err = nodeAPI.Save(node)
		if err != nil {
			return "", err
		}
	} else { // update assigned firmware config details
		fwCfgRes.Type = uint16(fwTypeID)
		fwCfgRes.Version = uint16(fwVersionID)
		fwCfgRes.Blocks = uint16(fwFile.Blocks)
		fwCfgRes.CRC = uint16(fwFile.CRC)
	}

	// convert the struct to hex string and return
	return toHex(fwCfgRes)
}

func hexByteToLocalFormat(hexByte []byte) (*firmware, error) {
	return nil, nil
}
