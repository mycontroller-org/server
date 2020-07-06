package storage

import (
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	"github.com/mycontroller-org/mycontroller-v2/plugin/storage/mongo"
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

// Init storage
func Init(config map[string]interface{}) (*Client, error) {
	// Update this code, if we have more than one storage options

	c, err := mongo.NewClient(config)
	if err != nil {
		return nil, err
	}
	var cl Client = c
	return &cl, nil
}
