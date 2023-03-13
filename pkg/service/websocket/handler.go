package mcwebsocket

import (
	"net/http"

	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	upgrader = ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// registers it into the handlers and starts the event listener
func (svc *WebsocketService) Start() error {
	err := svc.startEventListener()
	if err != nil {
		svc.logger.Error("error on websocket event listener", zap.Error(err))
		return err
	}
	svc.router.HandleFunc("/api/ws", svc.wsFunc)
	return nil
}

// this is simple example websocket
// yet to implement actual version
func (svc *WebsocketService) wsFunc(w http.ResponseWriter, r *http.Request) {
	wsCon, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		svc.logger.Info("websocket upgrade error", zap.Error(err))
		return
	}

	// register the new client
	svc.store.register(wsCon)

	// NOTE: for now not serving any request, only sending the events to the listeners(ex: remote browsers)
	// this loop is used to close the connection immediately on remote side close
	for {
		_, _, err := wsCon.ReadMessage()
		if err != nil {
			svc.logger.Debug("websocket read error", zap.Any("remoteAddress", wsCon.RemoteAddr()), zap.Error(err))
			svc.store.unregister(wsCon)
			break
		}
	}
}

func (svc *WebsocketService) Close() error {
	return nil
}
