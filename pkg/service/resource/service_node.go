package resource

import (
	"errors"
	"fmt"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

func nodeService(reqEvent *rsModel.Event) error {
	resEvent := &rsModel.Event{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getNode(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}
		zap.L().Info("node sent", zap.String("bytes", string(resEvent.Data)))

	case rsModel.CommandSet:
		node := &nodeML.Node{}
		err := reqEvent.ToStruct(node)
		if err != nil {
			return err
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
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getNodeIDs(request *rsModel.Event) ([]string, error) {
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
		err := request.ToStruct(&ids)
		if err != nil {
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
	if nodes, ok := response.Data.([]nodeML.Node); ok {
		for _, node := range nodes {
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}
	return nodeIDs, nil
}

func getNode(request *rsModel.Event) (interface{}, error) {
	if request.ID != "" {
		cfg, err := nodeAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := nodeAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	} else {
		ids := make(map[string]interface{})
		err := request.ToStruct(&ids)
		if err != nil {
			return nil, err
		}
		// get NodeId and GatewayId
		nodeId := utils.GetMapValueString(ids, model.KeyNodeID, "")
		gatewayId := utils.GetMapValueString(ids, model.KeyGatewayID, "")
		return nodeAPI.GetByGatewayAndNodeID(gatewayId, nodeId)
	}
}
