package voiddb

import (
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
)

const (
	PluginVoidDB = "void_db"
)

// Client for voiddb
type Client struct {
}

// NewClient creates a dummy client
func NewClient(config cmap.CustomMap) (metricType.Plugin, error) {
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
func (c *Client) Write(data *metricType.InputData) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(data *metricType.InputData) error { return nil }

// Query function
func (c *Client) Query(queryConfig *metricType.QueryConfig) (map[string][]metricType.ResponseData, error) {
	return nil, nil
}
