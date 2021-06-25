package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	backupRestoreAPI "github.com/mycontroller-org/server/v2/pkg/api/backup"
	json "github.com/mycontroller-org/server/v2/pkg/json"
	backupML "github.com/mycontroller-org/server/v2/pkg/model/backup"
	stgML "github.com/mycontroller-org/server/v2/plugin/storage"
)

// RegisterBackupRestoreRoutes registers backup/restore api
func RegisterBackupRestoreRoutes(router *mux.Router) {
	router.HandleFunc("/api/backup", listBackupFiles).Methods(http.MethodGet)
	router.HandleFunc("/api/backup", deleteBackupFile).Methods(http.MethodDelete)
	router.HandleFunc("/api/backup/run", runBackup).Methods(http.MethodPost)
	router.HandleFunc("/api/restore/run", runRestore).Methods(http.MethodGet)
}

func listBackupFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := handlerUtils.Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := backupRestoreAPI.List(f, p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func deleteBackupFile(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := backupRestoreAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func runRestore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")

	filter := stgML.Filter{Key: "id", Operator: stgML.OperatorEqual, Value: id}

	result, err := backupRestoreAPI.List([]stgML.Filter{filter}, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data, ok := result.Data.([]interface{})
	if !ok {
		http.Error(w, "error on converting slice interface", 500)
		return
	}

	if len(data) != 1 {
		http.Error(w, "no files or more than on entry found", 400)
		return
	}

	file, ok := data[0].(backupML.BackupFile)
	if !ok {
		http.Error(w, "error to convert to ExportedFile", 500)
		return
	}

	err = backupRestoreAPI.RunRestore(file)
	if len(data) != 1 {
		http.Error(w, err.Error(), 500)
		return
	}

	handlerUtils.WriteResponse(w, []byte("ok"))
}

func runBackup(w http.ResponseWriter, r *http.Request) {
	entity := &backupML.OnDemandBackupConfig{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.TargetLocation == "" {
		http.Error(w, "targetLocation should not be empty", 400)
		return
	}

	if entity.Handler == "" {
		http.Error(w, "handler should not be empty", 400)
		return
	}

	err = backupRestoreAPI.RunOnDemandBackup(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
