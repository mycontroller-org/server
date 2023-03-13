package service

import (
	"context"
	"fmt"
	"time"

	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwProvider "github.com/mycontroller-org/server/v2/plugin/gateway/provider"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	defaultQueueSize = int(50)
	defaultWorkers   = int(1)
)

type GatewayService struct {
	ctx         context.Context
	logger      *zap.Logger
	store       *Store
	bus         busTY.Plugin
	eventsQueue *queueUtils.QueueSpec
	filter      *sfTY.ServiceFilter
	enc         *encryptionAPI.Encryption
}

func New(ctx context.Context, filter *sfTY.ServiceFilter) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &GatewayService{
		ctx:    ctx,
		logger: logger.Named("gateway_service"),
		store:  &Store{services: make(map[string]*gwProvider.Service)},
		bus:    bus,
		filter: filter,
		enc:    enc,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Topic:          topic.TopicServiceGateway,
		Queue:          queueUtils.New(svc.logger, "gateway", defaultQueueSize, svc.processEvent, defaultWorkers),
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *GatewayService) Name() string {
	return "gateway_service"
}

// startGW gateway
func (svc *GatewayService) startGW(gatewayCfg *gwTY.Config) error {
	start := time.Now()

	// decrypt the secrets, tokens
	err := svc.enc.DecryptSecrets(gatewayCfg)
	if err != nil {
		return err
	}

	if svc.store.Get(gatewayCfg.ID) != nil {
		svc.logger.Info("no action needed. gateway service is in running state.", zap.String("gatewayId", gatewayCfg.ID))
		return nil
	}
	if !gatewayCfg.Enabled { // this gateway is not enabled
		return nil
	}
	svc.logger.Info("starting a gateway", zap.Any("id", gatewayCfg.ID))
	state := types.State{Since: time.Now()}

	service, err := gwProvider.GetService(svc.ctx, gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		svc.logger.Error("failed to start a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		svc.logger.Info("started a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = types.StatusUp
		svc.store.Add(service)
	}

	busUtils.SetGatewayState(svc.logger, svc.bus, gatewayCfg.ID, state)
	return nil
}

// stopGW gateway
func (svc *GatewayService) stopGW(id string) error {
	start := time.Now()
	svc.logger.Info("stopping a gateway", zap.Any("id", id))
	service := svc.store.Get(id)
	if service != nil {
		err := service.Stop()
		state := types.State{
			Status:  types.StatusDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			svc.logger.Error("failed to stop a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
			state.Message = fmt.Sprintf("Failed to stop: %s", err.Error())
			busUtils.SetGatewayState(svc.logger, svc.bus, id, state)
		} else {
			svc.logger.Info("stopped a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()))
			busUtils.SetGatewayState(svc.logger, svc.bus, id, state)
			svc.store.Remove(id)
		}
	}
	return nil
}

// unloadAll stops all the gateways
func (svc *GatewayService) unloadAll() {
	ids := svc.store.ListIDs()
	for _, id := range ids {
		err := svc.stopGW(id)
		if err != nil {
			svc.logger.Error("error on stopping a gateway", zap.String("id", id))
		}
	}
}

// returns sleeping queue messages from the given gateway ID
func (svc *GatewayService) getGatewaySleepingQueue(gatewayID string) *map[string][]msgTY.Message {
	service := svc.store.Get(gatewayID)
	if service != nil {
		messages := service.GetGatewaySleepingQueue()
		return &messages
	}
	return nil
}

// returns sleeping queue messages from the given gateway ID and node ID
func (svc *GatewayService) getNodeSleepingQueue(gatewayID, nodeID string) *[]msgTY.Message {
	service := svc.store.Get(gatewayID)
	if service != nil {
		messages := service.GetNodeSleepingQueue(nodeID)
		return &messages
	}
	return nil
}

// clears sleeping queue of a gateway
func (svc *GatewayService) clearGatewaySleepingQueue(gatewayID string) {
	service := svc.store.Get(gatewayID)
	if service != nil {
		service.ClearGatewaySleepingQueue()
	}
}

// clears sleeping queue of a node
func (svc *GatewayService) clearNodeSleepingQueue(gatewayID, nodeID string) {
	service := svc.store.Get(gatewayID)
	if service != nil {
		service.ClearNodeSleepingQueue(nodeID)
	}
}
