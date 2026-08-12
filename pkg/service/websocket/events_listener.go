package mcwebsocket

import (
	"time"

	ws "github.com/gorilla/websocket"
	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	"github.com/mycontroller-org/server/v2/pkg/json"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	wsTY "github.com/mycontroller-org/server/v2/pkg/types/websocket"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// starts events listener
func (svc *WebsocketService) startEventListener() error {
	// on message receive add it in to our local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEventReceive)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = sID
	return nil
}

func (svc *WebsocketService) CloseEventListener() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		return err
	}
	svc.eventsQueue.Close()
	return nil
}

func (svc *WebsocketService) onEventReceive(data *busTY.BusData) {
	status := svc.eventsQueue.Produce(data)
	if !status {
		svc.logger.Error("failed to post a event on the processor queue")
	}
}

func (svc *WebsocketService) processEvent(item interface{}) error {
	// if there is no clients, just ignore the event
	if svc.store.getSize() == 0 {
		return nil
	}

	data := item.(*busTY.BusData)

	event := &eventTY.Event{}
	err := data.LoadData(event)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Any("topic", data.Topic), zap.Error(err))
		return nil
	}

	svc.logger.Debug("event received", zap.Any("event", event))

	response := wsTY.Response{
		Type: wsTY.ResponseTypeEvent,
		Data: event,
	}

	// convert to json bytes
	dataBytes, err := json.Marshal(response)
	if err != nil {
		svc.logger.Error("error on converting to json", zap.Error(err))
		return nil
	}

	wsClients := svc.store.getClients()
	for client, subject := range wsClients {
		if !svc.eventAllowed(subject, event) {
			continue
		}

		// write with write timeout
		err := client.SetWriteDeadline(time.Now().Add(defaultWriteTimeout))
		if err != nil {
			svc.logger.Debug("error on setting write deadline", zap.Any("remoteAddress", client.RemoteAddr().String()), zap.Error(err))
			svc.store.unregister(client)
			continue
		}
		err = client.WriteMessage(ws.TextMessage, dataBytes)
		if err != nil {
			svc.logger.Debug("error on write data to a client", zap.Any("remoteAddress", client.RemoteAddr().String()), zap.Error(err))
			svc.store.unregister(client)
		}
	}
	return nil
}

// eventAllowed reports whether this principal may see the live event.
// Quick ids are checked as the named resource; otherwise the check is
// kind:entityId. Events with neither a parseable quick id nor EntityID are denied.
func (svc *WebsocketService) eventAllowed(subject policyAPI.Subject, event *eventTY.Event) bool {
	if subject.UserID == "" {
		return false
	}
	ac := svc.api.Policy()
	if ac == nil {
		return false
	}
	resource := ""
	if event.EntityQuickID != "" {
		if res, err := policyAPI.ResourceFromQuickID(event.EntityQuickID); err == nil {
			resource = res
		}
	}
	if resource == "" {
		if event.EntityID == "" {
			return false
		}
		kind := policyTY.NormalizeKind(event.EntityType)
		if kind == "" {
			return false
		}
		resource = policyAPI.FormatResource(kind, event.EntityID)
	}
	return ac.Allowed(subject, policyTY.ActionGet, resource) == nil
}
