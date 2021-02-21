package serial

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	rsUtils "github.com/mycontroller-org/backend/v2/pkg/utils/resource_service"
	msglogger "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol/message_logger"
	ser "github.com/tarm/serial"
	"go.uber.org/zap"
)

// Constants in serial protocol
const (
	KeyMessageSplitter      = "MessageSplitter"
	MaxDataLength           = 1000
	transmitPreDelayDefault = time.Millisecond * 1 // 1ms
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Config details
type Config struct {
	Portname         string
	BaudRate         int
	MessageSplitter  byte
	TransmitPreDelay string
	ReconnectDelay   string
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwml.Config
	Config         Config
	serCfg         *ser.Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Port           *ser.Port
	closeCh        chan bool
	messageLogger  msglogger.MessageLogger
	txPreDelay     time.Duration
	reconnectDelay time.Duration
}

// New serial client
func New(gwCfg *gwml.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	zap.L().Info("Init protocol", zap.String("gateway", gwCfg.ID))
	var cfg Config
	err := utils.MapToStruct(utils.TagNameNone, protocol, &cfg)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("config:", zap.Any("converted", cfg))

	serCfg := &ser.Config{Name: cfg.Portname, Baud: cfg.BaudRate}

	port, err := ser.OpenPort(serCfg)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		serCfg:         serCfg,
		receiveMsgFunc: rxMsgFunc,
		Port:           port,
		closeCh:        make(chan bool),
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
		reconnectDelay: utils.ToDuration(cfg.ReconnectDelay, reconnectDelayDefault),
	}

	// init and start message logger
	endpoint.messageLogger = msglogger.Init(gwCfg.ID, gwCfg.MessageLogger, messageFormatter)
	endpoint.messageLogger.Start()

	// start serail read listener
	go endpoint.dataListener()
	return endpoint, nil
}

func messageFormatter(rawMsg *msgml.RawMessage) string {
	direction := "Sent"
	if rawMsg.IsReceived {
		direction = "Recd"
	}
	return fmt.Sprintf("%v\t%v\t%s\n", rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"), direction, string(rawMsg.Data))
}

func (ep *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	time.Sleep(ep.txPreDelay) // transmit pre delay
	ep.messageLogger.AsyncWrite(rawMsg)

	_, err := ep.Port.Write(rawMsg.Data)
	return err
}

// Close the driver
func (ep *Endpoint) Close() error {
	go func() { ep.closeCh <- true }() // terminate the data listener

	if err := ep.Port.Flush(); err != nil {
		zap.L().Error("Error on flushing a serial port", zap.String("gateway", ep.GwCfg.Name), zap.String("port", ep.serCfg.Name), zap.Error(err))
	}
	err := ep.Port.Close()
	if err != nil {
		zap.L().Error("Error on closing a serial port", zap.String("gateway", ep.GwCfg.Name), zap.String("port", ep.serCfg.Name), zap.Error(err))
	}
	return err
}

// DataListener func
func (ep *Endpoint) dataListener() {
	readBuf := make([]byte, 128)
	data := make([]byte, 0)
	for {
		select {
		case <-ep.closeCh:
			zap.L().Debug("Received close signal.", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
			return
		default:
			rxLength, err := ep.Port.Read(readBuf)
			if err != nil {
				zap.L().Error("Error on reading data from a serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
				state := model.State{
					Status:  model.StateDown,
					Message: err.Error(),
					Since:   time.Now(),
				}
				rsUtils.SetGatewayState(ep.GwCfg.ID, state)
				go ep.reconnect()
				return
			}
			//zap.L().Debug("data", zap.Any("data", string(data)))
			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == ep.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					rawMsg := msgml.NewRawMessage(true, dataCloned)
					//	zap.L().Debug("new message received", zap.Any("rawMessage", rawMsg))
					ep.messageLogger.AsyncWrite(rawMsg)
					err := ep.receiveMsgFunc(rawMsg)
					if err != nil {
						zap.L().Error("Error on sending a raw message to queue", zap.String("gateway", ep.GwCfg.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
					}
				} else {
					data = append(data, b)
				}
				if len(data) >= MaxDataLength {
					data = nil
				}
			}
		}
	}
}

func (ep *Endpoint) reconnect() {
	ticker := time.NewTicker(ep.reconnectDelay)
	defer ticker.Stop()
	for {
		select {
		case <-ep.closeCh:
			zap.L().Debug("Received close signal", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
			return

		case <-ticker.C: // reconnect
			// close the port
			if ep.Port != nil {
				err := ep.Port.Close()
				if err != nil {
					zap.L().Error("Error on closing a serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
				}
				ep.Port = nil
			}
			// open the port
			port, err := ser.OpenPort(ep.serCfg)
			if err == nil {
				zap.L().Debug("serial port reconnected successfully", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
				ep.Port = port
				go ep.dataListener() // if connection success, start read listener
				state := model.State{
					Status:  model.StateUp,
					Message: "Reconnected successfully",
					Since:   time.Now(),
				}
				rsUtils.SetGatewayState(ep.GwCfg.ID, state)
				return
			}
			zap.L().Error("Error on opening a port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
			state := model.State{
				Status:  model.StateDown,
				Message: err.Error(),
				Since:   time.Now(),
			}
			rsUtils.SetGatewayState(ep.GwCfg.ID, state)
		}
	}

}
