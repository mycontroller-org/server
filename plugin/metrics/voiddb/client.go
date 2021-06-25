package voiddb

import (
	mtsml "github.com/mycontroller-org/server/v2/plugin/metrics"
)

// Client for voiddb
type Client struct {
}

// NewClient creates a dummy client
func NewClient(config map[string]interface{}) (*Client, error) {
	return &Client{}, nil
}

// Close function
func (c *Client) Close() error { return nil }

// Ping function
func (c *Client) Ping() error { return nil }

// Write function
func (c *Client) Write(data *mtsml.InputData) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(data *mtsml.InputData) error { return nil }

// Query function
func (c *Client) Query(queryConfig *mtsml.QueryConfig) (map[string][]mtsml.ResponseData, error) {
	return nil, nil
}
