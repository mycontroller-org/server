package storage

import (
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
)

// Client interface
type Client interface {
	Close() error
	Ping() error
	Insert(entityName string, data interface{}) error
	Upsert(entityName string, data interface{}, filter []pml.Filter) error
	Update(entityName string, data interface{}, filter []pml.Filter) error
	FindOne(entityName string, out interface{}, filter []pml.Filter) error
	Find(entityName string, out interface{}, filter []pml.Filter, pagination *pml.Pagination) (*pml.Result, error)
	Delete(entityName string, filter []pml.Filter) (int64, error)
}

// Storage database types
const (
	DBTypeMemory  = "memory"
	DBTypeMongoDB = "mongodb"
)

// Operators
const (
	OperatorNone             = ""
	OperatorEqual            = "eq"
	OperatorNotEqual         = "ne"
	OperatorIn               = "in"
	OperatorNotIn            = "nin"
	OperatorGreaterThan      = "gt"
	OperatorLessThan         = "lt"
	OperatorGreaterThanEqual = "gte"
	OperatorLessThanEqual    = "lte"
	OperatorExists           = "exists"
	OperatorRegex            = "regex"
)

// SortBy values
const (
	SortByASC  = "asc"
	SortByDESC = "desc"
)
