package handlerutils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	json "github.com/mycontroller-org/server/v2/pkg/json"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// ReceivedQueryMap returns all the user query and url input
func ReceivedQueryMap(request *http.Request) (map[string][]string, error) {
	data := make(map[string][]string)
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
func Params(request *http.Request) ([]storageTY.Filter, *storageTY.Pagination, error) {
	f := mux.Vars(request)
	q := request.URL.Query()
	for key, value := range q {
		f[key] = value[0] // TODO: FIX this to fetch all the values
	}

	// get Pagination arguments
	// start with pagination default values
	p := storageTY.Pagination{
		Limit:  50,
		Offset: 0,
		SortBy: []storageTY.Sort{},
	}

	lFunc := func(key string) (int64, error) {
		if value, ok := f[key]; ok {
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return 0, err
			}
			return int64(intValue), nil
		}
		return 0, fmt.Errorf("key '%s' not found in the map", key)
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
		s := &[]storageTY.Sort{}
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

	filters := make([]storageTY.Filter, 0)

	for k, v := range f {
		if k != "filter" {
			filters = append(filters, storageTY.Filter{
				Key:   k,
				Value: v,
			})
		}
	}

	if fj, ok := f["filter"]; ok {
		fs := &[]storageTY.Filter{}
		err := json.Unmarshal([]byte(fj), fs)
		if err != nil {
			return nil, nil, err
		}
		filters = append(filters, *fs...)
	}

	zap.L().Debug("received filters and pagination", zap.Any("filter", filters), zap.Any("pagination", p))

	return filters, &p, nil
}

func WriteResponse(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		zap.L().Error("error on writing response", zap.Error(err))
		return
	}
}

func PostErrorResponse(w http.ResponseWriter, message string, code int) {
	response := &webHandlerTY.Response{
		Success: false,
		Message: message,
	}
	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Error(w, string(out), code)
}

func PostSuccessResponse(w http.ResponseWriter, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		PostErrorResponse(w, err.Error(), 500)
		return
	}

	WriteResponse(w, out)
}
