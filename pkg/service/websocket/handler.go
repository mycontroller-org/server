package mcwebsocket

import (
	"net/http"

	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// RegisterWebsocketRoutes registers it into the handlers
func RegisterWebsocketRoutes(router *mux.Router) {
	err := initEventListener()
	if err != nil {
		zap.L().Error("error on calling websocket init", zap.Error(err))
	}
	router.HandleFunc("/api/ws", wsFunc)
}

var (
	clients  = make(map[*ws.Conn]bool) // connected clients
	upgrader = ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// this is simple example websocket
// yet to implement actual version
func wsFunc(w http.ResponseWriter, r *http.Request) {
	wsCon, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("websocket upgrade error", zap.Error(err))
		return
	}
	defer wsCon.Close()

	// Register our new client
	clients[wsCon] = true

	for {
		mt, message, err := wsCon.ReadMessage()
		if err != nil {
			zap.L().Info("websocket read error", zap.String("error", err.Error()), zap.Any("remoteAddress", wsCon.RemoteAddr()))
			delete(clients, wsCon)
			break
		}
		zap.L().Debug("websocket received message", zap.String("message", string(message)), zap.Any("from port", r.RemoteAddr))
		msg := string(message)
		msg += ", you are calling from: " + r.RemoteAddr
		err = wsCon.WriteMessage(mt, []byte(msg))
		if err != nil {
			zap.L().Info("websocket write error", zap.String("error", err.Error()), zap.Any("remoteAddress", wsCon.RemoteAddr()))
			break
		}
	}
}
