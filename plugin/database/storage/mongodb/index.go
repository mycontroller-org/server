package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// modelIndex for entities
var modelIndex = map[string][]mongo.IndexModel{
	//	ml.EntityGateway: {
	//		{Keys: bson.M{"name": -1}, Options: options.Index().SetName("index_name")},
	// // Error: Only one text index allowed
	// //{Keys: bson.M{"location": "text", "date": -1}, Options: options.Index().SetName("location_date_index").SetUnique(true)},
	// //},
	// // model.EntityNode: {
	// //	{Keys: bson.M{"name": "text"}, Options: options.Index().SetName("index_name").SetUnique(false)},
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
		c.logger.Debug("index created", zap.String("entity", k), zap.Any("index", names))
	}
	return nil
}
