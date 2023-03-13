package resource

import (
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

func (svc *ResourceService) nodeService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getNode(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandSet:
		node := &nodeTY.Node{}
		err := reqEvent.LoadData(node)
		if err != nil {
			svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
			return err
		}
		_, err = svc.api.Node().GetByID(node.ID)
		if err != nil {
			return err
		}
		node.Labels.Init()
		node.Others.Init()
		return svc.api.Node().Save(node, true)

	case rsTY.CommandGetIds:
		data, err := svc.getNodeIDs(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandFirmwareState:
		fwState := make(map[string]interface{})
		err := reqEvent.LoadData(&fwState)
		if err != nil {
			svc.logger.Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
			return err
		}
		if fwState == nil {
			svc.logger.Error("nil data received", zap.Any("data", reqEvent))
			return fmt.Errorf("nil data received")
		}
		return svc.api.Node().UpdateFirmwareState(reqEvent.ID, fwState)

	case rsTY.CommandSetLabel:
		labels := cmap.CustomStringMap{}
		err := reqEvent.LoadData(&labels)
		if err != nil {
			svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
			return err
		}
		node, err := svc.getNode(reqEvent)
		if err != nil {
			return err
		}
		node.Labels.CopyFrom(labels)
		return svc.api.Node().Save(node, true)

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getNodeIDs(request *rsTY.ServiceEvent) ([]string, error) {
	var response *storageTY.Result
	if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Node().List(filters, nil)
		if err != nil {
			return nil, err
		}
		response = result
	} else {
		ids := make(map[string]interface{})
		err := request.LoadData(&ids)
		if err != nil {
			svc.logger.Error("error on data conversion", zap.Any("request", request), zap.Error(err))
			return nil, err
		}

		// get NodeId and GatewayId
		gatewayId := utils.GetMapValueString(ids, types.KeyGatewayID, "")
		if gatewayId == "" {
			return nil, fmt.Errorf("%v not supplied", types.KeyGatewayID)
		}
		filters := []storageTY.Filter{{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: gatewayId}}
		result, err := svc.api.Node().List(filters, nil)
		if err != nil {
			return nil, err
		}
		response = result
	}

	if response == nil || response.Data == nil {
		return nil, errors.New("nil data supplied")
	}
	nodeIDs := make([]string, 0)
	if nodes, ok := response.Data.(*[]nodeTY.Node); ok {
		for _, node := range *nodes {
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}
	return nodeIDs, nil
}

func (svc *ResourceService) getNode(request *rsTY.ServiceEvent) (*nodeTY.Node, error) {
	if request.ID != "" {
		cfg, err := svc.api.Node().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil

	} else {
		ids := make(map[string]interface{})
		err := request.LoadData(&ids)
		if err != nil {
			svc.logger.Error("error on data conversion", zap.Any("request", request), zap.Error(err))
			return nil, err
		}

		// get NodeId and GatewayId
		nodeId := utils.GetMapValueString(ids, types.KeyNodeID, "")
		gatewayId := utils.GetMapValueString(ids, types.KeyGatewayID, "")
		return svc.api.Node().GetByGatewayAndNodeID(gatewayId, nodeId)
	}
}
