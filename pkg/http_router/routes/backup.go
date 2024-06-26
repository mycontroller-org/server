package routes

import (
	"errors"
	"fmt"
	"net/http"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterBackupRestoreRoutes registers backup/restore api
func (h *Routes) registerBackupRestoreRoutes() {
	h.router.HandleFunc("/api/backup", h.listBackupFiles).Methods(http.MethodGet)
	h.router.HandleFunc("/api/backup", h.deleteBackupFile).Methods(http.MethodDelete)
	h.router.HandleFunc("/api/backup/run", h.runBackup).Methods(http.MethodPost)
	h.router.HandleFunc("/api/restore/run", h.runRestore).Methods(http.MethodGet)
}

func (h *Routes) listBackupFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := handlerUtils.Params(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.backupAPI.List(f, p)
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

func (h *Routes) deleteBackupFile(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := h.backupAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) runRestore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")

	filter := storageTY.Filter{Key: "id", Operator: storageTY.OperatorEqual, Value: id}

	result, err := h.backupAPI.List([]storageTY.Filter{filter}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, ok := result.Data.([]interface{})
	if !ok {
		http.Error(w, "error on converting slice interface", http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		http.Error(w, "file not found", http.StatusBadRequest)
		return
	}

	if len(data) > 1 {
		http.Error(w, "more than on entries found", http.StatusBadRequest)
		return
	}

	file, ok := data[0].(backupTY.BackupFile)
	if !ok {
		http.Error(w, "error on converting to backupFile struct", http.StatusInternalServerError)
		return
	}

	err = h.backupAPI.RunRestore(file)
	if len(data) != 1 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerUtils.WriteResponse(w, []byte("ok"))
}

func (h *Routes) runBackup(w http.ResponseWriter, r *http.Request) {
	entity := &backupTY.OnDemandBackupConfig{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.TargetLocation == "" {
		http.Error(w, "targetLocation should not be empty", http.StatusBadRequest)
		return
	}

	if entity.Handler == "" {
		http.Error(w, "handler should not be empty", http.StatusBadRequest)
		return
	}

	err = h.backupAPI.RunOnDemandBackup(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
