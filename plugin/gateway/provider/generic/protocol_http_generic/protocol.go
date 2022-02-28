package http_generic

import (
	"fmt"
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
		err = hp.schedule(key, &cfg, hpCfg.Headers, hpCfg.QueryParameters)
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
			return fmt.Errorf("node not defined, nodeID:%s", msg.NodeID)
		}
		cfgRaw = defaultCfg
	}

	endpoint, err := toHttpNode(cfgRaw)
	if err != nil {
		zap.L().Error("error on converting to http endpoint config", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		return err
	}

	endpoint = endpoint.Clone()

	headers := endpoint.Headers
	queryParameters := endpoint.QueryParameters

	// merge with global config, if enabled
	if endpoint.IncludeGlobalConfig {
		headers, queryParameters = mergeHeadersQueryParameters(hp.Config.Headers, endpoint.Headers, hp.Config.QueryParameters, endpoint.QueryParameters)
	}

	body := ""
	// execute script, if available
	if endpoint.Script != "" {
		variables := map[string]interface{}{
			ScriptKeyConfigIn: *endpoint,
			ScriptKeyDataIn:   *msg,
		}

		scriptResponse, err := jsUtils.Execute(endpoint.Script, variables)
		if err != nil {
			zap.L().Error("error on executing script", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
		mapResponse, err := jsUtils.ToMap(scriptResponse)
		if err != nil {
			zap.L().Error("error on converting to map", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
		body = convertor.ToString(utils.GetMapValue(mapResponse, ScriptKeyDataOut, nil))
	} else {
		jsonString, err := json.MarshalToString(msg)
		if err != nil {
			zap.L().Error("error on converting to json string", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
		body = jsonString
	}

	client := httpclient.GetClient(endpoint.Insecure, defaultHttpRequestTimeout)
	_, err = client.Execute(endpoint.URL, endpoint.Method, headers, queryParameters, body, endpoint.ResponseCode)
	if err != nil {
		zap.L().Error("error on calling endpoint", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
	}
	return err
}

// executes the given request and post back the rawmessage to the queue
func (hp *HttpProtocol) executeHttpRequest(cfg *HttpConfig, globalHeaders map[string]string,
	globalQueryParameters map[string]interface{}) (*msgTY.RawMessage, error) {
	if cfg.Disabled {
		return nil, nil
	}

	client := httpclient.GetClient(cfg.Insecure, defaultHttpRequestTimeout)

	// execute pre run endpoints
	preRuns := make(map[string]httpclient.ResponseConfig)
	if len(cfg.PreRun) > 0 {
		preRunsResult, err := hp.executeSupportRuns(client, cfg.PreRun, cfg.IncludeGlobalConfig, globalHeaders, globalQueryParameters)
		if err != nil {
			zap.L().Error("error on pre run execution", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("url", cfg.URL), zap.Error(err))
			return nil, err
		}
		preRuns = preRunsResult
	}

	// execute actual endpoint
	// convert the body to json
	bodyString := getBodyString(cfg.Body, hp.GatewayConfig.ID, cfg.URL)
	response, err := client.Execute(cfg.URL, cfg.Method, cfg.Headers, cfg.QueryParameters, bodyString, cfg.ResponseCode)
	if err != nil {
		return nil, err
	}

	// execute post run, if any
	if len(cfg.PostRun) > 0 {
		_, err := hp.executeSupportRuns(client, cfg.PostRun, cfg.IncludeGlobalConfig, globalHeaders, globalQueryParameters)
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
			ScriptKeyConfigIn:     cfg,
			ScriptKeyDataIn:       response,
			ScriptKeyPreRunResult: preRuns,
		}
		scriptResponse, err := jsUtils.Execute(cfg.Script, variables)
		if err != nil {
			zap.L().Error("error on executing script", zap.String("address", cfg.URL), zap.Error(err))
			return nil, err
		}
		mapResponse, err := jsUtils.ToMap(scriptResponse)
		if err != nil {
			zap.L().Error("error on converting to map", zap.String("address", cfg.URL), zap.Error(err))
			return nil, err
		}
		messages := utils.GetMapValue(mapResponse, ScriptKeyDataOut, nil)
		if messages == nil {
			return nil, err
		}
		rawMessage.Data = messages
	}

	return rawMessage, nil
}

// execute pre runs and post runs
func (hp *HttpProtocol) executeSupportRuns(client *httpclient.Client, runs map[string]HttpNodeConfig, includeGlobalConfig bool, globalHeaders map[string]string, globalQueryParameters map[string]interface{}) (map[string]httpclient.ResponseConfig, error) {
	result := make(map[string]httpclient.ResponseConfig)
	for name, cfg := range runs {
		headers := cfg.Headers
		queryParameters := cfg.QueryParameters
		if includeGlobalConfig {
			headers, queryParameters = mergeHeadersQueryParameters(globalHeaders, cfg.Headers, globalQueryParameters, cfg.QueryParameters)
		}
		bodyString := getBodyString(cfg.Body, hp.GatewayConfig.ID, cfg.URL)
		if cfg.Script != "" {
			variables := map[string]interface{}{
				ScriptKeyPreRunResult: result,
			}
			response, err := executeScript(cfg.Script, variables, ScriptKeyDataOut)
			if err != nil {
				zap.L().Error("error on executing a support run script", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("name", name), zap.String("url", cfg.URL), zap.Error(err))
				return nil, err
			}

			bodyString = getBodyString(response, hp.GatewayConfig.ID, cfg.URL)
		}

		response, err := client.Execute(cfg.URL, cfg.Method, headers, queryParameters, bodyString, cfg.ResponseCode)
		if err != nil {
			zap.L().Error("error on executing a support run", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("name", name), zap.String("url", cfg.URL), zap.Error(err))
			return nil, err
		}
		result[name] = *response
	}
	return result, nil
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
func getBodyString(body interface{}, gatewayID, url string) string {
	bodyString, ok := body.(string)
	if ok {
		return bodyString
	}
	bodyString, err := json.MarshalToString(body)
	if err != nil {
		zap.L().Debug("error converting the body to json string, fall back to string conversion", zap.String("gatewayId", gatewayID), zap.String("url", url), zap.Error(err))
		bodyString = convertor.ToString(body)
	}
	return bodyString
}

// execute script
func executeScript(script string, variables map[string]interface{}, responseKey string) (interface{}, error) {
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
	response := utils.GetMapValue(mapResponse, responseKey, nil)
	return response, nil
}
