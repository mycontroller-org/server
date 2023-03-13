package mcwebsocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	defaultWriteTimeout = time.Second * 3
	defaultQueueSize    = int(1000)
	defaultWorkers      = int(1)
	websocketPath       = "/api/ws"
)

type WebsocketService struct {
	ctx         context.Context
	logger      *zap.Logger
	store       *Store
	bus         busTY.Plugin
	api         *entityAPI.API
	eventsQueue *queueUtils.QueueSpec
	router      *mux.Router
}

func New(ctx context.Context, router *mux.Router) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &WebsocketService{
		ctx:    ctx,
		logger: logger.Named("websocket_service"),
		bus:    bus,
		api:    api,
		router: router,
	}

	svc.store = &Store{clients: make(map[*ws.Conn]bool), mutex: sync.RWMutex{}, logger: svc.logger}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "websocket_event_listener", defaultQueueSize, svc.processEvent, defaultWorkers),
		Topic:          topic.TopicEventsAll,
		SubscriptionId: -1,
	}

	// register handler path
	// needs to be registered before passing into http_router
	svc.router.HandleFunc(websocketPath, svc.wsFunc)

	return svc, nil
}

func (svc *WebsocketService) Name() string {
	return "websocket_service"
}
