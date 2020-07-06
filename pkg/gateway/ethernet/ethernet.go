package ethernet

import (
	"net"
	"net/url"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	m2s "github.com/mitchellh/mapstructure"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
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

	d := &Endpoint{
		Config:    cfg,
		Client:    c,
		Queue:     queue,
		GatewayID: gID,
	}
	return d, nil
}

// Write sends a payload
func (d *Endpoint) Write(rm *msg.RawMessage) error {
	_, err := d.Client.Write(rm.Data)
	return err
}

// Close the connection
func (d *Endpoint) Close() error {
	return d.Close()
}
