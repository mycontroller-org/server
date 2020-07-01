package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
)

func registerNodeRoutes(router *mux.Router) {
	router.HandleFunc("/api/node", listnodes).Methods(http.MethodGet)
	router.HandleFunc("/api/node/{id}", getnode).Methods(http.MethodGet)
	router.HandleFunc("/api/node", updatenode).Methods(http.MethodPost)
}

func listnodes(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntityNode, &[]ml.Node{})
}

func getnode(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntityNode, &ml.Node{})
}

func updatenode(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]ml.Filter) error {
		e := d.(*ml.Node)
		if e.ID == "" {
			if e.GatewayID == "" || e.ShortID == "" {
				return errors.New("GatewayID or ShortID field should not be empty")
			}
			e.ID = fmt.Sprintf("%s_%s", e.GatewayID, e.ShortID)
		}
		return nil
	}
	saveEntity(w, r, ml.EntityNode, &ml.Node{}, bwFunc)
}
