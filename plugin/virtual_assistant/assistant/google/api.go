package google_assistant

import (
	"io"
	"net/http"

	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/json"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google/types"
	"go.uber.org/zap"
)

// Google Assistant support is in progress
// Needs to complete VirtualDevice struct and implementation to support this feature
// for now this is incomplete and not usable

const (
	PluginGoogleAssistant = "google_assistant"
)

type Assistant struct {
	cfg *vaTY.Config
}

func New(cfg *vaTY.Config) (vaTY.Plugin, error) {
	return &Assistant{cfg: cfg}, nil
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
	// zap.L().Info("a request from", zap.Any("RequestURI", r.RequestURI), zap.Any("method", r.Method), zap.Any("headers", r.Header), zap.Any("query", r.URL.RawQuery))
	d, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		zap.L().Error("error on getting body", zap.Error(err))
		http.Error(w, "error on getting body", 500)
		return
	}
	// zap.L().Info("received a request from google assistant", zap.Any("body", string(d)))

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
			response = runQueryRequest(queryRequest)

		case gaTY.IntentExecute:
			executeRequest := gaTY.ExecuteRequest{}
			err = json.Unmarshal(d, &executeRequest)
			if err != nil {
				http.Error(w, "error on parsing", 500)
				return
			}
			response = runExecuteRequest(executeRequest)

		case gaTY.IntentSync:
			response = runSyncRequest(request)

		case gaTY.IntentDisconnect:
			runDisconnectRequest(request)

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
		// zap.L().Info("response", zap.Any("responseBytes", string(responseBytes)))

		handlerUtils.WriteResponse(w, responseBytes)
	} else {
		handlerUtils.PostSuccessResponse(w, nil)
	}
}
