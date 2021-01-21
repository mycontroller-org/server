package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fwml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerFirmwareRoutes(router *mux.Router) {
	router.HandleFunc("/api/firmware", listFirmwares).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware/{id}", getFirmware).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware", updateFirmware).Methods(http.MethodPost)
	router.HandleFunc("/api/firmware", deleteFirmware).Methods(http.MethodDelete)
	router.HandleFunc("/api/firmware/upload/{id}", uploadFirmware).Methods(http.MethodPost)
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

func deleteFirmware(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := fwAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &ids, updateFn)
}

func uploadFirmware(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, "id not supplied", 400)
		return
	}

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close() // Close the file when we finish

	err = fwAPI.Upload(file, id, handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
