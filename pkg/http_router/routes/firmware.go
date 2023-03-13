package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	fwTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterFirmwareRoutes registers firmware api
func (h *Routes) registerFirmwareRoutes() {
	h.router.HandleFunc("/api/firmware", h.listFirmwares).Methods(http.MethodGet)
	h.router.HandleFunc("/api/firmware/{id}", h.getFirmware).Methods(http.MethodGet)
	h.router.HandleFunc("/api/firmware", h.updateFirmware).Methods(http.MethodPost)
	h.router.HandleFunc("/api/firmware", h.deleteFirmware).Methods(http.MethodDelete)
	h.router.HandleFunc("/api/firmware/upload/{id}", h.uploadFirmware).Methods(http.MethodPost)
}

func (h *Routes) listFirmwares(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityFirmware, &[]fwTY.Firmware{})
}

func (h *Routes) getFirmware(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityFirmware, &fwTY.Firmware{})
}

func (h *Routes) updateFirmware(w http.ResponseWriter, r *http.Request) {
	entity := &fwTY.Firmware{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	err = h.api.Firmware().Save(entity, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteFirmware(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := h.api.Firmware().Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) uploadFirmware(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, "id not supplied", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close() // Close the file when we finish

	err = h.api.Firmware().Upload(file, id, handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
