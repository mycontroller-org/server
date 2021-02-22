package noop

// Client struct
type Client struct {
}

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error { return nil }

// Post handler implementation
func (c *Client) Post(variables map[string]interface{}) error { return nil }
