package storage

import (
	"context"
	"errors"

	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
)

// Plugin interface storage
type Plugin interface {
	Name() string
	Ping() error
	Close() error
	Insert(entityName string, data interface{}) error
	Upsert(entityName string, data interface{}, filter []Filter) error
	Update(entityName string, data interface{}, filter []Filter) error
	FindOne(entityName string, out interface{}, filter []Filter) error
	Find(entityName string, out interface{}, filter []Filter, pagination *Pagination) (*Result, error)
	Delete(entityName string, filter []Filter) (int64, error)
	Pause() error
	Resume() error
	ClearDatabase() error
	// if your database needs start import, returns "true", "files-directory" and "file-format"
	// example: true, "/tmp/data", "yaml"
	// it is required for in-memory database at startup
	DoStartupImport() (bool, string, string)
}

func FromContext(ctx context.Context) (Plugin, error) {
	bus, ok := ctx.Value(contextTY.STORAGE_DB).(Plugin)
	if !ok {
		return nil, errors.New("invalid storage instance received in context")
	}
	if bus == nil {
		return nil, errors.New("storage instance not provided in context")
	}
	return bus, nil
}

func WithContext(ctx context.Context, bus Plugin) context.Context {
	return context.WithValue(ctx, contextTY.STORAGE_DB, bus)
}

// Storage database types
const (
	TypeMemory  = "memory"
	TypeMongoDB = "mongodb"
)

// Pagination options

// SortBy values
const (
	SortByASC  = "asc"
	SortByDESC = "desc"
)

// Operators
const (
	OperatorNone             = ""
	OperatorEqual            = "eq"
	OperatorNotEqual         = "ne"
	OperatorIn               = "in"
	OperatorNotIn            = "nin"
	OperatorRangeIn          = "range_in"
	OperatorRangeNotIn       = "range_not_in"
	OperatorGreaterThan      = "gt"
	OperatorLessThan         = "lt"
	OperatorGreaterThanEqual = "gte"
	OperatorLessThanEqual    = "lte"
	OperatorExists           = "exists"
	OperatorRegex            = "regex"
)

// Sort options
type Sort struct {
	Field   string `json:"f"`
	OrderBy string `json:"o"`
}

// Pagination configuration
type Pagination struct {
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
	SortBy []Sort `json:"sortBy"`
}

// Filter used to limit the result
type Filter struct {
	Key      string      `json:"k"`
	Operator string      `json:"o"`
	Value    interface{} `json:"v"`
}

// Result returns a list of data
type Result struct {
	Count  int64       `json:"count"`
	Limit  int64       `json:"limit"`
	Offset int64       `json:"offset"`
	Data   interface{} `json:"data"`
}
