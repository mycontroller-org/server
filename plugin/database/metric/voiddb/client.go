package voiddb

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
)

const (
	PluginVoidDB = "void_db"
)

// Client for voiddb
type Client struct {
}

// NewClient creates a dummy client
func NewClient(config cmap.CustomMap) (metricTY.Plugin, error) {
	return &Client{}, nil
}

func (c *Client) Name() string {
	return PluginVoidDB
}

// Close function
func (c *Client) Close() error { return nil }

// Ping function
func (c *Client) Ping() error { return nil }

// Write function
func (c *Client) Write(data *metricTY.InputData) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(data *metricTY.InputData) error { return nil }

// Query function
func (c *Client) Query(queryConfig *metricTY.QueryConfig) (map[string][]metricTY.ResponseData, error) {
	return nil, nil
}
