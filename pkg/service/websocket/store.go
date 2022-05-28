package mcwebsocket

import (
	"sync"

	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type store struct {
	clients map[*ws.Conn]bool
	mutex   sync.RWMutex
}

var clientStore = store{
	clients: make(map[*ws.Conn]bool),
	mutex:   sync.RWMutex{},
}

// register a websocket client connection
func (s *store) register(conn *ws.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[conn] = true
	zap.L().Debug("new websocket connection added", zap.String("remoteAddress", conn.RemoteAddr().String()))
}

// unregister a websocket client connection
func (s *store) unregister(conn *ws.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := conn.Close()
	if err != nil {
		zap.L().Debug("error on closing the connection", zap.String("remoteAddress", conn.RemoteAddr().String()), zap.Error(err))
	} else {
		zap.L().Debug("websocket connection closed", zap.String("remoteAddress", conn.RemoteAddr().String()))
	}
	delete(s.clients, conn)
}

// returns available websocket client connection
func (s *store) getClients() []*ws.Conn {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	wsClients := make([]*ws.Conn, 0)
	for client := range s.clients {
		wsClients = append(wsClients, client)
	}
	return wsClients
}

// returns the size of the client map
func (s *store) getSize() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.clients)
}
