package email

import (
	"fmt"
	"net/smtp"

	"go.uber.org/zap"
)

// smtp client
type smtpClient struct {
	cfg  *Config
	auth smtp.Auth
}

// init smtp client
func initSMTP(cfg *Config) (Client, error) {
	zap.L().Debug("Init smtp email client", zap.Any("config", cfg))

	var auth smtp.Auth

	if cfg.AuthType == AuthTypePlain || cfg.AuthType == "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	} else if cfg.AuthType == AuthTypeCRAMMD5 {
		auth = smtp.CRAMMD5Auth(cfg.Username, cfg.Password)
	} else {
		return nil, fmt.Errorf("Unknown auth type:%s", cfg.AuthType)
	}

	client := &smtpClient{
		cfg:  cfg,
		auth: auth,
	}
	return client, nil
}

// Close func implementation
func (sc *smtpClient) Close() error {
	return nil
}

// Send func implementation
func (sc *smtpClient) Send(from string, to []string, subject, body string) error {
	// set from address as username if non set
	if from == "" {
		from = sc.cfg.Username
	}

	addr := fmt.Sprintf("%s:%d", sc.cfg.Host, sc.cfg.Port)

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	emailSubject := "Subject: " + "Test Email" + "!\n"
	msg := []byte(emailSubject + mime + "\n" + body)

	return smtp.SendMail(addr, sc.auth, from, to, msg)
}
