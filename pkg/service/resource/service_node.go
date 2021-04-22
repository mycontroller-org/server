package resource

import (
	"errors"
	"fmt"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/storage"
)

func nodeService(reqEvent *rsModel.ServiceEvent) error {
	resEvent := &rsModel.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getNode(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsModel.CommandSet:
		node, ok := reqEvent.GetData().(nodeML.Node)
		if !ok {
			return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
		}

		nodeOrg, err := nodeAPI.GetByID(node.ID)
		if err != nil {
			return err
		}
		// update labels
		nodeOrg.Labels.CopyFrom(node.Labels)
		return nodeAPI.Save(nodeOrg)

	case rsModel.CommandGetIds:
		data, err := getNodeIDs(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsModel.CommandFirmwareState:
		fwState, ok := reqEvent.GetData().(map[string]interface{})
		if !ok {
			return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
		}
		return nodeAPI.UpdateFirmwareState(reqEvent.ID, fwState)

	case rsModel.CommandSetLabel:
		labels, ok := reqEvent.GetData().(cmap.CustomStringMap)
		if !ok {
			return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
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

func getNodeIDs(request *rsModel.ServiceEvent) ([]string, error) {
	var response *storage.Result
	if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := nodeAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		response = result
	} else {
		ids, ok := request.GetData().(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error on data conversion, receivedType: %T", request.GetData())
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
	if nodes, ok := response.Data.([]nodeML.Node); ok {
		for _, node := range nodes {
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}
	return nodeIDs, nil
}

func getNode(request *rsModel.ServiceEvent) (*nodeML.Node, error) {
	if request.ID != "" {
		cfg, err := nodeAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil

	} else {
		ids, ok := request.GetData().(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error on data conversion, receivedType: %T", request.GetData())
		}

		// get NodeId and GatewayId
		nodeId := utils.GetMapValueString(ids, model.KeyNodeID, "")
		gatewayId := utils.GetMapValueString(ids, model.KeyGatewayID, "")
		return nodeAPI.GetByGatewayAndNodeID(gatewayId, nodeId)
	}
}
