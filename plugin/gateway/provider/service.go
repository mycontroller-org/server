package provider

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	cloneUtils "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwPlugin "github.com/mycontroller-org/server/v2/plugin/gateway"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	messageQueueLimit         = 200
	rawMessageQueueLimit      = 200
	sleepingQueuePerNodeLimit = 20 // number of messages can be in sleeping queue per node
	messageWorkersCount       = 1
	rawMessageWorkersCount    = 1

	defaultValidity = time.Hour * 24 // by default all the sleeping queue messages are valid for 24 hours

	defaultReconnectDelay = "15s"
	loggerName            = "gateway_service"
)

// Service component of the provider
type Service struct {
	GatewayConfig        *gwTY.Config
	provider             providerTY.Plugin
	messageQueue         *queueUtils.QueueSpec
	rawMessageQueue      *queueUtils.QueueSpec
	sleepingMessageQueue map[string][]msgTY.Message
	mutex                *sync.RWMutex
	ctx                  context.Context
	logger               *zap.Logger
	bus                  busTY.Plugin
}

// GetService returns service instance
func GetService(ctx context.Context, gatewayCfg *gwTY.Config) (*Service, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// verify default reconnect delay
	gatewayCfg.ReconnectDelay = utils.ValidDuration(gatewayCfg.ReconnectDelay, defaultReconnectDelay)

	// get a plugin
	provider, err := gwPlugin.Create(ctx, gatewayCfg.Provider.GetString(types.KeyType), gatewayCfg)
	if err != nil {
		return nil, err
	}
	s := &Service{
		GatewayConfig: gatewayCfg,
		provider:      provider,
		ctx:           ctx,
		mutex:         &sync.RWMutex{},
		logger:        logger.Named(loggerName),
		bus:           bus,
	}

	// message queues
	s.rawMessageQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(s.logger, fmt.Sprintf("gw_provider_raw_msg_%s", s.GatewayConfig.ID), rawMessageQueueLimit, s.rawMessageProcessor, rawMessageWorkersCount),
		Topic:          topic.TopicPostMessageToProcessor,
		SubscriptionId: -1,
	}
	s.messageQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(s.logger, fmt.Sprintf("gw_provider_msg_%s", s.GatewayConfig.ID), messageQueueLimit, s.messageConsumer, messageWorkersCount),
		Topic:          fmt.Sprintf("%s.%s", topic.TopicPostMessageToProvider, s.GatewayConfig.ID),
		SubscriptionId: -1,
	}

	return s, nil
}

// Start gateway service
func (s *Service) Start() error {
	s.logger.Debug("starting a provider service", zap.String("gatewayId", s.GatewayConfig.ID), zap.String("providerType", s.GatewayConfig.Provider.GetString(types.NameType)))

	s.sleepingMessageQueue = make(map[string][]msgTY.Message)

	// start message listener
	s.startMessageListener()

	// start provider
	err := s.provider.Start(s.rawMessageQueueProduceFunc)
	if err != nil {
		s.logger.Error("error", zap.Error(err))
		err := s.bus.Unsubscribe(s.messageQueue.Topic, s.messageQueue.SubscriptionId)
		if err != nil {
			s.logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", s.messageQueue.Topic))
		}

		s.messageQueue.Close()
		s.rawMessageQueue.Close()
	} else {
		go s.startValidityCheckFn() // run the validity check in a go routine
	}

	return err
}

// unsubscribe and stop queues
func (s *Service) stopService() {
	err := s.bus.Unsubscribe(s.messageQueue.Topic, s.messageQueue.SubscriptionId)
	if err != nil {
		s.logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", s.messageQueue.Topic))
	}
	s.messageQueue.Close()
	s.rawMessageQueue.Close()
}

// Stop the service
func (s *Service) Stop() error {
	defer s.stopService()     // in any case when exit, call stopService
	err := s.provider.Close() // close protocol connection
	if err != nil {
		return err
	}
	return nil
}

