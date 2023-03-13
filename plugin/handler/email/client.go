package email

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// Config of email service
type Config struct {
	Type      string
	Host      string
	Port      int
	AuthType  string
	Username  string
	Password  string `json:"-" yaml:"-"`
	FromEmail string
	ToEmails  string // comma separated
	Insecure  bool
}

const (
	defaultSubject = "Email from MyController server"
	PluginEmail    = "email"
)

// Client for email service
type Client interface {
	Name() string
	Start() error
	Close() error
	Post(variables map[string]interface{}) error
	State() *types.State
	Send(from string, to []string, subject, body string) error
}

// service provider types
const (
	TypeNone = ""
	TypeSMTP = "smtp"
)

// authentication options
const (
	AuthTypePlain   = "plain"
	AuthTypeCRAMMD5 = "crammd5"
)

// email client
func New(ctx context.Context, rawCfg *handlerTY.Config) (handlerTY.Plugin, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = utils.MapToStruct(utils.TagNameNone, rawCfg.Spec, config)
	if err != nil {
		return nil, err
	}
	logger.Debug("email client", zap.Any("config", config))

	switch config.Type {
	case TypeSMTP, TypeNone:
		return NewSMTPClient(ctx, logger, rawCfg, config)

	default:
		return nil, fmt.Errorf("unknown email client:%s", rawCfg.Type)
	}
}
