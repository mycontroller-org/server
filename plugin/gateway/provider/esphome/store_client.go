package esphome

import (
	"sync"

	"go.uber.org/zap"
)

// ClientStore holds esphome node instances
type ClientStore struct {
	GatewayID string
	nodes     map[string]*ESPHomeNode
	mutex     sync.RWMutex
}

// AddNode adds a esphome node to the store
func (s *ClientStore) AddNode(nodeID string, client *ESPHomeNode) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.nodes[nodeID] = client
}

// Get returns a esphome node from the store
func (s *ClientStore) Get(nodeID string) *ESPHomeNode {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, found := s.nodes[nodeID]
	if found {
		return client
	}
	return nil
}

// Close disconnects all the available esphome nodes and removes from the store
func (s *ClientStore) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for nodeID, client := range s.nodes {
		err := client.Disconnect()
		if err != nil {
			zap.L().Error("error on disconnecting a client", zap.String("gatewayId", s.GatewayID), zap.String("nodeId", nodeID), zap.Error(err))
		}
	}
}
