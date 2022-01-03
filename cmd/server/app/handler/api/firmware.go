package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	fwAPI "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	fwTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// RegisterFirmwareRoutes registers firmware api
func RegisterFirmwareRoutes(router *mux.Router) {
	router.HandleFunc("/api/firmware", listFirmwares).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware/{id}", getFirmware).Methods(http.MethodGet)
	router.HandleFunc("/api/firmware", updateFirmware).Methods(http.MethodPost)
	router.HandleFunc("/api/firmware", deleteFirmware).Methods(http.MethodDelete)
	router.HandleFunc("/api/firmware/upload/{id}", uploadFirmware).Methods(http.MethodPost)
}

func listFirmwares(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityFirmware, &[]fwTY.Firmware{})
}

func getFirmware(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityFirmware, &fwTY.Firmware{})
}

func updateFirmware(w http.ResponseWriter, r *http.Request) {
	entity := &fwTY.Firmware{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", 400)
		return
	}
	err = fwAPI.Save(entity, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteFirmware(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := fwAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
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
