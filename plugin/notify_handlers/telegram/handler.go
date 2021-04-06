package telegram

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	httpClient "github.com/mycontroller-org/backend/v2/pkg/utils/http_client_json"
	variableUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

// Client struct
type Client struct {
	handlerCfg *handlerML.Config
	Config     *Config
	httpClient *httpClient.Client
}

// Init Telegram client
func Init(cfg *handlerML.Config) (*Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: httpClient.GetClient(false),
		handlerCfg: cfg,
		Config:     config,
	}

	user, err := client.GetMe()
	if err != nil {
		return nil, err
	}

	zap.L().Info("Telegram auth success", zap.String("handlerID", cfg.ID), zap.Any("firstname", user.FirstName))
	return client, nil
}

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error { return nil }

// State implementation
func (c *Client) State() *model.State {
	if c.handlerCfg != nil {
		if c.handlerCfg.State == nil {
			c.handlerCfg.State = &model.State{}
		}
		return c.handlerCfg.State
	}
	return &model.State{}
}

// Post handler implementation
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerML.DataTypeTelegram {
			continue
		}

		telegramData := handlerML.TelegramData{}
		err = variableUtils.UnmarshalBase64Yaml(genericData.Data, &telegramData)
		if err != nil {
			zap.L().Error("error on converting telegram data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
			continue
		}

		chatIDs := c.Config.ChatIDs

		if len(telegramData.ChatIDs) > 0 {
			chatIDs = telegramData.ChatIDs
		}

		parseMode := telegramData.ParseMode
		if parseMode == ParseModeText {
			parseMode = ""
		}

		if len(chatIDs) == 0 {
			continue
		}

		start := time.Now()
		errors := make([]error, 0)
		for _, chatID := range chatIDs {
			msg := &Message{
				ChatID:    chatID,
				ParseMode: parseMode,
				Text:      telegramData.Text,
			}
			err = c.SendMessage(msg)
			if err != nil {
				zap.L().Error("error on telegram sendMessage", zap.Error(err), zap.String("parameter", name))
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			for _, err := range errors {
				zap.L().Error("telegram sendMessage error", zap.String("id", c.handlerCfg.ID), zap.Error(err))
			}
			return errors[0]
		}
		zap.L().Debug("telegram sendMessage success", zap.String("id", c.handlerCfg.ID), zap.String("timeTaken", time.Since(start).String()))

	}
	return nil
}

// SendMessage func
func (c *Client) SendMessage(message *Message) error {
	url := c.getURL(APISendMessage)
	_, data, err := c.httpClient.Request(url, http.MethodPost, nil, nil, message, 200)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil
	}

	if !resp.IsOK {
		return fmt.Errorf("request failed: %+v", resp)
	}
	return nil
}

// GetMe returns about the user profile
func (c *Client) GetMe() (*User, error) {
	url := c.getURL(APIGetMe)
	_, data, err := c.httpClient.Request(url, http.MethodGet, nil, nil, nil, 200)
	if err != nil {
		return nil, err
	}
	resp := &Response{}
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK {
		return nil, fmt.Errorf("request failed: %+v", resp)
	}

	user := &User{}
	err = utils.MapToStruct(utils.TagNameJSON, resp.Result, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *Client) getURL(api string) string {
	return fmt.Sprintf("%s/bot%s%s", ServerURL, c.Config.Token, api)
}
