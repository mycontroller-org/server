package extrav1

import (
	"encoding/base64"
	"fmt"
	"time"

	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
)

const (
	timeout = time.Second * 15
)

func newClient(uri string, insecure bool, username, password string) (map[string]string, *httpclient.Client) {
	headers := make(map[string]string)
	if username != "" {
		base64String := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		headers["Authorization"] = fmt.Sprintf("Basic %s", base64String)
	}
	return headers, httpclient.GetClient(insecure, timeout)
}
