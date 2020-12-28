package natsio

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Implementation guide taken from
// https://github.com/eclipse/paho.mqtt.golang/blob/54b5153cb955d512bc82f56455769e61ca3d21e9/websocket.go

// NewWebsocket returns a net.Conn compatible interface using the gorilla/websocket package
func NewWebsocket(serverURL string, tlsConfig *tls.Config, timeout time.Duration, wsOptions *WebsocketOptions) (net.Conn, error) {
	if wsOptions == nil { // Apply default options
		wsOptions = &WebsocketOptions{}
	}

	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  timeout,
		EnableCompression: false,
		TLSClientConfig:   tlsConfig,
		Subprotocols:      []string{"nats"},
		ReadBufferSize:    wsOptions.ReadBufferSize,
		WriteBufferSize:   wsOptions.WriteBufferSize,
	}

	wsConn, _, err := dialer.Dial(serverURL, wsOptions.RequestHeader)

	if err != nil {
		return nil, err
	}

	connWrapper := &websocketConnector{
		Conn: wsConn,
	}
	return connWrapper, err
}

// websocketConnector is a websocket wrapper so it satisfies the net.Conn interface so it is a
// drop in replacement of the golang.org/x/net/websocket package.
// Implementation guide taken from https://github.com/gorilla/websocket/issues/282
type websocketConnector struct {
	*websocket.Conn
	r   io.Reader
	rio sync.Mutex
	wio sync.Mutex
}

// SetDeadline sets both the read and write deadlines
func (c *websocketConnector) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	err := c.SetWriteDeadline(t)
	return err
}

// Write writes data to the websocket
func (c *websocketConnector) Write(p []byte) (int, error) {
	c.wio.Lock()
	defer c.wio.Unlock()

	err := c.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read reads the current websocket frame
func (c *websocketConnector) Read(p []byte) (int, error) {
	c.rio.Lock()
	defer c.rio.Unlock()
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			}
			// No data read, continue to next message.
			continue
		}
		return n, err
	}
}
