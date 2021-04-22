package mysensors

import (
	"bytes"
	"context"
	"encoding/binary"
	hexENC "encoding/hex"
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

var (
	fwStore   = concurrency.NewStore()
	nodeStore = concurrency.NewStore()
)

func firmwarePurge() {
	for _, fwID := range fwStore.Keys() {
		fwInf := fwStore.Get(fwID)
		fw, ok := fwInf.(*firmwareRaw)
		if !ok {
			continue
		}
		if time.Since(fw.LastAccess) >= firmwarePurgeInactiveTime { // eligible for purging
			fwStore.Remove(fwID)
		}
	}
}

// getNode returns the node
func getNode(gatewayID, nodeID string) (*nodeML.Node, error) {
	id := getNodeStoreID(gatewayID, nodeID)

	toNode := func(item interface{}) (*nodeML.Node, error) {
		if node, ok := item.(*nodeML.Node); ok {
			return node, nil
		}
		return nil, fmt.Errorf("unknown data received in the place node: %T", item)
	}

	data := nodeStore.Get(id)
	if data != nil {
		return toNode(data)
	}

	err := updateNode(gatewayID, nodeID)
	if err != nil {
		return nil, err
	}
	data = nodeStore.Get(id)
	if data != nil {
		return toNode(data)
	}
	return nil, fmt.Errorf("node not available. gatewayID:%s, nodeID:%s", gatewayID, nodeID)
}

func setNodeLabels(node *nodeML.Node) {
	busUtils.PostToResourceService(node.ID, node, rsML.TypeNode, rsML.CommandSet, "")
}

// toHex returns hex string
func toHex(in interface{}) (string, error) {
	var bBuf bytes.Buffer
	err := binary.Write(&bBuf, binary.LittleEndian, in)
	if err != nil {
		return "", err
	}
	return hexENC.EncodeToString(bBuf.Bytes()), nil
}

// toStruct updates struct from hex string
func toStruct(hex string, out interface{}) error {
	hb, err := hexENC.DecodeString(hex)
	if err != nil {
		return err
	}
	r := bytes.NewReader(hb)
	return binary.Read(r, binary.LittleEndian, out)
}

func getNodeStoreID(gatewayID, nodeID string) string {
	return fmt.Sprintf("%s_%s", gatewayID, nodeID)
}

// get node details via bus
func updateNode(gatewayID, nodeID string) error {
	closeChan := concurrency.NewChannel(0)
	defer closeChan.SafeClose()

	replyTopic := mcbus.FormatTopic(fmt.Sprintf("node_response_%s", utils.RandIDWithLength(5)))
	sID, err := mcbus.Subscribe(replyTopic, nodeResponse(closeChan))
	if err != nil {
		return err
	}

	defer func() {
		err := mcbus.Unsubscribe(replyTopic, sID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", replyTopic))
		}
	}()

	timeoutDuration := 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	ids := map[string]interface{}{
		model.KeyGatewayID: gatewayID,
		model.KeyNodeID:    nodeID,
	}
	busUtils.PostToResourceService("", ids, rsML.TypeNode, rsML.CommandGet, replyTopic)

	select {
	case <-closeChan.CH:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("reached timeout: %s", timeoutDuration.String())
	}
}

func nodeResponse(closeChan *concurrency.Channel) func(data *busML.BusData) {
	return func(data *busML.BusData) {
		defer closeChan.SafeClose()

		node := &nodeML.Node{}
		err := data.ToStruct(node)
		if err != nil {
			zap.L().Error("error on converting to target type", zap.Error(err))
		}
		zap.L().Info("node received", zap.Any("node", node))

		nodeStore.Add(getNodeStoreID(node.GatewayID, node.NodeID), node)
		zap.L().Info("node added", zap.String("id", getNodeStoreID(node.GatewayID, node.NodeID)))

	}
}
