package model

import (
	"sync"
	"time"

	"github.com/jaegertracing/jaeger/pkg/queue"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	"go.uber.org/zap"
)

// Gateway Types
const (
	GatewayTypeMQTT     = "mqtt"
	GatewayTypeSerial   = "serial"
	GatewayTypeEthernet = "ethernet"
)

// AckConfig data
type AckConfig struct {
	Enabled       bool   `json:"enabled"`
	StreamEnabled bool   `json:"streamEnabled"`
	RetryCount    int    `json:"retryCount"`
	WaitTime      string `json:"waitTime"`
}

// Gateway providers
const (
	GatewayProviderMySensors = "MySensors"
)

// GatewayProvider data
type GatewayProvider struct {
	Type        string                 `json:"type"`
	GatewayType string                 `json:"gatewayType"`
	Config      map[string]interface{} `json:"config"`
}

// GatewayConfigMQTT data
type GatewayConfigMQTT struct {
	Broker    string
	Subscribe string
	Publish   string
	QoS       int
	Username  string
	Password  string `json:"-"`
}

// GatewayConfig entity
type GatewayConfig struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	AckConfig AckConfig       `json:"ackConfig"`
	State     State           `json:"state"`
	Provider  GatewayProvider `json:"providerConfig"`
	LastSeen  time.Time       `json:"lastSeen"`
}

// GatewayMessageParser interface for provider
type GatewayMessageParser interface {
	ToRawMessage(message *msg.Message) (*msg.RawMessage, error)
	ToMessage(rawMesage *msg.RawMessage) (*msg.Message, error)
}

// Gateway instance
type Gateway interface {
	Close() error
	Write(rawMessage *msg.RawMessage) error
}

// GatewayService details
type GatewayService struct {
	Config              *GatewayConfig
	Parser              GatewayMessageParser
	Gateway             Gateway
	TxMsgQueue          *queue.BoundedQueue
	TopicMsg2GW         string
	TopicSleepingMsg2GW string
	SleepMsgQueue       map[string][]*msg.Message
	mutex               sync.RWMutex
}

// AddSleepMsg into queue
func (s *GatewayService) AddSleepMsg(mcMsg *msg.Message, limit int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// add into sleeping queue
	queue, ok := s.SleepMsgQueue[mcMsg.NodeID]
	if !ok {
		queue = make([]*msg.Message, 0)
		s.SleepMsgQueue[mcMsg.NodeID] = queue
	}
	queue = append(queue, mcMsg)
	// if queue size exceeds maximum defined size, do resize
	oldSize := len(queue)
	if oldSize > limit {
		queue = queue[:limit]
		zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
	}
}

// ClearSleepingQueue clears all the messages on the queue
func (s *GatewayService) ClearSleepingQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SleepMsgQueue = make(map[string][]*msg.Message)
}

// GetSleepingQueue returns message for a specific nodeID, also removes it from the queue
func (s *GatewayService) GetSleepingQueue(nodeID string) []*msg.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if queue, ok := s.SleepMsgQueue[nodeID]; ok {
		s.SleepMsgQueue[nodeID] = make([]*msg.Message, 0)
		return queue
	}
	return nil
}
