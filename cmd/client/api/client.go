package api

import (
	"fmt"

	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	httpUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
)

type Client struct {
	ServerAddress string
	Token         string
	Insecure      bool
	Timeout       string
}

func NewClient(serverAddress string, token string, insecure bool) *Client {
	return &Client{ServerAddress: serverAddress, Token: token, Insecure: insecure}
}

func (c *Client) executeJson(api, httpMethod string, headers map[string]string, queryParams map[string]interface{},
	body interface{}, responseCode int) (*httpUtils.ResponseConfig, error) {
	client := httpUtils.New(c.Insecure, c.Timeout)
	return client.ExecuteJson(fmt.Sprintf("%s%s", c.ServerAddress, api), httpMethod, c.getHeaders(headers), queryParams, body, responseCode)
}

func (c *Client) getHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		headers = map[string]string{}
	}
	headers[handlerTY.HeaderAuthorization] = c.Token
	return headers
}
