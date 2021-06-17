package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/backend/v2/pkg/api/field"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	mtsML "github.com/mycontroller-org/backend/v2/plugin/metrics"
)

// global constants
const (
	QuickID = "quick_id"
)

// RegisterMetricRoutes registers metric api
func RegisterMetricRoutes(router *mux.Router) {
	router.HandleFunc("/api/metric", getMetricList).Methods(http.MethodPost)
	router.HandleFunc("/api/metric", getMetric).Methods(http.MethodGet)
}

func getMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// returns values from parameters / user input
	getValue := func(key string) string {
		if values, ok := params[key]; ok {
			if len(values) > 0 {
				return values[0]
			}
		}
		return ""
	}
	// returns all the values
	getValues := func(key string) []string {
		if values, ok := params[key]; ok {
			if len(values) > 0 {
				return values
			}
		}
		return nil
	}

	queryConfig := &mtsML.QueryConfig{}
	queryConfig.Individual = []mtsML.Query{{Name: QuickID, Tags: map[string]string{}}}

	if quickID, ok := params[QuickID]; ok {
		if len(quickID) > 0 {
			rt, kvMap, err := quickIdUL.EntityKeyValueMap(quickID[0])
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			// get resource details
			switch {
			case utils.ContainsString(quickIdUL.QuickIDField, rt):
				// get field details
				field, err := field.GetByIDs(kvMap[model.KeyGatewayID], kvMap[model.KeyNodeID], kvMap[model.KeySourceID], kvMap[model.KeyFieldID])
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				queryConfig.Individual[0].Tags[model.KeyID] = field.ID
				queryConfig.Individual[0].MetricType = field.MetricType

			default:
				http.Error(w, fmt.Sprintf("resource type not supported in metric. ResourceType:%s", rt), 500)
				return
			}
		} else {
			http.Error(w, fmt.Sprintf("%s not supplied", QuickID), 500)
			return
		}
	} else {
		http.Error(w, fmt.Sprintf("%s not supplied", QuickID), 500)
		return
	}

	// update optional parameters
	if value := getValue(mtsML.QueryKeyStart); value != "" {
		queryConfig.Global.Start = value
	}
	if value := getValue(mtsML.QueryKeyStop); value != "" {
		queryConfig.Global.Stop = value
	}
	if value := getValue(mtsML.QueryKeyWindow); value != "" {
		queryConfig.Global.Window = value
	}
	if values := getValues(mtsML.QueryKeyFunctions); values != nil {
		queryConfig.Global.Functions = values
	}

	result, err := metrics.SVC.Query(queryConfig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func getMetricList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	queryConfig := &mtsML.QueryConfig{}

	err = json.Unmarshal(d, queryConfig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := metrics.SVC.Query(queryConfig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handlerUtils.WriteResponse(w, od)
}
