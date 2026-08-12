package http_handler

import (
	"io"
	"net/http"
	"strings"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// FindOne func
func FindOne(storage storageTY.Plugin, w http.ResponseWriter, r *http.Request, entityName string, entity interface{}) {
	w.Header().Set("Content-Type", "application/json")

	filters, _, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = storage.FindOne(entityName, entity, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	od, err := json.Marshal(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteResponse(w, od)
}

// LoadData loads data
func LoadData(w http.ResponseWriter, r *http.Request, entityFn func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error)) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// RBAC: inject resource-scope filters into the storage query
	kind := kindFromRequest(r)
	f, err = applyListQueryScope(r, kind, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	result, err := entityFn(f, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteResponse(w, od)
}

// UpdateData loads data
func UpdateData(w http.ResponseWriter, r *http.Request, entity interface{}, updateFn func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error)) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	d, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(d, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := updateFn(f, p, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteResponse(w, od)
}

// FindMany func
func FindMany(storage storageTY.Plugin, w http.ResponseWriter, r *http.Request, entityName string, entities interface{}) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// RBAC: inject resource-scope filters into the storage query
	kind := EntityNameToKind(entityName)
	f, err = applyListQueryScope(r, kind, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	result, err := storage.Find(entityName, entities, f, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteResponse(w, od)
}

// EntityNameToKind maps a storage entity name to a policy kind.
func EntityNameToKind(entityName string) string {
	return policyTY.NormalizeKind(entityName)
}

// kindFromRequest derives the policy kind from the first api path segment.
func kindFromRequest(r *http.Request) string {
	const prefix = "/api/"
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	segment := strings.TrimPrefix(path, prefix)
	if index := strings.IndexByte(segment, '/'); index >= 0 {
		segment = segment[:index]
	}
	return policyTY.NormalizeKind(segment)
}

// SaveEntity func
func SaveEntity(storage storageTY.Plugin, w http.ResponseWriter, r *http.Request, entityName string, entity interface{}, bwFunc func(entity interface{}, filters *[]storageTY.Filter) error) {
	w.Header().Set("Content-Type", "application/json")

	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filters := make([]storageTY.Filter, 0)
	if bwFunc != nil {
		err = bwFunc(entity, &filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = storage.Upsert(entityName, entity, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// LoadEntity func
func LoadEntity(w http.ResponseWriter, r *http.Request, entity interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	d, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &entity)
	if err != nil {
		return err
	}
	return nil
}
