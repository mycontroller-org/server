package voiddb

import (
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metrics"
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
func (c *Client) Write(variable *fml.Field) error { return nil }

// WriteBlocking function
func (c *Client) WriteBlocking(variable *fml.Field) error { return nil }

// Query function
func (c *Client) Query(queryConfig *mtrml.QueryConfig) (map[string][]mtrml.Data, error) {
	return nil, nil
}
