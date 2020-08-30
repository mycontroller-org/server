package mysensors

import (
	"fmt"

	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway_protocol"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_protocol/mqtt"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_protocol/serial"
)

// Provider implementation
type Provider struct {
	GWConfig *gwml.Config
	Gateway  gwpl.Gateway
}

// Post func
func (p *Provider) Post(rawMsg *msgml.RawMessage) error {
	return p.Gateway.Write(rawMsg)
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error {
	var err error
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeMQTT:
		ms, _err := mqtt.New(p.GWConfig, rxMessageFunc)
		err = _err
		p.Gateway = ms
	case gwpl.TypeSerial:
		// update serial message splitter
		p.GWConfig.Provider.Config[serial.KeyMessageSplitter] = serialMessageSplitter
		ms, _err := serial.New(p.GWConfig, rxMessageFunc)
		err = _err
		p.Gateway = ms
	}
	if err != nil {
		return err
	}

	// load firmware purge job
	fwPurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GWConfig.ID)
	return svc.SCH.AddFunc(fwPurgeJobName, firmwarePurgeJobCron, fwStore.purge)
}

// Close func
func (p *Provider) Close() error {
	// remove firmware purge job
	fwPurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GWConfig.ID)
	svc.SCH.RemoveFunc(fwPurgeJobName)
	// close gateway
	return p.Gateway.Close()
}
