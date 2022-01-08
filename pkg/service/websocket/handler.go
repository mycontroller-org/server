package mcwebsocket

import (
	"net/http"
	"sync"

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
	clients = make(map[*ws.Conn]bool) // connected clients
	mutex   = sync.Mutex{}

	upgrader = ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// register a websocket client
func registerClient(conn *ws.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	clients[conn] = true
}

// unregister a websocket client
func unregisterClient(conn *ws.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clients, conn)
}

// returns available websocket clients
func getClients() []*ws.Conn {
	mutex.Lock()
	defer mutex.Unlock()

	wsClients := make([]*ws.Conn, 0)
	for client := range clients {
		wsClients = append(wsClients, client)
	}
	return wsClients
}

// this is simple example websocket
// yet to implement actual version
func wsFunc(w http.ResponseWriter, r *http.Request) {
	wsCon, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Info("websocket upgrade error", zap.String("error", err.Error()))
		return
	}
	defer wsCon.Close()

	// Register our new client
	registerClient(wsCon)

	for {
		mt, message, err := wsCon.ReadMessage()
		if err != nil {
			zap.L().Debug("websocket read error", zap.String("error", err.Error()), zap.Any("remoteAddress", wsCon.RemoteAddr()))
			unregisterClient(wsCon)
			break
		}
		zap.L().Debug("websocket received message", zap.String("message", string(message)), zap.Any("from port", r.RemoteAddr))
		msg := string(message)
		msg += ", you are calling from: " + r.RemoteAddr
		err = wsCon.WriteMessage(mt, []byte(msg))
		if err != nil {
			zap.L().Debug("websocket write error", zap.String("error", err.Error()), zap.Any("remoteAddress", wsCon.RemoteAddr()))
			break
		}
	}
}
