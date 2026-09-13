package client

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

const (
	EncodePrefix = "BASE64/"
)

// Config holds named server aliases.
type Config struct {
	Aliases map[string]Alias `json:"aliases" yaml:"aliases" mapstructure:"aliases"`
}

// Alias is one named connection (server + user session).
type Alias struct {
	URL       string `json:"url" yaml:"url" mapstructure:"url"`
	Insecure  bool   `json:"insecure" yaml:"insecure" mapstructure:"insecure"`
	Username  string `json:"username" yaml:"username" mapstructure:"username"`
	Password  string `json:"password" yaml:"password" mapstructure:"password"`
	LoginTime string `json:"loginTime" yaml:"loginTime" mapstructure:"loginTime"`
	ExpiresIn string `json:"expiresIn" yaml:"expiresIn" mapstructure:"expiresIn"`
}

func (c *Config) EnsureAliases() {
	if c.Aliases == nil {
		c.Aliases = map[string]Alias{}
	}
}

func (a *Alias) GetPassword() string {
	if strings.HasPrefix(a.Password, EncodePrefix) {
		password := strings.Replace(a.Password, EncodePrefix, "", 1)
		decodedPassword, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			log.Fatal("error on decoding the password", err)
		}
		return string(decodedPassword)
	}
	return a.Password
}

func (a *Alias) EncodePassword() {
	if a.Password != "" && !strings.HasPrefix(a.Password, EncodePrefix) {
		encodedPassword := base64.StdEncoding.EncodeToString([]byte(a.Password))
		a.Password = fmt.Sprintf("%s%s", EncodePrefix, encodedPassword)
	}
}

func (c *Config) EncodePasswords() {
	c.EnsureAliases()
	for name, alias := range c.Aliases {
		alias.EncodePassword()
		c.Aliases[name] = alias
	}
}
