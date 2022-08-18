package alexa

// https://stackoverflow.com/questions/38230157/cannot-finding-my-unpublished-alexa-skills-kit

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/json"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/alexa/types"
	"go.uber.org/zap"
)

// Google Assistant support is in progress
// Needs to complete VirtualDevice struct and implementation to support this feature
// for now this is incomplete and not usable

func RegisterAlexaRoutes(router *mux.Router) {
	router.HandleFunc("/api/plugin/bot/alexa", processRequest)
}

func processRequest(w http.ResponseWriter, r *http.Request) {
	// zap.L().Info("a request from", zap.Any("RequestURI", r.RequestURI), zap.Any("method", r.Method), zap.Any("headers", r.Header), zap.Any("query", r.URL.RawQuery))
	d, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		zap.L().Error("error on getting body", zap.Error(err))
		http.Error(w, "error on getting body", 500)
		return
	}

	request := alexaTY.Request{}
	err = json.Unmarshal(d, &request)
	if err != nil {
		zap.L().Error("error on forming directive object", zap.Error(err))
		http.Error(w, "error on forming directive object", 500)
		return
	}

	var response interface{}

	if request.Directive.Header.Name == alexaTY.RequestReportState {
		_response, err := reportState(request.Directive)
		if err != nil {
			http.Error(w, "error on getting device state", 500)
			return
		}
		response = _response
	} else if request.Directive.Header.Namespace == alexaTY.NamespaceDiscovery {
		_response, err := executeDiscover(request.Directive)
		if err != nil {
			http.Error(w, "error on executing discover devices", 500)
			return
		}
		response = _response
	} else {
		response = executiveDirective(request.Directive)
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
