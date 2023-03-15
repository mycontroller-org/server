package telegram

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpClient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginTelegram = "telegram"

	timeout    = time.Second * 15
	loggerName = "handler_telegram"
)

// TelegramClient struct
type TelegramClient struct {
	handlerCfg *handlerTY.Config
	Config     *Config
	httpClient *httpClient.Client
	logger     *zap.Logger
}

// telegram handler
func New(ctx context.Context, cfg *handlerTY.Config) (handlerTY.Plugin, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}

	client := &TelegramClient{
		httpClient: httpClient.GetClient(false, timeout),
		handlerCfg: cfg,
		Config:     config,
		logger:     logger.Named(loggerName),
	}

	user, err := client.GetMe()
	if err != nil {
		return nil, err
	}

	client.logger.Info("telegram auth success", zap.String("handlerID", cfg.ID), zap.Any("firstname", user.FirstName))
	return client, nil
}

func (p *TelegramClient) Name() string {
	return PluginTelegram
}

// Start handler implementation
func (c *TelegramClient) Start() error { return nil }

// Close handler implementation
func (c *TelegramClient) Close() error { return nil }

// State implementation
func (c *TelegramClient) State() *types.State {
	if c.handlerCfg != nil {
		if c.handlerCfg.State == nil {
			c.handlerCfg.State = &types.State{}
		}
		return c.handlerCfg.State
	}
	return &types.State{}
}

// Post handler implementation
func (c *TelegramClient) Post(parameters map[string]interface{}) error {

	for name, rawParameter := range parameters {
		parameter, ok := handlerTY.IsTypeOf(rawParameter, handlerTY.DataTypeTelegram)
		if !ok {
			continue
		}
		c.logger.Debug("data", zap.Any("name", name), zap.Any("parameter", parameter))

		telegramData := handlerTY.TelegramData{}
		err := utils.MapToStruct(utils.TagNameNone, parameter, &telegramData)
		if err != nil {
			c.logger.Error("error on converting telegram data", zap.Error(err), zap.String("name", name), zap.Any("parameter", parameter))
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
				c.logger.Error("error on telegram sendMessage", zap.Error(err), zap.String("parameter", name))
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			for _, err := range errors {
				c.logger.Error("telegram sendMessage error", zap.String("id", c.handlerCfg.ID), zap.Error(err))
			}
			return errors[0]
		}
		c.logger.Debug("telegram sendMessage success", zap.String("id", c.handlerCfg.ID), zap.String("timeTaken", time.Since(start).String()))

	}
	return nil
}

// SendMessage func
func (c *TelegramClient) SendMessage(message *Message) error {
	url := c.getURL(APISendMessage)
	response, err := c.httpClient.ExecuteJson(url, http.MethodPost, nil, nil, message, 200)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(response.Body, resp)
	if err != nil {
		return nil
	}

	if !resp.IsOK {
		return fmt.Errorf("request failed: %+v", resp)
	}
	return nil
}

// GetMe returns about the user profile
func (c *TelegramClient) GetMe() (*User, error) {
	url := c.getURL(APIGetMe)
	response, err := c.httpClient.ExecuteJson(url, http.MethodGet, nil, nil, nil, 200)
	if err != nil {
		return nil, err
	}
	resp := &Response{}
	err = json.Unmarshal(response.Body, resp)
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

func (c *TelegramClient) getURL(api string) string {
	return fmt.Sprintf("%s/bot%s%s", ServerURL, c.Config.Token, api)
}
