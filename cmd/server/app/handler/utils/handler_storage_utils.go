package handlerutils

import (
	"io/ioutil"
	"net/http"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/store"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// FindOne func
func FindOne(w http.ResponseWriter, r *http.Request, entityName string, entity interface{}) {
	w.Header().Set("Content-Type", "application/json")

	filters, _, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = store.STORAGE.FindOne(entityName, entity, filters)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	od, err := json.Marshal(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteResponse(w, od)
}

// LoadData loads data
func LoadData(w http.ResponseWriter, r *http.Request, entityFn func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error)) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := entityFn(f, p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteResponse(w, od)
}

// UpdateData loads data
func UpdateData(w http.ResponseWriter, r *http.Request, entity interface{}, updateFn func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error)) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = json.Unmarshal(d, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := updateFn(f, p, d)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteResponse(w, od)
}

// FindMany func
func FindMany(w http.ResponseWriter, r *http.Request, entityName string, entities interface{}) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := store.STORAGE.Find(entityName, entities, f, p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteResponse(w, od)
}

// SaveEntity func
func SaveEntity(w http.ResponseWriter, r *http.Request, entityName string, entity interface{}, bwFunc func(entity interface{}, filters *[]storageTY.Filter) error) {
	w.Header().Set("Content-Type", "application/json")

	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	filters := make([]storageTY.Filter, 0)
	if bwFunc != nil {
		err = bwFunc(entity, &filters)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	err = store.STORAGE.Upsert(entityName, entity, filters)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// LoadEntity func
func LoadEntity(w http.ResponseWriter, r *http.Request, entity interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &entity)
	if err != nil {
		return err
	}
	return nil
}
