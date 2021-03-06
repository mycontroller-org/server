package utils

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// contants
const (
	charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	TagNameYaml = "yaml"
	TagNameJSON = "json"
	TagNameNone = ""
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandUUID returns random uuid
func RandUUID() string {
	return uuid.New().String()
}

// RandID returns random id
func RandID() string {
	return RandIDWithLength(10)
}

// RandIDWithLength returns random id with supplied charset
func RandIDWithLength(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// UpdatePagination updates if nil
func UpdatePagination(pagination *stgml.Pagination) *stgml.Pagination {
	if pagination == nil {
		pagination = &stgml.Pagination{Limit: -1, Offset: -1}
	}
	if len(pagination.SortBy) == 0 {
		pagination.SortBy = []stgml.Sort{{Field: "ID", OrderBy: "ASC"}}
	}
	if pagination.Limit == 0 {
		pagination.Limit = -1
	}
	//if p.Offset == 0 {
	//	p.Offset = -1
	//}
	return pagination
}

// JoinMap joins two maps. put all the values into 'dst' map from 'src' map
func JoinMap(dst, src map[string]interface{}) {
	if src == nil {
		return
	}
	if dst == nil {
		dst = map[string]interface{}{}
	}
	for k, v := range src {
		dst[k] = v
	}
}

// GetMapValue returns fetch and returns with a key. if not available returns default value
func GetMapValue(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if m == nil {
		return defaultValue
	}
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

// FindItem returns the availability status and location
func FindItem(slice []string, value string) (int, bool) {
	for i, item := range slice {
		if item == value {
			return i, true
		}
	}
	return -1, false
}

// ContainsString returns the available status
func ContainsString(slice []string, value string) bool {
	_, available := FindItem(slice, value)
	return available
}

// DeepClone a interface
func DeepClone(data interface{}) interface{} {
	newData := reflect.New(reflect.TypeOf(data).Elem())

	value := reflect.ValueOf(data).Elem()
	newValue := newData.Elem()
	for i := 0; i < value.NumField(); i++ {
		newField := newValue.Field(i)
		newField.Set(value.Field(i))
	}

	return newData.Interface()
}
