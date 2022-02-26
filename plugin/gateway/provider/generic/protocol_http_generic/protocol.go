package http_generic

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
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

	// merge with global config, if enabled
	if endpoint.IncludeGlobalConfig {
		headers := hp.Config.Headers
		utils.JoinStringMap(headers, endpoint.Headers)
		endpoint.Headers = headers

		queryParameters := hp.Config.QueryParameters
		utils.JoinMap(queryParameters, endpoint.QueryParameters)
		endpoint.QueryParameters = queryParameters
	}

	var body interface{}
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
		body = utils.GetMapValue(mapResponse, ScriptKeyDataOut, nil)
	} else {
		body = msg
	}

	client := httpclient.GetClient(endpoint.Insecure, defaultHttpRequestTimeout)
	_, _, err = client.Request(endpoint.URL, endpoint.Method, endpoint.Headers, endpoint.QueryParameters, body, endpoint.ResponseCode)
	if err != nil {
		zap.L().Error("error on calling endpoint", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
	}
	return err
}

// executes the given request and post back the rawmessage to the queue
func (hp *HttpProtocol) executeHttpRequest(cfg *HttpConfig) (*msgTY.RawMessage, error) {
	if cfg.Disabled {
		return nil, nil
	}

	client := httpclient.GetClient(cfg.Insecure, defaultHttpRequestTimeout)
	res, resBytes, err := client.Request(cfg.URL, cfg.Method, cfg.Headers, cfg.QueryParameters, cfg.Body, cfg.ResponseCode)
	if err != nil {
		return nil, err
	}

	rawMessage := &msgTY.RawMessage{
		IsReceived:   true,
		IsAckEnabled: false,
		Timestamp:    time.Now(),
		Data:         string(resBytes),
		Others:       cmap.CustomMap{"url": cfg.URL},
	}

	variables := map[string]interface{}{
		ScriptKeyConfigIn:   cfg,
		ScriptKeyResponseIn: res,
		ScriptKeyDataIn:     string(resBytes),
	}

	if cfg.Script != "" {
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
