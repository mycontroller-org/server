package http_generic

import (
	"fmt"
	"sort"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	jsUtils "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	defaultHttpRequestTimeout = time.Second * 30
)

// New returns new instance of generic http protocol
func New(gwCfg *gwTY.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgTY.RawMessage) error) (*HttpProtocol, error) {
	hpCfg := &HttpProtocolConf{}
	err := json.ToStruct(protocol, hpCfg)
	if err != nil {
		zap.L().Error("error on converting to http protocol")
		return nil, err
	}

	hp := &HttpProtocol{
		GatewayConfig:     gwCfg,
		Config:            hpCfg,
		rawMessageHandler: rxMsgFunc,
	}

	if len(hpCfg.Endpoints) == 0 {
		return hp, nil
	}
	for key := range hpCfg.Endpoints {
		cfg := hpCfg.Endpoints[key]
		err = hp.schedule(key, &cfg)
		if err != nil {
			zap.L().Error("error on schedule", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", cfg.URL), zap.Error(err))
		}
	}

	return hp, nil
}

// Close closes the generic http protocol
func (hp *HttpProtocol) Close() error {
	hp.unscheduleAll()
	return nil
}

// Post the received message to the specified target url
// if none matched uses "default" named endpoint
func (hp *HttpProtocol) Post(msg *msgTY.Message) error {
	cfgRaw, ok := hp.Config.Nodes[msg.NodeID]
	if !ok {
		defaultCfg, ok := hp.Config.Nodes[DefaultNode]
		if !ok {
			return fmt.Errorf("node not defined, gatewayId:%s, nodeId:%s", msg.GatewayID, msg.NodeID)
		}
		cfgRaw = defaultCfg
	}

	endpoint, err := toHttpNode(cfgRaw)
	if err != nil {
		zap.L().Error("error on converting to http endpoint config", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		return err
	}

	endpoint = endpoint.Clone()

	client := httpclient.GetClient(endpoint.Insecure, defaultHttpRequestTimeout)

	// execute pre run
	preRunResponse := make(map[string]*httpclient.ResponseConfig)
	if len(endpoint.PreRun) > 0 {
		preRunRes, err := hp.executeSupportRuns(client, PreRun, nil, endpoint.PreRun, endpoint.IncludeGlobalConfig)
		if err != nil {
			zap.L().Error("error on node endpoint pre run execution", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", endpoint.URL), zap.Error(err))
			return err
		}
		preRunResponse = preRunRes
	}

	// post run, if any
	defer func() {
		if len(endpoint.PostRun) > 0 {
			_, err := hp.executeSupportRuns(client, PostRun, preRunResponse, endpoint.PostRun, endpoint.IncludeGlobalConfig)
			if err != nil {
				zap.L().Error("error on node endpoint post run execution", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", endpoint.URL), zap.Error(err))
			}
		}
	}()

	// execute actual run
	if endpoint.Script != "" {
		variables := map[string]interface{}{
			ScriptKeyConfigIn:       endpoint,
			ScriptKeyDataIn:         msg,
			ScriptKeyPreRunResponse: preRunResponse,
		}

		responseMap, err := executeScript(endpoint.Script, variables)
		if err != nil {
			zap.L().Error("error on executing node endpoint script", zap.String("address", endpoint.URL), zap.Error(err))
			return err
		}

		// get messages
		messages := utils.GetMapValue(responseMap, ScriptKeyDataOut, nil)
		if messages != nil {
			rawMessage := &msgTY.RawMessage{
				IsReceived:   true,
				IsAckEnabled: false,
				Timestamp:    time.Now(),
				Data:         messages,
			}
			err = hp.rawMessageHandler(rawMessage)
			if err != nil {
				zap.L().Error("error on posting raw messages into queue", zap.String("gatewayId", hp.GatewayConfig.ID), zap.Any("request", msg))
				return err
			}
		}

		// if asked to not execute, return from here
		execute := convertor.ToBool(utils.GetMapValue(responseMap, ScriptKeyExecute, false))
		if !execute {
			return nil
		}
	}

	// execute
	// convert the body to string
	bodyString := hp.getBodyString(endpoint.Body, endpoint.BodyLanguage, endpoint.URL)
	_, err = client.Execute(endpoint.URL, endpoint.Method, endpoint.Headers, endpoint.QueryParameters, bodyString, endpoint.ResponseCode)
	if err != nil {
		zap.L().Error("error on calling node endpoint", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		return err
	}

	return nil
}

// executes the given request and post back the rawmessage to the queue
func (hp *HttpProtocol) executeHttpRequest(cfg *HttpConfig) (*msgTY.RawMessage, error) {
	if cfg.Disabled {
		return nil, nil
	}

	client := httpclient.GetClient(cfg.Insecure, defaultHttpRequestTimeout)

	// execute pre run endpoints
	preRunResponse := make(map[string]*httpclient.ResponseConfig)
	if len(cfg.PreRun) > 0 {
		preRunRes, err := hp.executeSupportRuns(client, PreRun, nil, cfg.PreRun, cfg.IncludeGlobalConfig)
		if err != nil {
			zap.L().Error("error on pre run execution", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", cfg.URL), zap.Error(err))
			return nil, err
		}
		preRunResponse = preRunRes
	}

	// execute actual run
	// convert the body to json
	bodyString := hp.getBodyString(cfg.Body, cfg.BodyLanguage, cfg.URL)
	response, err := client.Execute(cfg.URL, cfg.Method, cfg.Headers, cfg.QueryParameters, bodyString, cfg.ResponseCode)
	if err != nil {
		return nil, err
	}

	// execute post run, if any
	if len(cfg.PostRun) > 0 {
		_, err := hp.executeSupportRuns(client, PostRun, preRunResponse, cfg.PostRun, cfg.IncludeGlobalConfig)
		if err != nil {
			zap.L().Error("error on post run execution", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", cfg.URL), zap.Error(err))
		}
	}

	rawMessage := &msgTY.RawMessage{
		IsReceived:   true,
		IsAckEnabled: false,
		Timestamp:    time.Now(),
		Data:         response.StringBody(),
		Others:       cmap.CustomMap{"url": cfg.URL},
	}

	if cfg.Script != "" {
		variables := map[string]interface{}{
			ScriptKeyConfigIn:       cfg,
			ScriptKeyDataIn:         response,
			ScriptKeyPreRunResponse: preRunResponse,
		}

		responseMap, err := executeScript(cfg.Script, variables)
		if err != nil {
			zap.L().Error("error on executing script", zap.String("address", cfg.URL), zap.Error(err))
			return nil, err
		}
		messages := utils.GetMapValue(responseMap, ScriptKeyDataOut, nil)
		if messages == nil {
			return nil, fmt.Errorf("'%s' can not be empty", ScriptKeyDataOut)
		}
		rawMessage.Data = messages
	}

	return rawMessage, nil
}

// execute pre runs and post runs
func (hp *HttpProtocol) executeSupportRuns(client *httpclient.Client, runType string, preRunResponse map[string]*httpclient.ResponseConfig, runs map[string]HttpNodeConfig, includeGlobalConfig bool) (map[string]*httpclient.ResponseConfig, error) {
	runResponses := make(map[string]*httpclient.ResponseConfig)

	// get runKeys and order
	runKeys := []string{}
	for key := range runs {
		runKeys = append(runKeys, key)
	}
	sort.Strings(runKeys)

	for _, name := range runKeys {
		cfg := runs[name]
		headers := cfg.Headers
		queryParameters := cfg.QueryParameters
		if includeGlobalConfig {
			headers, queryParameters = mergeHeadersQueryParameters(hp.Config.Headers, cfg.Headers, hp.Config.QueryParameters, cfg.QueryParameters)
		}
		bodyString := hp.getBodyString(cfg.Body, cfg.BodyLanguage, cfg.URL)
		if cfg.Script != "" {
			// update variables
			variables := map[string]interface{}{}
			if runType == PreRun {
				variables[ScriptKeyPreRunResponse] = runResponses
			} else if runType == PostRun {
				variables[ScriptKeyPreRunResponse] = preRunResponse
				variables[ScriptKeyPostRunResponse] = runResponses
			}

			responseMap, err := executeScript(cfg.Script, variables)
			if err != nil {
				zap.L().Error("error on executing a support run script", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("name", name), zap.String("url", cfg.URL), zap.Error(err))
				return nil, err
			}
			bodyString = convertor.ToString(utils.GetMapValue(responseMap, ScriptKeyDataOut, ""))
		}

		response, err := client.Execute(cfg.URL, cfg.Method, headers, queryParameters, bodyString, cfg.ResponseCode)
		if err != nil {
			zap.L().Error("error on executing a support run", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("name", name), zap.String("url", cfg.URL), zap.Error(err))
			return nil, err
		}
		runResponses[name] = response
	}
	return runResponses, nil
}

// merges headers and queryParameters
func mergeHeadersQueryParameters(headers1, headers2 map[string]string,
	queryParameters1, queryParameters2 map[string]interface{}) (map[string]string, map[string]interface{}) {

	// final headers and query parameters
	finalHeaders := make(map[string]string)
	finalQueryParameters := make(map[string]interface{})

	// merge headers
	utils.JoinStringMap(finalHeaders, headers1)
	utils.JoinStringMap(finalHeaders, headers2)

	// merge query parameters
	utils.JoinMap(finalQueryParameters, queryParameters1)
	utils.JoinMap(finalQueryParameters, queryParameters2)

	return finalHeaders, finalQueryParameters
}

// returns body string
func (hp *HttpProtocol) getBodyString(body interface{}, bodyLanguage, url string) string {
	switch bodyLanguage {
	case BodyLanguageJSON:
		bodyString, err := json.MarshalToString(body)
		if err != nil {
			zap.L().Debug("error converting the body to json string, fall back to string conversion", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", url), zap.Error(err))
			bodyString = convertor.ToString(body)
		}
		return bodyString

	case BodyLanguageYAML:
		bodyBytes, err := yaml.Marshal(body)
		if err != nil {
			zap.L().Debug("error converting the body to yaml string, fall back to string conversion", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", url), zap.Error(err))
			return convertor.ToString(body)
		}
		return string(bodyBytes)

	default:
		return convertor.ToString(body)

	}

}

// execute script
func executeScript(script string, variables map[string]interface{}) (map[string]interface{}, error) {
	scriptResponse, err := jsUtils.Execute(script, variables)
	if err != nil {
		zap.L().Error("error on executing script", zap.Error(err))
		return nil, err
	}
	mapResponse, err := jsUtils.ToMap(scriptResponse)
	if err != nil {
		zap.L().Error("error on converting to map", zap.Error(err))
		return nil, err
	}
	return mapResponse, nil
}
