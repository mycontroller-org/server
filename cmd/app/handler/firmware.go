package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fwml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerFirmwareRoutes(router *mux.Router) {
	router.HandleFunc("/api/firmware", listFirmwares).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware/{id}", getFirmware).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware", updateFirmware).Methods(http.MethodPost)
}

func listFirmwares(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityFirmware, &[]fwml.Firmware{})
}

func getFirmware(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityFirmware, &fwml.Firmware{})
}

func updateFirmware(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		e := d.(*fwml.Firmware)
		if e.ID == "" {
			e.ID = ut.RandID()
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityFirmware, &fwml.Firmware{}, bwFunc)
}
