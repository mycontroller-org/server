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

// Config used across to access the mycontroller
type Config struct {
	URL       string `yaml:"url" mapstructure:"url"`
	Insecure  bool   `yaml:"insecure" mapstructure:"insecure"`
	Username  string `yaml:"username" mapstructure:"username"`
	Password  string `yaml:"password" mapstructure:"password"` // encode as base64
	LoginTime string `yaml:"loginTime" mapstructure:"loginTime"`
	ExpiresIn string `yaml:"expiresIn" mapstructure:"expiresIn"`
}

// GetPassword decodes and returns the password
func (c *Config) GetPassword() string {
	if strings.HasPrefix(c.Password, EncodePrefix) {
		password := strings.Replace(c.Password, EncodePrefix, "", 1)
		decodedPassword, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			log.Fatal("error on decoding the password", err)
		}
		return string(decodedPassword)
	}
	return c.Password
}

// EncodePassword encodes and update the password
func (c *Config) EncodePassword() {
	if c.Password != "" && !strings.HasPrefix(c.Password, EncodePrefix) {
		encodedPassword := base64.StdEncoding.EncodeToString([]byte(c.Password))
		c.Password = fmt.Sprintf("%s%s", EncodePrefix, encodedPassword)
	}
}
