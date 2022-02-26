package http_generic

import (
	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

const (
	ScriptKeyDataIn   = "dataIn"
	ScriptKeyDataOut  = "dataOut"
	ScriptKeyConfigIn = "configIn"

	DefaultNode = "default"
)

// http protocol
type HttpProtocol struct {
	GatewayConfig     *gwTY.Config
	Config            *HttpProtocolConf
	rawMessageHandler func(rawMsg *msgTY.RawMessage) error
}

// http protocol config
type HttpProtocolConf struct {
	Type            string                 `json:"type"`
	Headers         map[string]string      `json:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters"`
	Endpoints       map[string]HttpConfig  `json:"endpoints"`
	Nodes           map[string]HttpNode    `json:"nodes"`
}

// http config
type HttpConfig struct {
	HttpNode
	Disabled          bool   `json:"disabled"`
	ExecutionInterval string `json:"executionInterval"`
}

// nodes details

// http node config
type HttpNode struct {
	URL                 string                 `json:"url"`
	Method              string                 `json:"method"`
	Insecure            bool                   `json:"insecure"`
	Headers             map[string]string      `json:"headers"`
	QueryParameters     map[string]interface{} `json:"queryParameters"`
	Body                cmap.CustomMap         `json:"body"`
	ResponseCode        int                    `json:"responseCode"`
	Script              string                 `json:"script"`
	IncludeGlobalConfig bool                   `json:"includeGlobalConfig"`
}

// Clone cones HttpNode
func (hn *HttpNode) Clone() *HttpNode {
	cloned := &HttpNode{
		URL:                 hn.URL,
		Method:              hn.Method,
		Insecure:            hn.Insecure,
		Headers:             make(map[string]string),
		QueryParameters:     make(map[string]interface{}),
		Body:                hn.Body,
		ResponseCode:        hn.ResponseCode,
		Script:              hn.Script,
		IncludeGlobalConfig: hn.IncludeGlobalConfig,
	}

	// update headers
	for k, v := range hn.Headers {
		cloned.Headers[k] = v
	}

	// update query parameters
	for k, v := range hn.QueryParameters {
		cloned.QueryParameters[k] = v
	}

	return cloned
}

func toHttpNode(config interface{}) (*HttpNode, error) {
	node := &HttpNode{}
	err := json.ToStruct(config, node)
	return node, err
}
