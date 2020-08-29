package serial

import (
	m2s "github.com/mitchellh/mapstructure"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	s "github.com/tarm/serial"
)

// Constants in serial gateway
const (
	KeyMessageSplitter = "MessageSplitter"
)

// Config details
type Config struct {
	Portname       string
	BaudRate       int
	MessageSpliter byte
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwml.Config
	Config         Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Port           *s.Port
}

// New serial client
func New(gwCfg *gwml.Config, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	var cfg Config
	err := m2s.Decode(gwCfg.Provider.Config, &cfg)
	if err != nil {
		return nil, err
	}
	c := &s.Config{Name: cfg.Portname, Baud: cfg.BaudRate}
	port, err := s.OpenPort(c)
	if err != nil {
		return nil, err
	}
	d := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		receiveMsgFunc: rxMsgFunc,
		Port:           port,
	}
	return d, nil
}

func (d *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	_, err := d.Port.Write(rawMsg.Data)
	return err
}

// Close the driver
func (d *Endpoint) Close() error {
	return d.Port.Close()
}
