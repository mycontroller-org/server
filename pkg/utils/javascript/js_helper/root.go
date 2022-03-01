package javascript_helper

import (
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
)

const (
	KeyMcUtils = "mcUtils"

	KeyCrypto     = "crypto"
	KeyLowLevel   = "lowLevel"
	KeyHttpClient = "httpClient"
)

// includes helper functions
func GetHelperUtils() map[string]interface{} {
	return map[string]interface{}{
		KeyCrypto:     getCrypto(),
		KeyLowLevel:   &LowLevel{},
		KeyHttpClient: httpclient.New,
	}
}
