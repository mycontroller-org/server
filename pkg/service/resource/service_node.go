package resource

import (
	"errors"
	"fmt"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

func nodeService(reqEvent *rsML.ServiceEvent) error {
	resEvent := &rsML.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsML.CommandGet:
		data, err := getNode(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsML.CommandSet:
		node := &nodeML.Node{}
		err := reqEvent.LoadData(node)
		if err != nil {
			zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
			return err
		}

		nodeOrg, err := nodeAPI.GetByID(node.ID)
		if err != nil {
			return err
		}
		// update labels
		nodeOrg.Labels.CopyFrom(node.Labels)
		return nodeAPI.Save(nodeOrg)

	case rsML.CommandGetIds:
		data, err := getNodeIDs(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsML.CommandFirmwareState:
		fwState := make(map[string]interface{})
		err := reqEvent.LoadData(&fwState)
		if err != nil {
			zap.L().Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
			return err
		}
		if fwState == nil {
			zap.L().Error("nil data received", zap.Any("data", reqEvent))
			return fmt.Errorf("nil data received")
		}
		return nodeAPI.UpdateFirmwareState(reqEvent.ID, fwState)

	case rsML.CommandSetLabel:
		labels := cmap.CustomStringMap{}
		err := reqEvent.LoadData(&labels)
		if err != nil {
			zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
			return err
		}
		node, err := getNode(reqEvent)
		if err != nil {
			return err
		}
		node.Labels.CopyFrom(labels)
		return nodeAPI.Save(node)

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getNodeIDs(request *rsML.ServiceEvent) ([]string, error) {
	var response *storage.Result
	if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := nodeAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		response = result
	} else {
		ids := make(map[string]interface{})
		err := request.LoadData(&ids)
		if err != nil {
			zap.L().Error("error on data conversion", zap.Any("request", request), zap.Error(err))
			return nil, err
		}

		// get NodeId and GatewayId
		gatewayId := utils.GetMapValueString(ids, model.KeyGatewayID, "")
		if gatewayId == "" {
			return nil, fmt.Errorf("%v not supplied", model.KeyGatewayID)
		}
		filters := []storage.Filter{{Key: model.KeyGatewayID, Operator: storage.OperatorEqual, Value: gatewayId}}
		result, err := nodeAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		response = result
	}

	if response == nil || response.Data == nil {
		return nil, errors.New("nil data supplied")
	}
	nodeIDs := make([]string, 0)
	if nodes, ok := response.Data.(*[]nodeML.Node); ok {
		for _, node := range *nodes {
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}
	return nodeIDs, nil
}

func getNode(request *rsML.ServiceEvent) (*nodeML.Node, error) {
	if request.ID != "" {
		cfg, err := nodeAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil

	} else {
		ids := make(map[string]interface{})
		err := request.LoadData(&ids)
		if err != nil {
			zap.L().Error("error on data conversion", zap.Any("request", request), zap.Error(err))
			return nil, err
		}

		// get NodeId and GatewayId
		nodeId := utils.GetMapValueString(ids, model.KeyNodeID, "")
		gatewayId := utils.GetMapValueString(ids, model.KeyGatewayID, "")
		return nodeAPI.GetByGatewayAndNodeID(gatewayId, nodeId)
	}
}
