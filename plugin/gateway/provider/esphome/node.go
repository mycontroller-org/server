package esphome

import (
	"bytes"
	"fmt"
	"time"

	esphomeAPI "github.com/mycontroller-org/esphome_api/pkg/api"
	esphomeClient "github.com/mycontroller-org/esphome_api/pkg/client"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// NewESPHomeNode creates a esphome node instance
func NewESPHomeNode(gatewayID, nodeID string, config ESPHomeNodeConfig, entityStore *EntityStore, rxMessageFunc func(rawMsg *msgTY.RawMessage) error) *ESPHomeNode {
	client := &ESPHomeNode{
		GatewayID:     gatewayID,
		NodeID:        nodeID,
		Config:        config,
		entityStore:   entityStore,
		rxMessageFunc: rxMessageFunc,
		imageBuffer:   new(bytes.Buffer),
	}
	return client
}

// Connect establish a connection to a esphome node
// performs: login, subscribe states, list available entities and updates node details
func (en *ESPHomeNode) Connect() error {
	timeoutDuration := utils.ToDuration(en.Config.Timeout, 10*time.Second)
	clientID := fmt.Sprintf("%s_%s", en.GatewayID, utils.RandIDWithLength(3))
	ehClient, err := esphomeClient.Init(clientID, en.Config.Address, timeoutDuration, en.onReceive)
	if err != nil {
		return err
	}
	err = ehClient.Login(en.Config.Password)
	if err != nil {
		return err
	}

	err = ehClient.SubscribeStates()
	if err != nil {
		return err
	}

	err = ehClient.ListEntities()
	if err != nil {
		return err
	}

	en.Client = ehClient

	en.scheduleAliveCheck()
	en.sendNodeInfo()
	return nil
}

// doAliveCheck verifies the aliveness of a esphome node
func (en *ESPHomeNode) doAliveCheck() {
	err := en.Client.Ping()
	if err != nil {
		Unschedule(en.aliveScheduleID())
		en.ScheduleReconnect()
		zap.L().Info("error on ping, reconnect scheduled", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.String("error", err.Error()), zap.String("reconnectDelay", en.Config.ReconnectDelay))
	}
}

// Disconnect performs logout from esphome node
func (en *ESPHomeNode) Disconnect() error {
	Unschedule(en.aliveScheduleID())
	Unschedule(en.reconnectScheduleID())
	if en.Client != nil {
		return en.Client.Close()
	}
	return nil
}

// Post sends message to a esphome node
func (en *ESPHomeNode) Post(msg proto.Message) error {
	zap.L().Debug("posting a message", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Any("msg", msg), zap.String("type", fmt.Sprintf("%T", msg)))
	err := en.Client.Send(msg)
	if err != nil {
		return err
	}
	// There is no acknowledgement received for stream set action
	// generate a successfully acknowledgement to mycontroller
	if imageReq, ok := msg.(*esphomeAPI.CameraImageRequest); ok {
		entity := en.entityStore.GetByEntityType(en.NodeID, EntityTypeCamera)
		if entity == nil {
			return nil
		}
		streamResponse := getMessage(en.GatewayID, en.NodeID, entity.SourceID, msgTY.TypeSet, time.Now())
		data := msgTY.NewPayload()
		data.Key = FieldStream
		data.Value = convertor.ToString(imageReq.Stream)
		data.MetricType = metricTY.MetricTypeNone
		streamResponse.Payloads = append(streamResponse.Payloads, data)
		topic := mcbus.GetTopicPostMessageToServer()
		err = mcbus.Publish(topic, streamResponse)
		if err != nil {
			zap.L().Error("error on posting a message", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
		}
	}
	return nil
}

// onReceive used to send received proto messages to the queue
func (en *ESPHomeNode) onReceive(msg proto.Message) {
	// check the message type and process camera image response locally
	// hold all the camera image data
	// post to next service when it receives the complete image
	if cameraMsg, ok := msg.(*esphomeAPI.CameraImageResponse); ok {
		en.imageBuffer.Write(cameraMsg.Data)
		if cameraMsg.Done {
			cameraMsg.Data = en.imageBuffer.Bytes()
			en.imageBuffer.Reset()
		} else {
			return
		}
	}

	// create a raw message and post into queue
	rawMsg := msgTY.NewRawMessage(true, nil)
	rawMsg.Data = msg
	msgTypeID := esphomeAPI.TypeID(msg)
	rawMsg.Others.Set(MessageTypeID, msgTypeID, nil)
	rawMsg.Others.Set(NodeID, en.NodeID, nil)

	err := en.rxMessageFunc(rawMsg)
	if err != nil {
		zap.L().Error("error on posting a message", zap.Error(err), zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Any("message", msg))
	}
}

// reconnect performs a reconnection
// if error happens on a connection will be retry again after a reconnect delay
func (en *ESPHomeNode) reconnect() {
	if en.Client != nil {
		err := en.Client.Close()
		if err != nil {
			zap.L().Debug("error on disconnect", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.String("error", err.Error()))
		}
	}

	err := en.Connect()
	if err != nil {
		zap.L().Debug("error on reconnect", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.String("error", err.Error()))
	} else {
		Unschedule(en.reconnectScheduleID())
	}
}

// scheduleAliveCheck adds a schedule for alivecheck
func (en *ESPHomeNode) scheduleAliveCheck() {
	err := Schedule(en.aliveScheduleID(), en.Config.AliveCheckInterval, en.doAliveCheck)
	if err != nil {
		zap.L().Error("error on configure alive check interval", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
	}
}

// ScheduleReconnect adds a schedule for a reconnect
func (en *ESPHomeNode) ScheduleReconnect() {
	err := Schedule(en.reconnectScheduleID(), en.Config.ReconnectDelay, en.reconnect)
	if err != nil {
		zap.L().Error("error on configure reconnect", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
	}
}

// aliveScheduleID returns a schedule id for alivecheck job
func (en *ESPHomeNode) aliveScheduleID() string {
	return fmt.Sprintf("%s_%s_%s_alive_check", schedulePrefix, en.GatewayID, en.NodeID)
}

// reconnectScheduleID returns a schedule id for reconnect job
func (en *ESPHomeNode) reconnectScheduleID() string {
	return fmt.Sprintf("%s_%s_%s_reconnect", schedulePrefix, en.GatewayID, en.NodeID)
}

// sendNodeInfo sends node details directly to mycontroller
func (en *ESPHomeNode) sendNodeInfo() {
	deviceInfo, err := en.Client.DeviceInfo()
	if err != nil {
		zap.L().Error("error on getting device info", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
		return
	}

	nodeMsg := getMessage(en.GatewayID, en.NodeID, "", msgTY.TypePresentation, time.Now())
	data := msgTY.NewPayload()
	data.Key = types.FieldName
	data.Value = deviceInfo.Name
	data.Labels.Set(types.LabelNodeVersion, deviceInfo.EsphomeVersion)
	data.Others.Set("mac", deviceInfo.MacAddress, nil)
	data.Others.Set("compilation_time", deviceInfo.CompilationTime, nil)
	data.Others.Set("has_deep_sleep", deviceInfo.HasDeepSleep, nil)
	data.Others.Set("model", deviceInfo.Model, nil)
	data.Others.Set("uses_password", deviceInfo.UsesPassword, nil)

	nodeMsg.Payloads = append(nodeMsg.Payloads, data)

	topic := mcbus.GetTopicPostMessageToServer()
	err = mcbus.Publish(topic, nodeMsg)
	if err != nil {
		zap.L().Error("error on posting a message", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
	}
}

// sendRestartRequest sends a restat request to esphome node
func (en *ESPHomeNode) sendRestartRequest() {
	entity := en.entityStore.GetBySourceID(en.NodeID, SourceIDRestart)
	if entity != nil {
		restartRequest := &esphomeAPI.SwitchCommandRequest{State: true, Key: entity.Key}
		err := en.Post(restartRequest)
		if err != nil {
			zap.L().Error("error on sending restart request", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID), zap.Error(err))
		}
		return
	}
	zap.L().Info("seems 'restart' switch is not configured for this node", zap.String("gatewayId", en.GatewayID), zap.String("nodeId", en.NodeID))
}
