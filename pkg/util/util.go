package util

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
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
func UpdatePagination(p *pml.Pagination) {
	if p == nil {
		p = &pml.Pagination{}
	}
	if len(p.SortBy) == 0 {
		p.SortBy = []pml.Sort{{Field: "ID", OrderBy: "ASC"}}
	}
	if p.Limit == 0 {
		p.Limit = -1
	}
}

// JoinMap joins two maps. put all the values into 'p' map from 'o' map
func JoinMap(p, o map[string]interface{}) {
	if o == nil {
		return
	}
	if p == nil {
		p = map[string]interface{}{}
	}
	for k, v := range o {
		p[k] = v
	}
}

// MapToStruct converts string to struct
func MapToStruct(tagName string, in map[string]interface{}, out interface{}) error {
	if tagName == "" {
		return mapstructure.Decode(in, out)
	}
	cfg := &mapstructure.DecoderConfig{TagName: tagName, Result: out}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(in)
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
