package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	httpclient "github.com/mycontroller-org/backend/v2/pkg/utils/http_client_json"
	variableUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

// data format
const (
	DataTypeJSON = "json"
	DataTypeYAML = "yaml"
	DataTypeText = "text"
)

// Config for webhook
type Config struct {
	Server             string
	API                string
	Method             string
	InsecureSkipVerify bool
	Headers            map[string]string
	QueryParameters    map[string]interface{}
	ResponseCode       int
	AllowOverride      bool
}

// Clone config data
func (cfg *Config) Clone() *Config {
	config := &Config{
		Server:             cfg.Server,
		API:                cfg.API,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Method:             cfg.Method,
		QueryParameters:    make(map[string]interface{}),
		Headers:            make(map[string]string),
		AllowOverride:      cfg.AllowOverride,
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

// Client struct
type Client struct {
	HandlerCfg *handlerML.Config
	Config     *Config
	httpClient *httpclient.Client
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
func (c *Client) Start() error {
	if c.httpClient == nil {
		c.httpClient = httpclient.GetClient(c.Config.InsecureSkipVerify)
	}

	return nil
}

// Close handler implementation
func (c *Client) Close() error {
	c.httpClient = nil
	return nil

}

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

	for name, value := range data {
		zap.L().Debug("data", zap.Any("name", name), zap.Any("value", value))
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerML.DataTypeWebhook {
			continue
		}

		webhookData := handlerML.WebhookData{}
		err = variableUtils.UnmarshalBase64Yaml(genericData.Data, &webhookData)
		if err != nil {
			zap.L().Error("error on converting webhook data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
			continue
		}

		// overide basic config, if any
		if config.AllowOverride {
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

		respose, _, err := c.httpClient.Request(url, config.Method, config.Headers, config.QueryParameters, webhookData.Data, config.ResponseCode)
		responseCode := 0
		if respose != nil {
			responseCode = respose.StatusCode
		}
		if err != nil {
			zap.L().Error("error on webhook handler call", zap.Int("responseStatusCode", responseCode), zap.Error(err))
		}
		return err
	}

	return nil
}
