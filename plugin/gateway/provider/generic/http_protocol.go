package generic

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	jsUtils "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

const (
	defaultHttpRequestTimeout = time.Second * 30
)

func (p *Provider) postHTTP(msg *msgTY.Message) error {
	cfgRaw, ok := p.Config.Nodes[msg.NodeID]
	if !ok {
		return fmt.Errorf("node not defined, nodeID:%s", msg.NodeID)
	}

	endpoint, err := toHttpNode(cfgRaw)
	if err != nil {
		zap.L().Error("error on converting to http endpoint config", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		return err
	}

	headers := endpoint.Headers
	queryParameters := endpoint.QueryParameters

	// merge with global config, if enabled
	if endpoint.IncludeGlobal {
		headers = p.HttpProtocol.Headers
		utils.JoinStringMap(headers, endpoint.Headers)

		queryParameters = p.HttpProtocol.QueryParameters
		utils.JoinMap(queryParameters, endpoint.QueryParameters)
	}

	var body interface{}
	// execute script, if available
	if endpoint.Script != "" {
		variables := make(map[string]interface{})
		variables["headers"] = headers
		variables["queryParameters"] = queryParameters
		variables["message"] = *msg
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
		body = utils.GetMapValue(mapResponse, "body", nil)
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

func (p *Provider) executeHttpRequest(cfg *HttpConfig) (*msgTY.RawMessage, error) {
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
		Others:       cmap.CustomMap{"address": cfg.URL},
	}

	variables := make(map[string]interface{})
	variables["response"] = res
	variables["responseBytes"] = resBytes

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
		messages := utils.GetMapValue(mapResponse, KeyReceivedMessages, nil)
		if messages == nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(messages)
		if err != nil {
			return nil, err
		}
		rawMessage.Data = jsonBytes
	}

	return rawMessage, nil
}
