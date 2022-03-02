package javascript_helper

import httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"

type HttpClient struct {
}

func (hc *HttpClient) New(insecure bool, timeout string) *httpclient.Client {
	return httpclient.New(insecure, timeout)
}
