package httpclient

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	json "github.com/mycontroller-org/backend/v2/pkg/json"
)

// ResponseConfig of a request
type ResponseConfig struct {
	Method     string
	URL        string
	StatusCode int
	Headers    map[string]string
}

// Client struct
type Client struct {
	httpClient *http.Client
}

// GetClient returns http client
func GetClient(insecure bool) *Client {
	var httpClient *http.Client
	if insecure {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: customTransport}
	} else {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

// Request implementation
func (c *Client) Request(url, method string, headers map[string]string, queryParams map[string]interface{}, body interface{}, responseCode int) (*ResponseConfig, []byte, error) {
	// add body, if available
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if responseCode > 0 {
		if resp.StatusCode != responseCode {
			return nil, nil, fmt.Errorf("Failed with status code. [url:%v, status: %v, statusCode: %v, body: %s]", url, resp.Status, resp.StatusCode, string(respBodyBytes))
		}
	}

	respCfg := &ResponseConfig{
		StatusCode: resp.StatusCode,
		URL:        url,
		Method:     method,
	}

	// update headers
	respCfg.Headers = make(map[string]string)
	for k := range resp.Header {
		respCfg.Headers[k] = resp.Header.Get(k)
	}

	return respCfg, respBodyBytes, nil
}
