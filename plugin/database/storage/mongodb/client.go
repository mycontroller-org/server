package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var ctx = context.TODO()

const (
	PluginMongoDB = storageTY.TypeMongoDB

	DefaultCollectionPrefix = "mc_"

	loggerName = "mongodb"
)

// Config of the database
type Config struct {
	Name             string `yaml:"name"`
	Database         string `yaml:"database"`
	URI              string `yaml:"uri"`
	CollectionPrefix string `yaml:"collection_prefix"`
}

// Client of the mongo db
type Client struct {
	Client *mongoDriver.Client
	Config Config
	ctx    context.Context
	logger *zap.Logger
}

// New mongodb
func New(ctx context.Context, config cmap.CustomMap) (storageTY.Plugin, error) {
	logger := storageTY.GetStorageLogger()

	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	// update collection prefix
	if cfg.CollectionPrefix == "" {
		cfg.CollectionPrefix = DefaultCollectionPrefix
	}

	clientOptions := options.Client().ApplyURI(cfg.URI)

	mongoClient, err := mongoDriver.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Config: cfg,
		Client: mongoClient,
		ctx:    ctx,
		logger: logger.Named(loggerName),
	}
	err = client.initIndex()
	return client, err
}

func (s *Client) Name() string {
	return PluginMongoDB
}

// DoStartupImport returns the needs, files location, and file format
func (s *Client) DoStartupImport() (bool, string, string) {
	return false, "", ""
}

// Pause the database to perform import like jobs
func (c *Client) Pause() error {
	return nil
}

// Resume the database if Paused
func (c *Client) Resume() error {
	return nil
}

// ClearDatabase removes all the data from the database
func (c *Client) ClearDatabase() error {
	filter := bson.D{{Key: "name", Value: primitive.Regex{Pattern: fmt.Sprintf("^%s*", c.Config.CollectionPrefix), Options: "i"}}}
	collections, err := c.Client.Database(c.Config.Database).ListCollectionNames(c.ctx, filter)
	if err != nil {
		return err
	}
	c.logger.Info("about to drop the collections", zap.Any("collections", collections))

	for _, collectionName := range collections {
		err = c.Client.Database(c.Config.Database).Collection(collectionName).Drop(c.ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close the connection
func (c *Client) Close() error {
	return c.Client.Disconnect(ctx)
}

// Ping to the target database
func (c *Client) Ping() error {
	return c.Client.Ping(ctx, nil)
}

func (c *Client) getCollection(entityName string) *mongoDriver.Collection {
	collectionName := fmt.Sprintf("%s%s", c.Config.CollectionPrefix, entityName)
	return c.Client.Database(c.Config.Database).Collection(collectionName)
}

// Insert the entity
func (c *Client) Insert(entityName string, data interface{}) error {
	if data == nil {
		return storageTY.ErrNilData
	}
	collection := c.getCollection(entityName)
	_, err := collection.InsertOne(ctx, data)
	return err
}

// Update the entity
func (c *Client) Update(entityName string, data interface{}, filters []storageTY.Filter) error {
	if data == nil {
		return storageTY.ErrNilData
	}
	collection := c.getCollection(entityName)
	_, err := collection.ReplaceOne(ctx, defaultFilter(filters, data), data)
	if err == mongoDriver.ErrNoDocuments {
		return storageTY.ErrNoDocuments
	}
	return err
}

// Upsert date into database
func (c *Client) Upsert(entityName string, data interface{}, filters []storageTY.Filter) error {
	if data == nil {
		return storageTY.ErrNilData
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
func (c *Client) FindOne(entityName string, out interface{}, filters []storageTY.Filter) error {
	cl := c.getCollection(entityName)
	result := cl.FindOne(ctx, filter(filters))
	if result.Err() != nil {
		return result.Err()
	}
	return result.Decode(out)
}

// Delete by filter
func (c *Client) Delete(entityName string, filters []storageTY.Filter) (int64, error) {
	if filters == nil {
		return -1, storageTY.ErrNilFilter
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
func (c *Client) Count(entityName string, filters []storageTY.Filter) (int64, error) {
	collection := c.getCollection(entityName)
	filterOption := filter(filters)
	return collection.CountDocuments(ctx, filterOption)
}

// Find returns data
func (c *Client) Find(entityName string, out interface{}, filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	pagination = utils.UpdatePagination(pagination)
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
	result := &storageTY.Result{
		Count:  count,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Data:   out,
	}
	return result, nil
}

func idFilter(data interface{}) *bson.M {
	id := filterUtils.GetID(data)
	if id == "" {
		return &bson.M{}
	}
	return &bson.M{"id": id}
}

func defaultFilter(filters []storageTY.Filter, data interface{}) *bson.M {
	if len(filters) == 0 {
		return idFilter(data)
	}
	return filter(filters)
}

func filter(filters []storageTY.Filter) *bson.M {
	bm := bson.M{}
	if len(filters) == 0 {
		return &bm
	}
	for _, _f := range filters {
		fl := strings.ToLower(_f.Key)
		switch strings.ToLower(_f.Operator) {
		case storageTY.OperatorNone:
			bm[fl] = _f.Value

		case storageTY.OperatorEqual:
			bm[fl] = bson.M{"$eq": _f.Value}

		case storageTY.OperatorNotEqual:
			bm[fl] = bson.M{"$ne": _f.Value}

		case storageTY.OperatorIn:
			bm[fl] = bson.M{"$in": _f.Value}

		case storageTY.OperatorNotIn:
			bm[fl] = bson.M{"$nin": _f.Value}

		case storageTY.OperatorGreaterThan:
			bm[fl] = bson.M{"$gt": _f.Value}

		case storageTY.OperatorLessThan:
			bm[fl] = bson.M{"$lt": _f.Value}

		case storageTY.OperatorGreaterThanEqual:
			bm[fl] = bson.M{"$gte": _f.Value}

		case storageTY.OperatorLessThanEqual:
			bm[fl] = bson.M{"$lte": _f.Value}

		case storageTY.OperatorExists:
			bm[fl] = bson.M{"$exists": _f.Value}

		case storageTY.OperatorRegex:
			bm[fl] = bson.M{"$regex": _f.Value, "$options": "i"}
		}
	}
	return &bm
}

func sort(sort []storageTY.Sort) *bson.M {
	bm := bson.M{}
	if len(sort) == 0 {
		return &bm
	}
	for _, _s := range sort {
		filed := strings.ToLower(_s.Field)
		switch strings.ToLower(_s.OrderBy) {
		case "", storageTY.SortByASC:
			bm[filed] = 1
		case storageTY.SortByDESC:
			bm[filed] = -1
		}
	}
	return &bm
}
