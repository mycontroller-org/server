package webhook

import (
	"net/http"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

// Config for webhook
type Config struct {
	Server       string
	API          string
	Method       string
	Headers      map[string]string
	AllowOveride bool
}

// Clone config data
func (cfg *Config) Clone() *Config {
	config := &Config{
		Server:       cfg.Server,
		API:          cfg.API,
		Method:       cfg.Method,
		AllowOveride: cfg.AllowOveride,
		Headers:      make(map[string]string),
	}
	// update headers
	if len(cfg.Headers) > 0 {
		for key, value := range cfg.Headers {
			config.Headers[key] = value
		}
	}
	return config
}

// Client struct
type Client struct {
	HandlerCfg *handlerML.Config
	Config     *Config
}

// Init the webhook client
func Init(handlerCfg *handlerML.Config) (*Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, handlerCfg.Spec, config)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("Webhook client", zap.String("ID", handlerCfg.ID), zap.Any("config", config))

	client := &Client{
		HandlerCfg: handlerCfg,
		Config:     config,
	}
	return client, nil
}

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error { return nil }

// State implementation
func (c *Client) State() *model.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &model.State{}
		}
		return c.HandlerCfg.State
	}
	return &model.State{}
}

// Post handler implementation
func (c *Client) Post(data map[string]interface{}) error {
	config := c.Config.Clone()
	// overide basic config, if any
	if config.AllowOveride {
		for name, value := range data {
			zap.L().Info("data", zap.Any("name", name), zap.Any("value", value))
		}
	}

	if config.Method == "" {
		config.Method = http.MethodPost
	}

	switch config.Method {
	case http.MethodPost, http.MethodPut:

	case http.MethodGet:

	}

	return nil
}
