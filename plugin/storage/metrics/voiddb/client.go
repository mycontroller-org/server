package voiddb

import (
	sml "github.com/mycontroller-org/backend/pkg/model/sensor"
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
func (c *Client) Write(variable *sml.SensorField) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(variable *sml.SensorField) error { return nil }
