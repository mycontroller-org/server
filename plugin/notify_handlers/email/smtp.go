package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/utils"
	tplUtils "github.com/mycontroller-org/backend/v2/pkg/utils/template"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// smtp client
type smtpClient struct {
	id   string
	cfg  *Config
	auth smtp.Auth
}

// init smtp client
func initSMTP(id string, cfg *Config) (Client, error) {
	zap.L().Info("Init smtp email client", zap.Any("config", cfg))

	var auth smtp.Auth

	if cfg.AuthType == AuthTypePlain || cfg.AuthType == "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	} else if cfg.AuthType == AuthTypeCRAMMD5 {
		auth = smtp.CRAMMD5Auth(cfg.Username, cfg.Password)
	} else {
		return nil, fmt.Errorf("Unknown auth type:%s", cfg.AuthType)
	}

	client := &smtpClient{
		id:   id,
		cfg:  cfg,
		auth: auth,
	}
	return client, nil
}

func (sc *smtpClient) Start() error {
	// nothing to do here
	return nil
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
	msg := []byte(subject + "\n" + mime + "\n" + body)
	return smtp.SendMail(addr, sc.auth, from, to, msg)
}

func (sc *smtpClient) sendEmailSSL(from string, to []string, subject, body string) error {
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
func (sc *smtpClient) Post(variables map[string]interface{}) error {
	fromEmail := sc.cfg.FromEmail
	toEmails := sc.cfg.ToEmails
	subject := defaultSubject
	body := ""

	if from, ok := variables[keyFromEmail]; ok {
		fromEmail = utils.ToString(from)
	}

	if to, ok := variables[keyToEmails]; ok {
		toEmails = utils.ToString(to)
	}

	if sub, ok := variables[keySubject]; ok {
		subjectRaw := utils.ToString(sub)
		subjectFinal, err := tplUtils.Execute(subjectRaw, variables)
		if err != nil {
			subject = "Error: " + err.Error()
		}
		subject = subjectFinal
	}

	if bd, ok := variables[keyBody]; ok {
		bodyRaw := utils.ToString(bd)
		bodyFinal, err := tplUtils.Execute(bodyRaw, variables)
		if err != nil {
			body = "Error: " + err.Error()
		} else {
			body = bodyFinal
		}
	} else {
		bodyBytes, err := yaml.Marshal(variables)
		if err != nil {
			body = "Marshal Error: " + err.Error()
		} else {
			body = string(bodyBytes)
		}
	}

	to := strings.Split(toEmails, ",")

	start := time.Now()
	err := sc.sendEmailSSL(fromEmail, to, subject, body)
	if err != nil {
		zap.L().Error("error on email sent", zap.Error(err))
	}
	zap.L().Debug("email sent", zap.String("id", sc.id), zap.String("timeTaken", time.Since(start).String()))
	return err
}
