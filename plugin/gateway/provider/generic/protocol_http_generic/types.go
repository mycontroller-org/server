package http_generic

import (
	"github.com/mycontroller-org/server/v2/pkg/json"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	ScriptKeyDataIn          = "dataIn"          // post data to script
	ScriptKeyDataOut         = "dataOut"         // get data from script
	ScriptKeyConfigIn        = "configIn"        // post config data to script
	ScriptKeyPreRunResponse  = "preRunResponse"  // post pre run response to script
	ScriptKeyPostRunResponse = "postRunResponse" // post post-run response to script
	ScriptKeyExecute         = "execute"         // get execute binary signal from script
	DefaultNode              = "default"         // default node endpoint name, if none matches

	BodyLanguageJSON      = "json"
	BodyLanguageYAML      = "yaml"
	BodyLanguagePlainText = "plaintext"

	PreRun  = "pre_run"
	PostRun = "post_run"
)

// http protocol
type HttpProtocol struct {
	GatewayConfig     *gwTY.Config
	Config            *HttpProtocolConf
	rawMessageHandler func(rawMsg *msgTY.RawMessage) error
	logger            *zap.Logger
	scheduler         schedulerTY.CoreScheduler
}

// http protocol config
type HttpProtocolConf struct {
	Type            string                 `json:"type" yaml:"type"`
	Headers         map[string]string      `json:"headers" yaml:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters" yaml:"queryParameters"`
	Endpoints       map[string]HttpConfig  `json:"endpoints" yaml:"endpoints"`
	Nodes           map[string]HttpNode    `json:"nodes" yaml:"nodes"`
}

// http config
type HttpConfig struct {
	HttpNode
	Disabled          bool   `json:"disabled" yaml:"disabled"`
	ExecutionInterval string `json:"executionInterval" yaml:"executionInterval"`
}

// nodes details

// http node
type HttpNode struct {
	HttpNodeConfig
	Insecure bool                      `json:"insecure" yaml:"insecure"`
	PreRun   map[string]HttpNodeConfig `json:"preRun" yaml:"preRun"`
	PostRun  map[string]HttpNodeConfig `json:"postRun" yaml:"postRun"`
}

// Clone cones HttpNode
func (hn *HttpNode) Clone() *HttpNode {
	cloned := &HttpNode{
		HttpNodeConfig: *hn.HttpNodeConfig.Clone(),
		Insecure:       hn.Insecure,
		PreRun:         make(map[string]HttpNodeConfig),
		PostRun:        make(map[string]HttpNodeConfig),
	}

	// update pre runs
	for k, v := range hn.PreRun {
		cloned.PreRun[k] = *v.Clone()
	}

	// update post runs
	for k, v := range hn.PostRun {
		cloned.PostRun[k] = *v.Clone()
	}

	return cloned
}

// Http node config
type HttpNodeConfig struct {
	URL                 string                 `json:"url" yaml:"url"`
	Method              string                 `json:"method" yaml:"method"`
	Headers             map[string]string      `json:"headers" yaml:"headers"`
	QueryParameters     map[string]interface{} `json:"queryParameters" yaml:"queryParameters"`
	BodyLanguage        string                 `json:"bodyLanguage" yaml:"bodyLanguage"`
	Body                interface{}            `json:"body" yaml:"body"`
	ResponseCode        int                    `json:"responseCode" yaml:"responseCode"`
	IncludeGlobalConfig bool                   `json:"includeGlobalConfig" yaml:"includeGlobalConfig"`
	Script              string                 `json:"script" yaml:"script"`
}

// Clone cones HttpNodeConfig
func (hn *HttpNodeConfig) Clone() *HttpNodeConfig {
	cloned := &HttpNodeConfig{
		URL:                 hn.URL,
		Method:              hn.Method,
		Headers:             make(map[string]string),
		QueryParameters:     make(map[string]interface{}),
		BodyLanguage:        hn.BodyLanguage,
		Body:                hn.Body,
		ResponseCode:        hn.ResponseCode,
		IncludeGlobalConfig: hn.IncludeGlobalConfig,
		Script:              hn.Script,
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
