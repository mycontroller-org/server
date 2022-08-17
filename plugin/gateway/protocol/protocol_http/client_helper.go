package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	json "github.com/mycontroller-org/server/v2/pkg/json"
)

func (ep *Endpoint) newRequest(cfg RequestConfig, body interface{}) (*ResponseConfig, []byte, error) {
	relativePath := &url.URL{Path: cfg.Path}
	fullPath := ep.BaseURL.ResolveReference(relativePath)

	// add body, if available
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, nil, err
		}
	}
	req, err := http.NewRequest(cfg.Method, fullPath.String(), buf)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// set as json content
	req.Header.Set("Accept", "application/json")
	// update headers
	for k, v := range ep.Config.Headers {
		req.Header.Set(k, v)
	}

	if cfg.QueryParams != nil {
		q := req.URL.Query()
		for k, v := range cfg.QueryParams {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := ep.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != cfg.ResponseCode {
		return nil, nil, fmt.Errorf("Failed with status code. [url:%v, status: %v, statusCode: %v]", fullPath.String(), resp.Status, resp.StatusCode)
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	respCfg := &ResponseConfig{
		StatusCode: resp.StatusCode,
		FullPath:   fullPath.String(),
		Path:       cfg.Path,
		Method:     cfg.Method,
	}

	// update headers
	respCfg.Headers = make(map[string]string)
	for k := range resp.Header {
		respCfg.Headers[k] = resp.Header.Get(k)
	}

	return respCfg, respBodyBytes, nil
}
