package mcwebsocket

import (
	"sync"

	ws "github.com/gorilla/websocket"
	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	"go.uber.org/zap"
)

type Store struct {
	clients map[*ws.Conn]policyAPI.Subject
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// register a websocket client connection
func (s *Store) register(conn *ws.Conn, subject policyAPI.Subject) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[conn] = subject
	s.logger.Debug("new websocket connection added", zap.String("remoteAddress", conn.RemoteAddr().String()), zap.String("userId", subject.UserID))
}

// unregister a websocket client connection
func (s *Store) unregister(conn *ws.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := conn.Close()
	if err != nil {
		s.logger.Debug("error on closing the connection", zap.String("remoteAddress", conn.RemoteAddr().String()), zap.Error(err))
	} else {
		s.logger.Debug("websocket connection closed", zap.String("remoteAddress", conn.RemoteAddr().String()))
	}
	delete(s.clients, conn)
}

// returns available websocket client connections with their access subject
func (s *Store) getClients() map[*ws.Conn]policyAPI.Subject {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	out := make(map[*ws.Conn]policyAPI.Subject, len(s.clients))
	for client, subject := range s.clients {
		out[client] = subject
	}
	return out
}

// returns the size of the client map
func (s *Store) getSize() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.clients)
}
