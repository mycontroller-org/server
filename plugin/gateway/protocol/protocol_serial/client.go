package serial

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	msglogger "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/message_logger"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	serialDriver "github.com/tarm/serial"
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
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwTY.Config
	Config         Config
	serCfg         *serialDriver.Config
	Port           *serialDriver.Port
	receiveMsgFunc func(rm *msgTY.RawMessage) error
	safeClose      *concurrency.Channel
	messageLogger  msglogger.MessageLogger
	txPreDelay     time.Duration
	reconnectDelay time.Duration
	mutex          sync.RWMutex
}

// New serial client
func New(gwCfg *gwTY.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgTY.RawMessage) error) (*Endpoint, error) {
	var cfg Config
	err := utils.MapToStruct(utils.TagNameNone, protocol, &cfg)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("updated config data", zap.Any("config", cfg))

	serCfg := &serialDriver.Config{Name: cfg.Portname, Baud: cfg.BaudRate}

	zap.L().Info("opening a serial port", zap.String("gateway", gwCfg.ID), zap.String("port", cfg.Portname))
	port, err := serialDriver.OpenPort(serCfg)
	if err != nil {
		// zap.L().Error("error on opening port", zap.String("gateway", gwCfg.ID), zap.String("port", serCfg.Name), zap.String("error", err.Error()))
		return nil, err
	}

	endpoint := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		serCfg:         serCfg,
		receiveMsgFunc: rxMsgFunc,
		Port:           port,
		safeClose:      concurrency.NewChannel(0),
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
		reconnectDelay: utils.ToDuration(gwCfg.ReconnectDelay, reconnectDelayDefault),
		mutex:          sync.RWMutex{},
	}

	// init and start message logger
	endpoint.messageLogger = msglogger.Init(gwCfg.ID, gwCfg.MessageLogger, messageFormatter)
	endpoint.messageLogger.Start()

	// start serail read listener
	go endpoint.dataListener()
	return endpoint, nil
}

func messageFormatter(rawMsg *msgTY.RawMessage) string {
	direction := "sent"
	if rawMsg.IsReceived {
		direction = "recd"
	}
	data := strings.TrimSuffix(convertor.ToString(rawMsg.Data), "\n")
	return fmt.Sprintf("%v\t%v\t%s\n", rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"), direction, data)
}

func (ep *Endpoint) Write(rawMsg *msgTY.RawMessage) error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	time.Sleep(ep.txPreDelay) // transmit pre delay
	ep.messageLogger.AsyncWrite(rawMsg)

	dataBytes, ok := rawMsg.Data.([]byte)
	if !ok {
		zap.L().Error("error on converting to bytes", zap.Any("rawMessage", rawMsg))
		return fmt.Errorf("error on converting to bytes. received: %T", rawMsg.Data)
	}
	_, err := ep.Port.Write(dataBytes)
	return err
}

// Close the driver
func (ep *Endpoint) Close() error {
	go func() { ep.safeClose.SafeSend(true) }() // terminate the data listener

	if ep.Port != nil {
		if err := ep.Port.Flush(); err != nil {
			zap.L().Error("error on flushing the serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
		}
		err := ep.Port.Close()
		if err != nil {
			zap.L().Error("error on closing the serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
		}
		return err
	}
	return nil
}

// DataListener func
func (ep *Endpoint) dataListener() {
	readBuf := make([]byte, 128)
	data := make([]byte, 0)
	for {
		select {
		case <-ep.safeClose.CH:
			zap.L().Info("received close signal.", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
			return
		default:
			rxLength, err := ep.Port.Read(readBuf)
			if err != nil {
				zap.L().Error("error on reading data from the serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
				state := types.State{
					Status:  types.StatusDown,
					Message: err.Error(),
					Since:   time.Now(),
				}
				busUtils.SetGatewayState(ep.GwCfg.ID, state)

				// channel close panic issue with internal reconnect
				// let it reconnected from gateway service
				// go ep.reconnect()

				return
			}
			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == ep.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					rawMsg := msgTY.NewRawMessage(true, dataCloned)
					ep.messageLogger.AsyncWrite(rawMsg)
					err := ep.receiveMsgFunc(rawMsg)
					if err != nil {
						zap.L().Error("error on sending a raw message to queue", zap.String("gateway", ep.GwCfg.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
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

// func (ep *Endpoint) reconnect() {
// 	ticker := time.NewTicker(ep.reconnectDelay)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ep.safeClose.CH:
// 			zap.L().Debug("Received close signal", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
// 			return
//
// 		case <-ticker.C: // reconnect
// 			// close the port
// 			if ep.Port != nil {
// 				err := ep.Port.Close()
// 				if err != nil {
// 					zap.L().Error("Error on closing a serial port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
// 				}
// 				ep.Port = nil
// 			}
// 			// open the port
// 			port, err := ser.OpenPort(ep.serCfg)
// 			if err == nil {
// 				zap.L().Debug("serial port reconnected successfully", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name))
// 				ep.Port = port
// 				go ep.dataListener() // if connection success, start read listener
// 				state := types.State{
// 					Status:  types.StatusUp,
// 					Message: "Reconnected successfully",
// 					Since:   time.Now(),
// 				}
// 				busUtils.SetGatewayState(ep.GwCfg.ID, state)
// 				return
// 			}
// 			zap.L().Error("Error on opening a port", zap.String("gateway", ep.GwCfg.ID), zap.String("port", ep.serCfg.Name), zap.Error(err))
// 			state := types.State{
// 				Status:  types.StatusDown,
// 				Message: err.Error(),
// 				Since:   time.Now(),
// 			}
// 			busUtils.SetGatewayState(ep.GwCfg.ID, state)
// 		}
// 	}
//
// }
