package ethernet

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	msgLogger "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol/message_logger"
	"go.uber.org/zap"
)

// Constants in ethernet protocol
const (
	KeyMessageSplitter      = "MessageSplitter"
	MaxDataLength           = 1000
	transmitPreDelayDefault = time.Millisecond * 1 // 1ms
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Config details
type Config struct {
	Server             string
	MessageSplitter    byte
	TransmitPreDelay   string
	InsecureSkipVerify bool
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwML.Config
	Config         Config
	connUrl        *url.URL
	conn           net.Conn
	receiveMsgFunc func(rm *msgML.RawMessage) error
	safeClose      *concurrency.Channel
	messageLogger  msgLogger.MessageLogger
	txPreDelay     time.Duration
	reconnectDelay time.Duration
}

// New ethernet driver
func New(gwCfg *gwML.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgML.RawMessage) error) (*Endpoint, error) {
	var cfg Config
	err := utils.MapToStruct(utils.TagNameNone, protocol, &cfg)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("config:", zap.Any("converted", cfg))

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
	}

	// init and start message logger
	endpoint.messageLogger = msgLogger.Init(gwCfg.ID, gwCfg.MessageLogger, messageFormatter)
	endpoint.messageLogger.Start()

	// start serail read listener
	go endpoint.dataListener()
	return endpoint, nil
}

func messageFormatter(rawMsg *msgML.RawMessage) string {
	direction := "Sent"
	if rawMsg.IsReceived {
		direction = "Recd"
	}
	return fmt.Sprintf("%v\t%v\t%s\n", rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"), direction, string(rawMsg.Data))
}

func (ep *Endpoint) Write(rawMsg *msgML.RawMessage) error {
	time.Sleep(ep.txPreDelay) // transmit pre delay
	ep.messageLogger.AsyncWrite(rawMsg)

	_, err := ep.conn.Write(rawMsg.Data)
	return err
}

// Close the driver
func (ep *Endpoint) Close() error {
	go func() { ep.safeClose.SafeSend(true) }() // terminate the data listener

	if ep.conn != nil {
		err := ep.conn.Close()
		if err != nil {
			zap.L().Error("Error on closing a a connection", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server), zap.Error(err))
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
			zap.L().Info("Received close signal.", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server))
			return
		default:
			rxLength, err := ep.conn.Read(readBuf)
			if err != nil {
				zap.L().Error("Error on reading data from a ethernet connection", zap.String("gateway", ep.GwCfg.ID), zap.String("server", ep.Config.Server), zap.Error(err))
				state := model.State{
					Status:  model.StatusDown,
					Message: err.Error(),
					Since:   time.Now(),
				}
				busUtils.SetGatewayState(ep.GwCfg.ID, state)

				// channel close panic issue with internal reconnect
				// let it reconnected from gateway service
				// go ep.reconnect() // refer serial port protocol
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
					rawMsg := msgML.NewRawMessage(true, dataCloned)
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
