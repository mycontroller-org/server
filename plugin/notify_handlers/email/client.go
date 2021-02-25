package email

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
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
	defaultSubject = "Email from MyController.org server"
)

// Client for email service
type Client interface {
	Start() error
	Post(variables map[string]interface{}) error
	Send(from string, to []string, subject, body string) error
	Close() error
	State() *model.State
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

// Init email client
func Init(cfg *handlerML.Config) (Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("Email client", zap.Any("config", config))

	switch config.Type {
	case TypeSMTP, TypeNone:
		return initSMTP(cfg, config)

	default:
		return nil, fmt.Errorf("Unknown email client:%s", cfg.Type)
	}
}