// this function supplied to provider-protocol
// rawMessages will be added directly to here
func (s *Service) rawMessageQueueProduceFunc(rawMsg *msgTY.RawMessage) error {
	status := s.rawMessageQueue.Produce(rawMsg)
	if !status {
		return errors.New("failed to add rawMessage in to queue")
	}
	return nil
}

// listens messages from server
func (s *Service) startMessageListener() {
	subscriptionID, err := s.bus.Subscribe(s.messageQueue.Topic, func(event *busTY.BusData) {
		msg := &msgTY.Message{}
		err := event.LoadData(msg)
		if err != nil {
			s.logger.Warn("received invalid type", zap.Any("event", event))
			return
		}
		if msg.GatewayID == "" {
			s.logger.Warn("received message with empty gatewayId", zap.Any("message", msg))
			return
		}
		s.messageQueue.Produce(msg)
	})

	if err != nil {
		s.logger.Error("error on subscription", zap.String("topic", s.messageQueue.Topic), zap.Error(err))
	} else {
		s.messageQueue.SubscriptionId = subscriptionID
	}
}

func (s *Service) messageConsumer(item interface{}) error {
	msg, ok := item.(*msgTY.Message)
	if !ok {
		s.logger.Error("invalid message type", zap.Any("received", item))
		return nil // Don't requeue invalid messages
	}

	// include timestamp, if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// if it is awake message send the sleeping queue messages
	if msg.Type == msgTY.TypeAction && len(msg.Payloads) > 0 && msg.Payloads[0].Value == nodeTY.ActionAwake {
		s.publishSleepingMessageQueue(msg.NodeID)
		return nil
	} else if msg.IsSleepNode {
		// for sleeping node message to be added in to the sleeping queue
		// when node sends awake signal, queue message will be published
		s.addToSleepingMessageQueue(msg)
		return nil
	} else {
		return s.postMessage(msg, s.GatewayConfig.QueueFailedMessage)
	}
}

