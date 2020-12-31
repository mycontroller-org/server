package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func registerWebsocketRoutes(router *mux.Router) {
	router.HandleFunc("/api/ws", ws)
}

var upgrader = websocket.Upgrader{}

// this is simple example websocket
// yet to implement actual version
func ws(w http.ResponseWriter, r *http.Request) {
	wsCon, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("websocket upgrade error", zap.Error(err))
		return
	}
	defer wsCon.Close()
	for {
		mt, message, err := wsCon.ReadMessage()
		if err != nil {
			zap.L().Error("websocket read error", zap.Error(err))
			break
		}
		zap.L().Debug("websocket received message", zap.String("message", string(message)), zap.Any("from port", r.RemoteAddr))
		msg := string(message)
		msg += ", you are calling from: " + r.RemoteAddr
		err = wsCon.WriteMessage(mt, []byte(msg))
		if err != nil {
			zap.L().Error("websocket write error", zap.Error(err))
			break
		}
	}
}
