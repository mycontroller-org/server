package mongodb

import (
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// modelIndex for entities
var modelIndex = map[string][]mongo.IndexModel{
	ml.EntityGateway: {
		{Keys: bson.M{"name": -1}, Options: options.Index().SetName("index_name")},
		// Error: Only one text index allowed
		// {Keys: bson.M{"location": "text", "date": -1}, Options: options.Index().SetName("location_date_index").SetUnique(true)},
	},
	// model.EntityNode: {
	// 	{Keys: bson.M{"name": "text"}, Options: options.Index().SetName("index_name").SetUnique(false)},
	// },
}

// initIndex creates index, if not available
// https://docs.mongodb.com/manual/reference/method/db.collection.createIndexes/#recreating-an-existing-index
func (c *Client) initIndex() error {
	d := c.Client.Database(c.Config.Database)
	for k, v := range modelIndex {
		names, err := d.Collection(k).Indexes().CreateMany(ctx, v)
		if err != nil {
			return err
		}
		zap.L().Debug("Index created", zap.String("entity", k), zap.Any("index", names))
	}
	return nil
}
