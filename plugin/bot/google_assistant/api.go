package googleassistant

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/model"
	"go.uber.org/zap"
)

// Google Assistant support is in progress
// Needs to complete VirtualDevice model and implementation to support this feature
// for now this is incomplete and not usable

func RegisterGoogleAssistantRoutes(router *mux.Router) {
	router.HandleFunc("/api/plugin/bot/google_assistant", processRequest)
}

func processRequest(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("a request from", zap.Any("RequestURI", r.RequestURI), zap.Any("method", r.Method), zap.Any("headers", r.Header), zap.Any("query", r.URL.RawQuery))
	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		zap.L().Error("error on getting body", zap.Error(err))
		http.Error(w, "error on getting body", 500)
		return
	}
	zap.L().Info("received a request from google assistant", zap.Any("body", string(d)))

	request := model.Request{}
	err = json.Unmarshal(d, &request)
	if err != nil {
		http.Error(w, "error on parsing", 500)
		return
	}

	var response interface{}

	for _, input := range request.Inputs {
		switch input.Intent {
		case model.IntentQuery:
			queryRequest := model.QueryRequest{}
			err = json.Unmarshal(d, &queryRequest)
			if err != nil {
				http.Error(w, "error on parsing", 500)
				return
			}
			response = runQueryRequest(queryRequest)

		case model.IntentExecute:
			executeRequest := model.ExecuteRequest{}
			err = json.Unmarshal(d, &executeRequest)
			if err != nil {
				http.Error(w, "error on parsing", 500)
				return
			}
			response = runExecuteRequest(executeRequest)

		case model.IntentSync:
			response = runSyncRequest(request)

		case model.IntentDisconnect:
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
		handlerUtils.WriteResponse(w, responseBytes)
	} else {
		handlerUtils.PostSuccessResponse(w, nil)
	}
}
