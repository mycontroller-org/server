package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// ReceivedQueryMap returns all the user query and url input
func ReceivedQueryMap(request *http.Request) (map[string][]string, error) {
	data := make(map[string][]string, 0)
	// url parameters
	filters := mux.Vars(request)
	for key, value := range filters {
		data[key] = []string{value}
	}

	// query perameters
	query := request.URL.Query()
	for key, value := range query {
		data[key] = value
	}

	return data, nil
}

// Params func
func Params(request *http.Request) ([]stgml.Filter, *stgml.Pagination, error) {
	f := mux.Vars(request)
	q := request.URL.Query()
	for key, value := range q {
		f[key] = value[0] // TODO: FIX this to fetch all the values
	}

	// get Pagination arguments
	// start with pagination default values
	p := stgml.Pagination{
		Limit:  50,
		Offset: 0,
		SortBy: []stgml.Sort{},
	}

	lFunc := func(key string) (int64, error) {
		if value, ok := f[key]; ok {
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return 0, err
			}
			return int64(intValue), nil
		}
		return 0, fmt.Errorf("Key '%s' not found in the map", key)
	}

	v, err := lFunc("limit")
	if err == nil {
		p.Limit = v
	}

	v, err = lFunc("offset")
	if err == nil {
		p.Offset = v
	}

	// fetch sort options
	if sr, ok := f["sortBy"]; ok {
		s := &[]stgml.Sort{}
		err := json.Unmarshal([]byte(sr), s)
		if err != nil {
			return nil, nil, err
		}
		p.SortBy = *s
	}
	// remove these keys from map
	delete(f, "limit")
	delete(f, "offset")
	delete(f, "sortBy")

	filters := make([]stgml.Filter, 0)

	for k, v := range f {
		if k != "filter" {
			filters = append(filters, stgml.Filter{
				Key:   k,
				Value: v,
			})
		}
	}

	if fj, ok := f["filter"]; ok {
		fs := &[]stgml.Filter{}
		err := json.Unmarshal([]byte(fj), fs)
		if err != nil {
			return nil, nil, err
		}
		filters = append(filters, *fs...)
	}

	zap.L().Debug("received filters and pagination", zap.Any("filter", filters), zap.Any("pagination", p))

	return filters, &p, nil
}

// FindOne func
func FindOne(w http.ResponseWriter, r *http.Request, en string, e interface{}) {
	w.Header().Set("Content-Type", "application/json")

	f, _, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = stg.SVC.FindOne(en, e, f)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	od, err := json.Marshal(e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(od)
}

// LoadData loads data
func LoadData(w http.ResponseWriter, r *http.Request, entityFn func(f []stgml.Filter, p *stgml.Pagination) (interface{}, error)) {
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
	w.Write(od)
}

// UpdateData loads data
func UpdateData(w http.ResponseWriter, r *http.Request, entity interface{}, updateFn func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error)) {
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
	w.Write(od)
}

// FindMany func
func FindMany(w http.ResponseWriter, r *http.Request, entityName string, entities interface{}) {
	w.Header().Set("Content-Type", "application/json")

	f, p, err := Params(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result, err := stg.SVC.Find(entityName, entities, f, p)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	od, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(od)
}

// SaveEntity func
func SaveEntity(w http.ResponseWriter, r *http.Request, entityName string, entity interface{}, bwFunc func(entity interface{}, filters *[]stgml.Filter) error) {
	w.Header().Set("Content-Type", "application/json")

	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	filters := make([]stgml.Filter, 0)
	if bwFunc != nil {
		err = bwFunc(entity, &filters)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	err = stg.SVC.Upsert(entityName, entity, filters)
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
