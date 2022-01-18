package email

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// Config of email service
type Config struct {
	Type               string
	Host               string
	Port               int
	AuthType           string
	Username           string
	Password           string `json:"-"`
	FromEmail          string
	ToEmails           string // comma seperated
	InsecureSkipVerify bool
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

// NewEmailPlugin email client
func NewEmailPlugin(cfg *handlerTY.Config) (handlerTY.Plugin, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("Email client", zap.Any("config", config))

	switch config.Type {
	case TypeSMTP, TypeNone:
		return NewSMTPClient(cfg, config)

	default:
		return nil, fmt.Errorf("unknown email client:%s", cfg.Type)
	}
}
