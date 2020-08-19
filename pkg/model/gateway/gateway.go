package gateway

import (
	"sync"
	"time"

	"github.com/jaegertracing/jaeger/pkg/queue"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	"go.uber.org/zap"
)

// Global constants
const (
	// Gateway Types
	TypeMQTT     = "mqtt"
	TypeSerial   = "serial"
	TypeEthernet = "ethernet"

	// Gateway providers
	ProviderMySensors = "MySensors"
)

// Config struct
type Config struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Enabled   bool              `json:"enabled"`
	AckConfig AckConfig         `json:"ackConfig"`
	State     ml.State          `json:"state"`
	Provider  Provider          `json:"providerConfig"`
	LastSeen  time.Time         `json:"lastSeen"`
	Labels    map[string]string `json:"labels"`
}

// AckConfig data
type AckConfig struct {
	Enabled          bool   `json:"enabled"`
	StreamAckEnabled bool   `json:"streamAckEnabled"`
	RetryCount       int    `json:"retryCount"`
	Timeout          string `json:"timeout"`
}

// Provider data
type Provider struct {
	Type        string                 `json:"type"`
	GatewayType string                 `json:"gatewayType"`
	Config      map[string]interface{} `json:"config"`
}

// ConfigMQTT data
type ConfigMQTT struct {
	Broker    string `json:"broker"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Subscribe string `json:"subscribe"`
	Publish   string `json:"publish"`
	QoS       int    `json:"qos"`
}

// MessageParser interface for provider
type MessageParser interface {
	ToRawMessage(message *msg.Message) (*msg.RawMessage, error)
	ToMessage(rawMesage *msg.RawMessage) (*msg.Message, error)
}

// Gateway instance
type Gateway interface {
	Close() error
	Write(rawMessage *msg.RawMessage) error
}

// Service details
type Service struct {
	Config              *Config
	Parser              MessageParser
	Gateway             Gateway
	TxMsgQueue          *queue.BoundedQueue
	TopicMsg2GW         string
	TopicSleepingMsg2GW string
	SleepMsgQueue       map[string][]*msg.Message
	mutex               sync.RWMutex
}

// AddSleepMsg into queue
func (s *Service) AddSleepMsg(mcMsg *msg.Message, limit int) {
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
func (s *Service) ClearSleepingQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SleepMsgQueue = make(map[string][]*msg.Message)
}

// GetSleepingQueue returns message for a specific nodeID, also removes it from the queue
func (s *Service) GetSleepingQueue(nodeID string) []*msg.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if queue, ok := s.SleepMsgQueue[nodeID]; ok {
		s.SleepMsgQueue[nodeID] = make([]*msg.Message, 0)
		return queue
	}
	return nil
}
