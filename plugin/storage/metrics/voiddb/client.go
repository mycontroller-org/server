package voiddb

import ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"

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
func (c *Client) Write(variable *ml.SensorField) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(variable *ml.SensorField) error { return nil }
