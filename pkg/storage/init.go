package storage

import (
	"errors"
	"fmt"

	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"github.com/mycontroller-org/backend/v2/plugin/storage/mongodb"
)

// Storage database types
const (
	TypeMongoDB = "mongodb"
	TypeSqllite = "sqllite"
)

// Client interface
type Client interface {
	Close() error
	Ping() error
	Insert(entity string, data interface{}) error
	Update(entity string, filter []pml.Filter, data interface{}) error
	Upsert(entityName string, filter []pml.Filter, d interface{}) error
	FindOne(entityName string, filter []pml.Filter, out interface{}) error
	Find(entityName string, filter []pml.Filter, pagination *pml.Pagination, out interface{}) error
	Distinct(entityName string, fieldName string, filter []pml.Filter) ([]interface{}, error)
}

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

// Init storage
func Init(config map[string]interface{}) (Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		/*
			// NOTE: badger support needs gcc installed on the build environment
			// addeds additional 6 MB on application bin file
				case TypeBadger:
					c, err := badger.NewClient(config)
					if err != nil {
						return nil, err
					}
					var cl Client = c
					return &cl, nil
		*/
		case TypeMongoDB:
			return mongodb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}
