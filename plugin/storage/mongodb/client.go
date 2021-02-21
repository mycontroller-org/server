package mongodb

import (
	"context"
	"errors"
	"strings"

	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.TODO()

// Config of the database
type Config struct {
	Name     string
	Database string
	URI      string
}

// Client of the mongo db
type Client struct {
	Client *mg.Client
	Config Config
}

// NewClient mongodb
func NewClient(config map[string]interface{}) (*Client, error) {
	cfg := Config{}
	err := ut.MapToStruct(ut.TagNameNone, config, &cfg)
	if err != nil {
		return nil, err
	}
	clientOptions := options.Client().ApplyURI(cfg.URI)

	mongoClient, err := mg.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Config: cfg,
		Client: mongoClient,
	}
	err = client.initIndex()
	return client, err
}

// Close the connection
func (c *Client) Close() error {
	return c.Client.Disconnect(ctx)
}

// Ping to the target database
func (c *Client) Ping() error {
	return c.Client.Ping(ctx, nil)
}

func (c *Client) getCollection(entityName string) *mg.Collection {
	return c.Client.Database(c.Config.Database).Collection(entityName)
}

// Insert the entity
func (c *Client) Insert(entityName string, data interface{}) error {
	if data == nil {
		return errors.New("No data provided")
	}
	collection := c.getCollection(entityName)
	_, err := collection.InsertOne(ctx, data)
	return err
}

// Update the entity
func (c *Client) Update(entityName string, data interface{}, filters []stgml.Filter) error {
	if data == nil {
		return errors.New("No data provided")
	}
	collection := c.getCollection(entityName)
	_, err := collection.ReplaceOne(ctx, defaultFilter(filters, data), data)
	return err
}

// Upsert date into database
func (c *Client) Upsert(entityName string, data interface{}, filters []stgml.Filter) error {
	if data == nil {
		return errors.New("No data provided")
	}
	collection := c.getCollection(entityName)

	// find the entity, if available update it
	updateResult, err := collection.ReplaceOne(ctx, defaultFilter(filters, data), data)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		_, err := collection.InsertOne(ctx, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindOne returns data
func (c *Client) FindOne(entityName string, out interface{}, filters []stgml.Filter) error {
	cl := c.getCollection(entityName)
	result := cl.FindOne(ctx, filter(filters))
	if result.Err() != nil {
		return result.Err()
	}
	return result.Decode(out)
}

// Delete by filter
func (c *Client) Delete(entityName string, filters []stgml.Filter) (int64, error) {
	if filters == nil {
		return -1, errors.New("Filter should not be nil")
	}
	collection := c.getCollection(entityName)
	filterOption := options.Delete()
	deleteResult, err := collection.DeleteMany(ctx, filter(filters), filterOption)
	if err != nil {
		return -1, err
	}
	return deleteResult.DeletedCount, nil
}

// Count returns available documents count from a collection
func (c *Client) Count(entityName string, filters []stgml.Filter) (int64, error) {
	collection := c.getCollection(entityName)
	filterOption := filter(filters)
	return collection.CountDocuments(ctx, filterOption)
}

// Find returns data
func (c *Client) Find(entityName string, out interface{}, filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	pagination = ut.UpdatePagination(pagination)
	collection := c.getCollection(entityName)
	sortOption := sort(pagination.SortBy)
	findOption := options.Find()
	if pagination.Limit != -1 {
		findOption.SetLimit(pagination.Limit)
	}
	if pagination.Offset != -1 {
		findOption.SetSkip(pagination.Offset)
	}
	findOption.SetSort(sortOption)
	cur, err := collection.Find(ctx, filter(filters), findOption)
	if err != nil {
		return nil, err
	}
	err = cur.All(ctx, out)
	if err != nil {
		return nil, err
	}

	count, err := c.Count(entityName, filters)
	if err != nil {
		return nil, err
	}
	result := &stgml.Result{
		Count:  count,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Data:   out,
	}
	return result, nil
}

func idFilter(data interface{}) *bson.M {
	id := helper.GetID(data)
	if id == "" {
		return &bson.M{}
	}
	return &bson.M{"id": id}
}

func defaultFilter(filters []stgml.Filter, data interface{}) *bson.M {
	if filters == nil || len(filters) == 0 {
		return idFilter(data)
	}
	return filter(filters)
}

func filter(filters []stgml.Filter) *bson.M {
	bm := bson.M{}
	if filters == nil || len(filters) == 0 {
		return &bm
	}
	for _, _f := range filters {
		fl := strings.ToLower(_f.Key)
		switch strings.ToLower(_f.Operator) {
		case stgml.OperatorNone:
			bm[fl] = _f.Value

		case stgml.OperatorEqual:
			bm[fl] = bson.M{"$eq": _f.Value}

		case stgml.OperatorNotEqual:
			bm[fl] = bson.M{"$ne": _f.Value}

		case stgml.OperatorIn:
			bm[fl] = bson.M{"$in": _f.Value}

		case stgml.OperatorNotIn:
			bm[fl] = bson.M{"$nin": _f.Value}

		case stgml.OperatorGreaterThan:
			bm[fl] = bson.M{"$gt": _f.Value}

		case stgml.OperatorLessThan:
			bm[fl] = bson.M{"$lt": _f.Value}

		case stgml.OperatorGreaterThanEqual:
			bm[fl] = bson.M{"$gte": _f.Value}

		case stgml.OperatorLessThanEqual:
			bm[fl] = bson.M{"$lte": _f.Value}

		case stgml.OperatorExists:
			bm[fl] = bson.M{"$exists": _f.Value}

		case stgml.OperatorRegex:
			bm[fl] = bson.M{"$regex": _f.Value, "$options": "i"}
		}
	}
	return &bm
}

func sort(sort []stgml.Sort) *bson.M {
	bm := bson.M{}
	if sort == nil || len(sort) == 0 {
		return &bm
	}
	for _, _s := range sort {
		filed := strings.ToLower(_s.Field)
		switch strings.ToLower(_s.OrderBy) {
		case "", stgml.SortByASC:
			bm[filed] = 1
		case stgml.SortByDESC:
			bm[filed] = -1
		}
	}
	return &bm
}
