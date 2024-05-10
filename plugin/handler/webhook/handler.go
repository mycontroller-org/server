package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginWebhook = "webhook"

	timeout    = time.Second * 10
	loggerName = "handler_webhook"

	// data format
	// DataTypeJSON = "json"
	// DataTypeYAML = "yaml"
	// DataTypeText = "text"
)

// WebhookConfig for webhook
type WebhookConfig struct {
	Server          string
	API             string
	Method          string
	Insecure        bool
	Headers         map[string]string
	QueryParameters map[string]interface{}
	ResponseCode    int
	AllowOverwrite  bool
}

// Clone config data
func (cfg *WebhookConfig) Clone() *WebhookConfig {
	config := &WebhookConfig{
		Server:          cfg.Server,
		API:             cfg.API,
		Insecure:        cfg.Insecure,
		Method:          cfg.Method,
		QueryParameters: make(map[string]interface{}),
		Headers:         make(map[string]string),
		AllowOverwrite:  cfg.AllowOverwrite,
	}

	// update query parameters
	if len(cfg.QueryParameters) > 0 {
		for key, value := range cfg.Headers {
			config.Headers[key] = value
		}
	}

	// update headers
	if len(cfg.Headers) > 0 {
		for key, value := range cfg.QueryParameters {
			config.QueryParameters[key] = value
		}
	}
	return config
}

// WebhookClient struct
type WebhookClient struct {
	HandlerCfg *handlerTY.Config
	Config     *WebhookConfig
	httpClient *httpclient.Client
	logger     *zap.Logger
}

func New(ctx context.Context, handlerCfg *handlerTY.Config) (handlerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	config := &WebhookConfig{}
	err = utils.MapToStruct(utils.TagNameNone, handlerCfg.Spec, config)
	if err != nil {
		return nil, err
	}
	namedLogger := logger.Named(loggerName)

	namedLogger.Debug("webhook client", zap.String("ID", handlerCfg.ID), zap.Any("config", config))

	client := &WebhookClient{
		HandlerCfg: handlerCfg,
		Config:     config,
		logger:     namedLogger,
	}
	return client, nil
}

func (p *WebhookClient) Name() string {
	return PluginWebhook
}

// Start handler implementation
func (c *WebhookClient) Start() error {
	if c.httpClient == nil {
		c.httpClient = httpclient.GetClient(c.Config.Insecure, timeout)
	}

	return nil
}

// Close handler implementation
func (c *WebhookClient) Close() error {
	c.httpClient = nil
	return nil

}

// State implementation
func (c *WebhookClient) State() *types.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &types.State{}
		}
		return c.HandlerCfg.State
	}
	return &types.State{}
}

// Post handler implementation
func (c *WebhookClient) Post(parameters map[string]interface{}) error {
	config := c.Config.Clone()

	for name, rawParameter := range parameters {
		parameter, ok := handlerTY.IsTypeOf(rawParameter, handlerTY.DataTypeWebhook)
		if !ok {
			continue
		}
		c.logger.Debug("data", zap.Any("name", name), zap.Any("parameter", parameter))

		webhookData := handlerTY.WebhookData{}
		err := utils.MapToStruct(utils.TagNameNone, parameter, &webhookData)
		if err != nil {
			c.logger.Error("error on converting webhook data", zap.Error(err), zap.String("name", name), zap.Any("parameter", parameter))
			continue
		}

		// overide basic config, if any
		if config.AllowOverwrite {
			if webhookData.Server != "" {
				config.Server = webhookData.Server
			}

			if webhookData.API != "" {
				config.API = webhookData.API
			}

			if webhookData.Method != "" {
				config.Method = webhookData.Method
			}

			if webhookData.ResponseCode != 0 {
				config.ResponseCode = webhookData.ResponseCode
			}

			if len(webhookData.QueryParameters) > 0 {
				for key, value := range webhookData.Headers {
					config.Headers[key] = value
				}
			}

			if len(webhookData.Headers) > 0 {
				for key, value := range webhookData.QueryParameters {
					config.QueryParameters[key] = value
				}
			}
		}

		if config.Method == "" {
			config.Method = http.MethodPost
		}

		url := fmt.Sprintf("%s%s", config.Server, config.API)

		respose, err := c.httpClient.ExecuteJson(url, config.Method, config.Headers, config.QueryParameters, webhookData.Data, config.ResponseCode)
		responseCode := 0
		if respose != nil {
			responseCode = respose.StatusCode
		}
		if err != nil {
			c.logger.Error("error on webhook handler call", zap.Int("responseStatusCode", responseCode), zap.Error(err))
			return err
		}
	}

	return nil
}
