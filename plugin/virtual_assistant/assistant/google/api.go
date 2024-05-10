package google_assistant

import (
	"context"
	"io"
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google/types"
	deviceAPI "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/device_api"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	"go.uber.org/zap"
)

// Google Assistant support is in progress
// Needs to complete VirtualDevice struct and implementation to support this feature
// for now this is incomplete and not usable

const (
	PluginGoogleAssistant = "google_assistant"
	loggerName            = "virtual_assistant_google"
)

type Assistant struct {
	ctx       context.Context
	logger    *zap.Logger
	cfg       *vaTY.Config
	deviceAPI *deviceAPI.DeviceAPI
}

func New(ctx context.Context, cfg *vaTY.Config) (vaTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	_deviceAPI, err := deviceAPI.New(ctx)
	if err != nil {
		return nil, err
	}
	return &Assistant{
		ctx:       ctx,
		logger:    logger.Named(loggerName),
		cfg:       cfg,
		deviceAPI: _deviceAPI,
	}, nil
}

func (a *Assistant) Name() string {
	return PluginGoogleAssistant
}

func (a *Assistant) Start() error {
	return nil
}

func (a *Assistant) Stop() error {
	return nil
}

func (a *Assistant) Config() *vaTY.Config {
	return a.cfg
}

func (a *Assistant) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// a.logger.Info("a request from", zap.Any("RequestURI", r.RequestURI), zap.Any("method", r.Method), zap.Any("headers", r.Header), zap.Any("query", r.URL.RawQuery))
	d, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.logger.Error("error on getting body", zap.Error(err))
		http.Error(w, "error on getting body", 500)
		return
	}
	// a.logger.Info("received a request from google assistant", zap.Any("body", string(d)))

	request := gaTY.Request{}
	err = json.Unmarshal(d, &request)
	if err != nil {
		http.Error(w, "error on parsing", 500)
		return
	}

	var response interface{}

	for _, input := range request.Inputs {
		switch input.Intent {
		case gaTY.IntentQuery:
			queryRequest := gaTY.QueryRequest{}
			err = json.Unmarshal(d, &queryRequest)
			if err != nil {
				http.Error(w, "error on parsing", 500)
				return
			}
			response = a.runQueryRequest(queryRequest)

		case gaTY.IntentExecute:
			executeRequest := gaTY.ExecuteRequest{}
			err = json.Unmarshal(d, &executeRequest)
			if err != nil {
				http.Error(w, "error on parsing", 500)
				return
			}
			response = a.runExecuteRequest(executeRequest)

		case gaTY.IntentSync:
			response = a.runSyncRequest(request)

		case gaTY.IntentDisconnect:
			a.runDisconnectRequest(request)

		default:
			// noop
		}
	}

	if response != nil {
		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// a.logger.Info("response", zap.Any("responseBytes", string(responseBytes)))

		handlerUtils.WriteResponse(w, responseBytes)
	} else {
		handlerUtils.PostSuccessResponse(w, nil)
	}
}
