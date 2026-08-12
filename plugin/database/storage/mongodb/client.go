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
	logger := storageTY.GetStorageLogger().Named(loggerName)

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
		logger.Error("error on connecting to database", zap.Error(err))
		return nil, err
	}
	client := &Client{
		Config: cfg,
		Client: mongoClient,
		ctx:    ctx,
		logger: logger,
	}
	err = client.initIndex()
	if err != nil {
		logger.Error("error on creating index", zap.Error(err))
		return nil, err
	}
	logger.Debug("database connected successfully")
	return client, nil
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
		if result.Err() == mongoDriver.ErrNoDocuments {
			return storageTY.ErrNoDocuments
		}
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
		if err == mongoDriver.ErrNoDocuments {
			return -1, storageTY.ErrNoDocuments
		}
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
		if err == mongoDriver.ErrNoDocuments {
			return nil, storageTY.ErrNoDocuments
		}
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

// matchNothing is a predicate no document can satisfy. Used when a compound
// filter cannot be decoded, so an unusable constraint fails closed.
func matchNothing() bson.M {
	return bson.M{"_id": bson.M{"$in": []interface{}{}}}
}

func filter(filters []storageTY.Filter) *bson.M {
	bm := bson.M{}
	if len(filters) == 0 {
		return &bm
	}
	// Compound ops ($or / $nor) that must be AND-ed with field predicates.
	// Multiple top-level OperatorOr must not be flattened into one $or.
	andParts := make([]bson.M, 0)
	// fields already placed in bm, to detect a second predicate on the same field
	andedKeys := make(map[string]struct{}, len(filters))

	for _, _f := range filters {
		op := strings.ToLower(_f.Operator)

		// OR of AND-groups (RBAC list allow scope)
		if op == storageTY.OperatorOr {
			groups, ok := _f.Value.([][]storageTY.Filter)
			if !ok {
				// unusable scope filter: match nothing rather than drop the
				// constraint (these carry access-control scope)
				andParts = append(andParts, matchNothing())
				continue
			}
			orClauses := make([]bson.M, 0, len(groups))
			for _, group := range groups {
				sub := filter(group)
				if sub != nil && len(*sub) > 0 {
					orClauses = append(orClauses, *sub)
				}
			}
			if len(orClauses) == 1 {
				andParts = append(andParts, orClauses[0])
			} else if len(orClauses) > 1 {
				andParts = append(andParts, bson.M{"$or": orClauses})
			}
			continue
		}

		// NOR of one AND-group (RBAC Deny exclude)
		if op == storageTY.OperatorNor {
			group, ok := _f.Value.([]storageTY.Filter)
			if !ok {
				andParts = append(andParts, matchNothing())
				continue
			}
			sub := filter(group)
			if sub != nil && len(*sub) > 0 {
				andParts = append(andParts, bson.M{"$nor": []bson.M{*sub}})
			}
			continue
		}

		fl := strings.ToLower(_f.Key)
		predicate, ok := fieldPredicate(op, _f.Value)
		if !ok {
			continue
		}
		// Two filters on the same field must both apply. Merging them into one
		// bson.M key would silently keep only the last one, which drops access
		// control scope: an allow on id plus a deny on id would leave only the deny
		// (= everything except), and a client filter could override a path id.
		if existing, duplicate := bm[fl]; duplicate {
			andParts = append(andParts, bson.M{fl: existing}, bson.M{fl: predicate})
			delete(bm, fl)
			continue
		}
		if _, alreadyAnded := andedKeys[fl]; alreadyAnded {
			andParts = append(andParts, bson.M{fl: predicate})
			continue
		}
		bm[fl] = predicate
		andedKeys[fl] = struct{}{}
	}

	if len(andParts) == 0 {
		return &bm
	}
	if len(bm) == 0 && len(andParts) == 1 {
		return &andParts[0]
	}
	parts := make([]bson.M, 0, 1+len(andParts))
	if len(bm) > 0 {
		parts = append(parts, bm)
	}
	parts = append(parts, andParts...)
	return &bson.M{"$and": parts}
}

// fieldPredicate converts one operator into its mongo predicate.
// ok is false for an unsupported operator (the filter is then ignored, as before).
func fieldPredicate(operator string, value interface{}) (interface{}, bool) {
	switch operator {
	case storageTY.OperatorNone:
		return value, true
	case storageTY.OperatorEqual:
		return bson.M{"$eq": value}, true
	case storageTY.OperatorNotEqual:
		return bson.M{"$ne": value}, true
	case storageTY.OperatorIn:
		return bson.M{"$in": value}, true
	case storageTY.OperatorNotIn:
		return bson.M{"$nin": value}, true
	case storageTY.OperatorGreaterThan:
		return bson.M{"$gt": value}, true
	case storageTY.OperatorLessThan:
		return bson.M{"$lt": value}, true
	case storageTY.OperatorGreaterThanEqual:
		return bson.M{"$gte": value}, true
	case storageTY.OperatorLessThanEqual:
		return bson.M{"$lte": value}, true
	case storageTY.OperatorExists:
		return bson.M{"$exists": value}, true
	case storageTY.OperatorRegex:
		return bson.M{"$regex": value, "$options": "i"}, true
	case storageTY.OperatorRegexCaseSensitive:
		return bson.M{"$regex": value}, true
	default:
		return nil, false
	}
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
