package httpclient

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"go.uber.org/zap"
)

const DefaultTimeout = time.Second * 30

// ResponseConfig of a request
type ResponseConfig struct {
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"-"`
}

// returns body as string
func (rc *ResponseConfig) StringBody() string {
	return string(rc.Body)
}

// Client struct
type Client struct {
	httpClient *http.Client
}

// GetClient returns http client
func GetClient(insecure bool, timeout time.Duration) *Client {
	var httpClient *http.Client
	if insecure {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: customTransport}
	} else {
		httpClient = http.DefaultClient
	}
	if timeout > 0 {
		httpClient.Timeout = timeout
	} else {
		httpClient.Timeout = DefaultTimeout
	}

	// create cookiejar
	jar, err := cookiejar.New(nil)
	if err != nil {
		zap.L().Warn("error on cookiejar creation, continues without cookie jar", zap.Error(err))
	} else {
		httpClient.Jar = jar
	}

	return &Client{httpClient: httpClient}
}

// ExecuteJson execute http request and returns response
func (c *Client) ExecuteJson(url, method string, headers map[string]string, queryParams map[string]interface{},
	body interface{}, responseCode int) (*ResponseConfig, error) {
	// add body, if available
	var buf io.ReadWriter
	if method != http.MethodGet && body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// set as json content
	req.Header.Set("Accept", "application/json")
	// update headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if queryParams != nil {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if responseCode > 0 && resp.StatusCode != responseCode {
		return nil, fmt.Errorf("failed with status code. [status: %v, statusCode: %v, body: %s]", resp.Status, resp.StatusCode, string(respBodyBytes))
	}

	respCfg := &ResponseConfig{
		StatusCode: resp.StatusCode,
		URL:        url,
		Method:     method,
		Body:       respBodyBytes,
		Headers:    make(map[string]string),
	}

	// update headers
	for k := range resp.Header {
		respCfg.Headers[k] = resp.Header.Get(k)
	}

	return respCfg, nil
}

// Request implementation
func (c *Client) Execute(url, method string, headers map[string]string,
	queryParams map[string]interface{}, body string, responseCode int) (*ResponseConfig, error) {

	// add body, if available
	var buf io.ReadWriter
	if body != "" {
		buf = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	// update headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// update query parameters
	if queryParams != nil {
		q := req.URL.Query()
		for k, v := range queryParams {
			if vSlice, ok := v.([]string); ok {
				for _, sValue := range vSlice {
					q.Add(k, sValue)
				}
			} else {
				q.Add(k, fmt.Sprintf("%v", v))
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if responseCode > 0 && resp.StatusCode != responseCode {
		return nil, fmt.Errorf("failed with status code. [status: %v, statusCode: %v, body: %s]", resp.Status, resp.StatusCode, string(respBodyBytes))
	}

	respCfg := &ResponseConfig{
		StatusCode: resp.StatusCode,
		URL:        url,
		Method:     method,
		Body:       respBodyBytes,
		Headers:    make(map[string]string),
	}

	// update headers
	for k := range resp.Header {
		respCfg.Headers[k] = resp.Header.Get(k)
	}

	return respCfg, nil
}
