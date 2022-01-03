package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
	"go.uber.org/zap"
)

// smtp client
type SmtpClient struct {
	handlerCfg *handlerType.Config
	cfg        *Config
	auth       smtp.Auth
}

// init smtp client
func NewSMTPClient(handlerCfg *handlerType.Config, cfg *Config) (Client, error) {
	var auth smtp.Auth

	if cfg.AuthType == AuthTypePlain || cfg.AuthType == "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	} else if cfg.AuthType == AuthTypeCRAMMD5 {
		auth = smtp.CRAMMD5Auth(cfg.Username, cfg.Password)
	} else {
		return nil, fmt.Errorf("unknown auth type:%s", cfg.AuthType)
	}

	client := &SmtpClient{
		handlerCfg: handlerCfg,
		cfg:        cfg,
		auth:       auth,
	}
	zap.L().Info("init smtp email client success", zap.Any("handlerID", handlerCfg.ID))
	return client, nil
}

func (sc *SmtpClient) Name() string {
	return PluginEmail
}

func (sc *SmtpClient) Start() error {
	// nothing to do here
	return nil
}

// Close func implementation
func (sc *SmtpClient) Close() error {
	return nil
}

func (sc *SmtpClient) State() *types.State {
	if sc.handlerCfg != nil {
		if sc.handlerCfg.State == nil {
			sc.handlerCfg.State = &types.State{}
		}
		return sc.handlerCfg.State
	}
	return &types.State{}
}

// Send func implementation
func (sc *SmtpClient) Send(from string, to []string, subject, body string) error {
	// set from address as username if non set
	if from == "" {
		from = sc.cfg.Username
	}

	addr := fmt.Sprintf("%s:%d", sc.cfg.Host, sc.cfg.Port)
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + "\n" + mime + "\n" + body)
	return smtp.SendMail(addr, sc.auth, from, to, msg)
}

func (sc *SmtpClient) sendEmailSSL(from string, to []string, subject, body string) error {
	// set from address as username if non set
	if from == "" {
		from = sc.cfg.Username
	}

	servername := fmt.Sprintf("%s:%d", sc.cfg.Host, sc.cfg.Port)
	host, _, err := net.SplitHostPort(servername)
	if err != nil {
		return err
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: sc.cfg.InsecureSkipVerify,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = client.Auth(sc.auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, toAddr := range to {
		if err = client.Rcpt(toAddr); err != nil {
			return err
		}
	}

	write, err := client.Data()
	if err != nil {
		return err
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	if _, err = write.Write([]byte(message)); err != nil {
		return err
	}

	if err = write.Close(); err != nil {
		return err
	}

	return client.Quit()
}

// Post performs send operation
func (sc *SmtpClient) Post(data map[string]interface{}) error {
	for name, value := range data {
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerType.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerType.DataTypeEmail {
			continue
		}

		emailData := handlerType.EmailData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &emailData)
		if err != nil {
			zap.L().Error("error on converting email data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
			continue
		}

		fromEmail := sc.cfg.FromEmail
		toEmails := sc.cfg.ToEmails
		subject := defaultSubject
		body := ""

		if emailData.From != "" {
			fromEmail = emailData.From
		}

		if len(emailData.To) > 0 {
			toEmails = strings.Join(emailData.To, ",")
		}

		if emailData.Subject != "" {
			subject = emailData.Subject
		}
		if emailData.Body != "" {
			body = emailData.Body
		}
		to := strings.Split(toEmails, ",")

		start := time.Now()
		err = sc.sendEmailSSL(fromEmail, to, subject, body)
		if err != nil {
			zap.L().Error("error on email sent", zap.Error(err))
		}
		zap.L().Debug("email sent", zap.String("id", sc.handlerCfg.ID), zap.String("timeTaken", time.Since(start).String()))
	}
	return nil
}
