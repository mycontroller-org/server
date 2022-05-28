package mcwebsocket

import (
	"net/http"

	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	upgrader = ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// RegisterWebsocketRoutes registers it into the handlers
func RegisterWebsocketRoutes(router *mux.Router) {
	err := initEventListener()
	if err != nil {
		zap.L().Error("error on calling websocket init", zap.Error(err))
	}
	router.HandleFunc("/api/ws", wsFunc)
}

// this is simple example websocket
// yet to implement actual version
func wsFunc(w http.ResponseWriter, r *http.Request) {
	wsCon, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Info("websocket upgrade error", zap.String("error", err.Error()))
		return
	}

	// register the new client
	clientStore.register(wsCon)

	// NOTE: for now not serving any request, only sending the events to the listeners(ex: remote browsers)
	// this loop is used to close the connection immediately on remote side close
	for {
		_, _, err := wsCon.ReadMessage()
		if err != nil {
			zap.L().Debug("websocket read error", zap.String("error", err.Error()), zap.Any("remoteAddress", wsCon.RemoteAddr()))
			clientStore.unregister(wsCon)
			break
		}
	}
}
