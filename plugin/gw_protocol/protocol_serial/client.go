package serial

import (
	"fmt"
	"time"

	m2s "github.com/mitchellh/mapstructure"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	msglogger "github.com/mycontroller-org/backend/v2/plugin/gw_protocol/message_logger"
	s "github.com/tarm/serial"
	"go.uber.org/zap"
)

// Constants in serial gateway
const (
	KeyMessageSplitter = "MessageSplitter"
	MaxDataLength      = 1000
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
	GwCfg          *gwml.Config
	Config         Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Port           *s.Port
	closeCh        chan bool
	messageLogger  msglogger.MessageLogger
	txPreDelay     *time.Duration
}

// New serial client
func New(gwCfg *gwml.Config, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	var cfg Config
	err := m2s.Decode(gwCfg.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	c := &s.Config{Name: cfg.Portname, Baud: cfg.BaudRate}
	port, err := s.OpenPort(c)
	if err != nil {
		return nil, err
	}
	var txPreDelay *time.Duration
	// get transmit pre delay
	if cfg.TransmitPreDelay != "" {
		_txPreDelay, err := time.ParseDuration(cfg.TransmitPreDelay)
		if err != nil {
			zap.L().Error("Failed to parse transmit delay", zap.String("transmitPreDelay", cfg.TransmitPreDelay))
		}
		txPreDelay = &_txPreDelay
	}

	d := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		receiveMsgFunc: rxMsgFunc,
		Port:           port,
		closeCh:        make(chan bool),
		txPreDelay:     txPreDelay,
	}

	// init and start message logger
	d.messageLogger = msglogger.Init(gwCfg.ID, gwCfg.MessageLogger, messageFormatter)
	d.messageLogger.Start()

	// start serail read listener
	go d.dataListener()
	return d, nil
}

func messageFormatter(rawMsg *msgml.RawMessage) string {
	direction := "Sent"
	if rawMsg.IsReceived {
		direction = "Recd"
	}
	return fmt.Sprintf("%v\t%v\t%s\n", rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"), direction, string(rawMsg.Data))
}

func (d *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	// add transmit pre delay
	if d.txPreDelay != nil {
		time.Sleep(*d.txPreDelay)
	}
	d.messageLogger.AsyncWrite(rawMsg)

	_, err := d.Port.Write(rawMsg.Data)
	return err
}

// Close the driver
func (d *Endpoint) Close() error {
	d.closeCh <- true // terminate the data listener

	if err := d.Port.Flush(); err != nil {
		zap.L().Error("Failed to flush the serial port", zap.String("gateway", d.GwCfg.Name), zap.String("port", d.Config.Portname), zap.Error(err))
	}
	err := d.Port.Close()
	if err != nil {
		zap.L().Error("Failed to close the serial port connection", zap.String("gateway", d.GwCfg.Name), zap.String("port", d.Config.Portname), zap.Error(err))
	}
	return err
}

// DataListener func
func (d *Endpoint) dataListener() {
	readBuf := make([]byte, 128)
	data := make([]byte, 0)
	for {
		select {
		case <-d.closeCh:
			zap.L().Debug("Received read listener close signal.", zap.String("gateway", d.GwCfg.Name), zap.String("port", d.Config.Portname))
			return
		default:
			rxLength, err := d.Port.Read(readBuf)
			if err != nil {
				zap.L().Error("Failed to read data from the serial port", zap.String("gateway", d.GwCfg.Name), zap.String("port", d.Config.Portname), zap.Error(err))
				return
			}
			//zap.L().Debug("data", zap.Any("data", string(data)))
			for index := 0; index < rxLength; index++ {
				b := readBuf[index]
				if b == d.Config.MessageSplitter {
					// copy the received data
					dataCloned := make([]byte, len(data))
					copy(dataCloned, data)
					data = nil // reset local buffer
					rawMsg := msgml.NewRawMessage(true, dataCloned)
					//	zap.L().Debug("new message received", zap.Any("rawMessage", rawMsg))
					d.messageLogger.AsyncWrite(rawMsg)
					err := d.receiveMsgFunc(rawMsg)
					if err != nil {
						zap.L().Error("Failed to send a raw message to queue", zap.String("gateway", d.GwCfg.Name), zap.Any("rawMessage", rawMsg), zap.Error(err))
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

func (d *Endpoint) reconnect() {
	// add stuffs to reconnect
}
