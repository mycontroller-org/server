package http

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

// Config details
type Config struct {
	URL      string            `json:"url" yaml:"url"`
	Insecure bool              `json:"insecure" yaml:"insecure"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Username string            `json:"username" yaml:"username"`
	Password string            `json:"password" yaml:"password"`
}

// Endpoint data
type Endpoint struct {
	GwCfg     *gwTY.Config
	Config    Config
	Client    *http.Client
	BaseURL   *url.URL
	GatewayID string
}

// RequestConfig used for http request's
type RequestConfig struct {
	Method       string
	Path         string
	ResponseCode int
	QueryParams  map[string]interface{}
}

// ResponseConfig used to return response config
type ResponseConfig struct {
	Method     string
	FullPath   string
	Path       string
	StatusCode int
	Headers    map[string]string
}

// New ethernet driver
func New(gwCfg *gwTY.Config, apiPrefix string) (*Endpoint, error) {
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameNone, gwCfg.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	baseURL := cfg.URL
	if apiPrefix != "" {
		baseURL = fmt.Sprintf("%s%s", cfg.URL, apiPrefix)
	}

	uri, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var client *http.Client

	if cfg.Insecure {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: customTransport}
	} else {
		client = http.DefaultClient
	}

	// update empty map if nil
	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string)
	}

	// add basic auth if enabled
	if cfg.Username != "" {
		auth := fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		cfg.Headers["Authorization"] = fmt.Sprintf("Basic %s", encoded)
	}

	d := &Endpoint{
		Config:  cfg,
		BaseURL: uri,
		Client:  client,
		GwCfg:   gwCfg,
	}
	return d, nil
}

// Write sends a payload
func (ep *Endpoint) Write(rawMsg *msgTY.RawMessage) error {
	requestRaw := rawMsg.Others.Get(gwPtl.KeyHTTPRequestConf)

	if requestRaw == nil {
		return fmt.Errorf("There is no requestConfig found. Have you supplied cfg with key: %s?", gwPtl.KeyHTTPRequestConf)
	}
	reqCfg, ok := requestRaw.(RequestConfig)
	if !ok {
		return fmt.Errorf("Failed to convert request conf, %v", requestRaw)
	}

	if reqCfg.ResponseCode == 0 {
		reqCfg.ResponseCode = 200
	}

	// create request
	_, _, err := ep.newRequest(reqCfg, rawMsg.Data)

	if err != nil {
		return err
	}
	return nil
}

// Close the connection
func (ep *Endpoint) Close() error {
	//return ep.Client.Close()
	return nil
}
