package util

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

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
func UpdatePagination(p *ml.Pagination) {
	if p == nil {
		p = &ml.Pagination{}
	}
	if len(p.SortBy) == 0 {
		p.SortBy = []ml.Sort{ml.Sort{Field: "ID", OrderBy: "ASC"}}
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
func MapToStruct(in map[string]interface{}, out interface{}) error {
	return mapstructure.Decode(in, out)
}
