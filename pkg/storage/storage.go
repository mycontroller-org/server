package storage

import (
	"errors"
	"fmt"

	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	"github.com/mycontroller-org/mycontroller-v2/plugin/storage/mongodb"
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
	Update(entity string, filter []ml.Filter, data interface{}) error
	Upsert(entityName string, filter []ml.Filter, d interface{}) error
	FindOne(entityName string, filter []ml.Filter, out interface{}) error
	Find(entityName string, filter []ml.Filter, pagination ml.Pagination, out interface{}) error
	Distinct(entityName string, fieldName string, filter []ml.Filter) ([]interface{}, error)
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
			c, err := mongodb.NewClient(config)
			if err != nil {
				return nil, err
			}
			var cl Client = c
			return cl, nil
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}