// postMessage to the provider
func (s *Service) postMessage(msg *msgTY.Message, queueFailedMessage bool) error {
	err := s.provider.Post(msg)
	if err != nil {
		s.logger.Warn("error on sending", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
		if queueFailedMessage {
			msg.Labels = msg.Labels.Init()
			isSleepingQueueDisabled := msg.Labels.IsExists(types.LabelNodeSleepQueueDisabled) && msg.Labels.GetBool(types.LabelNodeSleepQueueDisabled)
			if !isSleepingQueueDisabled {
				s.addToSleepingMessageQueue(msg)
				return nil // Message was queued, no need to requeue
			}
			return err // Return error to requeue the message
		}
	}
	return nil
}

// process received raw messages from protocol
func (s *Service) rawMessageProcessor(data interface{}) error {
	rawMsg := data.(*msgTY.RawMessage)
	s.logger.Debug("received rawMessage", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
	messages, err := s.provider.ConvertToMessages(rawMsg)
	if err != nil {
		s.logger.Warn("failed to parse", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
		return nil // Don't requeue unparseable messages
	}
	if len(messages) == 0 {
		s.logger.Debug("messages not parsed", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
		return nil
	}
	// update gatewayID if not found
	for index := 0; index < len(messages); index++ {
		msg := messages[index]
		if msg != nil {
			if msg != nil && msg.GatewayID == "" {
				msg.GatewayID = s.GatewayConfig.ID
			}
			err = s.bus.Publish(s.rawMessageQueue.Topic, msg)
			if err != nil {
				s.logger.Debug("failed to post on topic", zap.String("topic", s.rawMessageQueue.Topic), zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
				return err // Return error to requeue the message
			}
		}
	}
	return nil
}

// add message in to sleeping queue
func (s *Service) addToSleepingMessageQueue(msg *msgTY.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// add into sleeping queue
	queue, ok := s.sleepingMessageQueue[msg.NodeID]
	if !ok {
		queue = make([]msgTY.Message, 0)
		s.sleepingMessageQueue[msg.NodeID] = queue
	}
	// verify if the message already in the queue, if then remove it
	newMsgId := msg.GetID()
	for index, oldMsg := range queue {
		if newMsgId == oldMsg.GetID() {
			queue[index] = *msg
			queue = append(queue[:index], queue[index+1:]...)
			break
		}
	}

	// add it to the queue
	queue = append(queue, *msg)

	// if queue size exceeds maximum defined size, do resize
	oldSize := len(queue)
	if oldSize > sleepingQueuePerNodeLimit {
		queue = queue[:sleepingQueuePerNodeLimit]
		s.logger.Debug("dropped messages from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)), zap.String("gatewayID", msg.GatewayID), zap.String("nodeID", msg.NodeID))
	}
	s.sleepingMessageQueue[msg.NodeID] = queue
}

// emptySleepingMessageQueue clears all the messages on the queue
func (s *Service) publishSleepingMessageQueue(nodeID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msgQueue, ok := s.sleepingMessageQueue[nodeID]
	if !ok {
		return
	}
	validMsgs := s.getValidMessages(msgQueue)

	// post messages
	for _, msg := range validMsgs {
		_ = s.postMessage(&msg, false) // Ignore error for sleeping queue messages
	}

	// remove messages from the map
	s.sleepingMessageQueue[nodeID] = make([]msgTY.Message, 0)
}

// returns sleeping queue messages
func (s *Service) GetGatewaySleepingQueue() map[string][]msgTY.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// clone the queue and return
	clonedQueue := cloneUtils.Clone(s.sleepingMessageQueue)
	return clonedQueue.(map[string][]msgTY.Message)
}

// returns sleeping queue messages
func (s *Service) GetNodeSleepingQueue(nodeID string) []msgTY.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msgQueue, ok := s.sleepingMessageQueue[nodeID]
	if !ok {
		return make([]msgTY.Message, 0)
	}
	validMsgs := s.getValidMessages(msgQueue)
	s.sleepingMessageQueue[nodeID] = validMsgs
	// clone the queue and return
	clonedQueue := cloneUtils.Clone(validMsgs)
	return clonedQueue.([]msgTY.Message)
}

// clear all sleeping messages
func (s *Service) ClearGatewaySleepingQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// clear all the messages
	s.sleepingMessageQueue = make(map[string][]msgTY.Message)
}

// clear sleeping queue message of a node
func (s *Service) ClearNodeSleepingQueue(nodeID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// remove messages for a node
	s.sleepingMessageQueue[nodeID] = make([]msgTY.Message, 0)
}

// verifies the validity of the message in queue, if expires remove them from the queue
func (s *Service) checkValidity() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for nodeId, msgs := range s.sleepingMessageQueue {
		s.sleepingMessageQueue[nodeId] = s.getValidMessages(msgs)
	}
}

func (s *Service) getValidMessages(msgs []msgTY.Message) []msgTY.Message {
	updatedMsgs := []msgTY.Message{}
	for _, msg := range msgs {
		if s.isValid(msg) {
			updatedMsgs = append(updatedMsgs, msg)
		}
	}
	return updatedMsgs
}

func (s *Service) isValid(msg msgTY.Message) bool {
	validity := utils.ToDuration(msg.Validity, defaultValidity)
	now := time.Now()
	var validTil time.Time
	if msg.Timestamp.IsZero() {
		validTil = now.Add(validity)
	} else {
		validTil = msg.Timestamp.Add(validity)
	}

	return now.Before(validTil)
}

func (s *Service) startValidityCheckFn() {
	// Create a context that cancels on OS signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ticker := time.NewTicker(time.Minute * 10) // execute 10 minutes once
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("received exit signal")
			return
		case <-ticker.C:
			s.checkValidity()
		}
	}
}
