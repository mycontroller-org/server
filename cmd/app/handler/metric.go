package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
)

func registerMetricRoutes(router *mux.Router) {
	router.HandleFunc("/api/metric", getMetric).Methods(http.MethodPost)
}

func getMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	/*
		f, p, err := Params(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	*/
	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	queryConfig := &mtrml.QueryConfig{}

	err = json.Unmarshal(d, queryConfig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := svc.MTS.Query(queryConfig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(od)
}
