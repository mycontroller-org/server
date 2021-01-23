package email

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
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
	InsecureSkipVerify bool
}

// Client for email service
type Client interface {
	Close() error
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

// Init email client
func Init(config cmap.CustomMap) (Client, error) {
	cfg := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, cfg)
	if err != nil {
		return nil, err
	}
	zap.L().Debug("Email client", zap.Any("config", cfg))

	switch cfg.Type {
	case TypeSMTP, TypeNone:
		return initSMTP(cfg)

	default:
		return nil, fmt.Errorf("Unknown email client:%s", cfg.Type)
	}
}
