package routes

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// registers virtual device api
func (h *Routes) registerVirtualDeviceRoutes() {
	h.router.HandleFunc("/api/virtualdevice", h.listVirtualDevices).Methods(http.MethodGet)
	h.router.HandleFunc("/api/virtualdevice/{id}", h.getVirtualDevice).Methods(http.MethodGet)
	h.router.HandleFunc("/api/virtualdevice", h.updateVirtualDevice).Methods(http.MethodPost)
	h.router.HandleFunc("/api/virtualdevice", h.deleteVirtualDevices).Methods(http.MethodDelete)
}

func (h *Routes) listVirtualDevices(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityVirtualDevice, &[]vdTY.VirtualDevice{})
}

func (h *Routes) getVirtualDevice(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityVirtualDevice, &vdTY.VirtualDevice{})
}

func (h *Routes) updateVirtualDevice(w http.ResponseWriter, r *http.Request) {
	entity := &vdTY.VirtualDevice{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update modified on
	entity.ModifiedOn = time.Now()

	err = h.api.VirtualDevice().Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteVirtualDevices(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.VirtualDevice().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}
