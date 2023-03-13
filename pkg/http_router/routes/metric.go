package routes

import (
	"fmt"
	"io"
	"net/http"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	quickIdUL "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	mtsTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
)

// global constants
const (
	QuickID = "quick_id"
)

// RegisterMetricRoutes registers metric api
func (h *Routes) registerMetricRoutes() {
	h.router.HandleFunc("/api/metric", h.getMetricList).Methods(http.MethodPost)
	h.router.HandleFunc("/api/metric", h.getMetric).Methods(http.MethodGet)
}

func (h *Routes) getMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	queryConfig := &mtsTY.QueryConfig{}
	queryConfig.Individual = []mtsTY.Query{{Name: QuickID, Tags: map[string]string{}}}

	if quickID, ok := params[QuickID]; ok {
		if len(quickID) > 0 {
			rt, kvMap, err := quickIdUL.EntityKeyValueMap(quickID[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// get resource details
			switch rt {
			case quickIdUL.QuickIdField:
				// get field details
				field, err := h.api.Field().GetByIDs(kvMap[types.KeyGatewayID], kvMap[types.KeyNodeID], kvMap[types.KeySourceID], kvMap[types.KeyFieldID])
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				queryConfig.Individual[0].Tags[types.KeyID] = field.ID
				queryConfig.Individual[0].MetricType = field.MetricType

			default:
				http.Error(w, fmt.Sprintf("resource type not supported in metric. ResourceType:%s", rt), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, fmt.Sprintf("%s not supplied", QuickID), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, fmt.Sprintf("%s not supplied", QuickID), http.StatusInternalServerError)
		return
	}

	// update optional parameters
	if value := getValue(mtsTY.QueryKeyStart); value != "" {
		queryConfig.Global.Start = value
	}
	if value := getValue(mtsTY.QueryKeyStop); value != "" {
		queryConfig.Global.Stop = value
	}
	if value := getValue(mtsTY.QueryKeyWindow); value != "" {
		queryConfig.Global.Window = value
	}
	if values := getValues(mtsTY.QueryKeyFunctions); values != nil {
		queryConfig.Global.Functions = values
	}

	result, err := h.metric.Query(queryConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func (h *Routes) getMetricList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	d, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queryConfig := &mtsTY.QueryConfig{}

	err = json.Unmarshal(d, queryConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.metric.Query(queryConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerUtils.WriteResponse(w, od)
}
