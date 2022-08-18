package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	vdAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// registers virtual device api
func RegisterVirtualDeviceRoutes(router *mux.Router) {
	router.HandleFunc("/api/virtualdevice", listVirtualDevices).Methods(http.MethodGet)
	router.HandleFunc("/api/virtualdevice/{id}", getVirtualDevice).Methods(http.MethodGet)
	router.HandleFunc("/api/virtualdevice", updateVirtualDevice).Methods(http.MethodPost)
	router.HandleFunc("/api/virtualdevice", deleteVirtualDevices).Methods(http.MethodDelete)
}

func listVirtualDevices(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityVirtualDevice, &[]vdTY.VirtualDevice{})
}

func getVirtualDevice(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityVirtualDevice, &vdTY.VirtualDevice{})
}

func updateVirtualDevice(w http.ResponseWriter, r *http.Request) {
	entity := &vdTY.VirtualDevice{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// update modified on
	entity.ModifiedOn = time.Now()

	err = vdAPI.Save(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteVirtualDevices(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := vdAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}
