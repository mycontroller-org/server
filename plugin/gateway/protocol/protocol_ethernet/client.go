package ethernet

import (
	"fmt"
	"net"
	"net/url"
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
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	msgLogger "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/message_logger"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// Constants in ethernet protocol
const (
	KeyMessageSplitter      = "MessageSplitter"
	MaxDataLength           = 1000
	transmitPreDelayDefault = time.Millisecond * 1 // 1ms
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
	loggerName              = "protocol_ethernet"
)

// Config details
type Config struct {
	Server           string
	MessageSplitter  byte
	TransmitPreDelay string
	Insecure         bool
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwTY.Config
	Config         Config
	connUrl        *url.URL
	conn           net.Conn
	receiveMsgFunc func(rm *msgTY.RawMessage) error
	safeClose      *concurrency.Channel
	messageLogger  msgLogger.MessageLogger
	txPreDelay     time.Duration
	reconnectDelay time.Duration
	mutex          sync.RWMutex
	logger         *zap.Logger
	bus            busTY.Plugin
}

// New ethernet driver
func New(logger *zap.Logger, gwCfg *gwTY.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgTY.RawMessage) error, bus busTY.Plugin, logRootDir string) (*Endpoint, error) {
	var cfg Config
	err := utils.MapToStruct(utils.TagNameNone, protocol, &cfg)
	if err != nil {
		return nil, err
	}

	namedLogger := logger.Named(loggerName)

	namedLogger.Debug("updated config data", zap.Any("config", cfg))

	serverURL, err := url.Parse(cfg.Server)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(serverURL.Scheme, serverURL.Host)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		connUrl:        serverURL,
		conn:           conn,
		receiveMsgFunc: rxMsgFunc,
		safeClose:      concurrency.NewChannel(0),
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
		reconnectDelay: utils.ToDuration(gwCfg.ReconnectDelay, reconnectDelayDefault),
		mutex:          sync.RWMutex{},
		logger:         namedLogger,
		bus:            bus,
	}

	// init and start message logger
	endpoint.messageLogger = msgLogger.New(logger, gwCfg.ID, gwCfg.MessageLogger, messageFormatter, logRootDir)
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
		ep.logger.Error("error on converting to bytes", zap.Any("rawMessage", rawMsg))
		return fmt.Errorf("error on converting to bytes. received: %T", rawMsg.Data)
	}
	_, err := ep.conn.Write(dataBytes)
	return err
}

// Close the driver
func (ep *Endpoint) Close() error {
	go func() { ep.safeClose.SafeSend(true) }() // terminate the data listener

	if ep.conn != nil {
		err := ep.conn.Close()
		if err != nil {
			ep.logger.Error("error on closing the connection", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server), zap.Error(err))
		}
		ep.conn = nil
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
			ep.logger.Info("received close signal.", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server))
			return
		default:
			rxLength, err := ep.conn.Read(readBuf)
			if err != nil {
				ep.logger.Error("error on reading the data from the ethernet connection", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server), zap.Error(err))
				state := types.State{
					Status:  types.StatusDown,
					Message: err.Error(),
					Since:   time.Now(),
				}
				busUtils.SetGatewayState(ep.logger, ep.bus, ep.GwCfg.ID, state)
				return
			}

			//ep.logger.Debug("data", zap.Any("data", string(data)))
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
						ep.logger.Error("error on sending a raw message to queue", zap.String("gateway", ep.GwCfg.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
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
