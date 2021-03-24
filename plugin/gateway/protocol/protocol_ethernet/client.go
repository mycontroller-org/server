package ethernet

import (
	"net"
	"net/url"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	m2s "github.com/mitchellh/mapstructure"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// Config details
type Config struct {
	URI string
}

// Endpoint data
type Endpoint struct {
	Config    Config
	Client    net.Conn
	Queue     *q.BoundedQueue
	GatewayID string
}

// New ethernet driver
func New(config map[string]interface{}, queue *q.BoundedQueue, gID string) (*Endpoint, error) {
	var cfg Config
	err := m2s.Decode(config, &cfg)
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(cfg.URI)
	if err != nil {
		return nil, err
	}

	c, err := net.Dial(uri.Scheme, uri.Host)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		Config:    cfg,
		Client:    c,
		Queue:     queue,
		GatewayID: gID,
	}
	return endpoint, nil
}

// Write sends a payload
func (ep *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	_, err := ep.Client.Write(rawMsg.Data)
	return err
}

// Close the connection
func (ep *Endpoint) Close() error {
	return ep.Client.Close()
}
