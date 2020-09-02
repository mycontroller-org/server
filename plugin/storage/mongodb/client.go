package mongodb

import (
	"context"
	"errors"
	"reflect"
	"strings"

	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
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
	mc, err := mg.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Config: cfg,
		Client: mc,
	}
	err = c.initIndex()
	return c, err
}

// Close the connection
func (c *Client) Close() error {
	return c.Client.Disconnect(ctx)
}

// Ping to the target database
func (c *Client) Ping() error {
	return c.Client.Ping(ctx, nil)
}

func (c *Client) getCollection(e string) *mg.Collection {
	return c.Client.Database(c.Config.Database).Collection(e)
}

// Insert the entity
func (c *Client) Insert(e string, d interface{}) error {
	if d == nil {
		return errors.New("No data provided")
	}
	cl := c.getCollection(e)
	_, err := cl.InsertOne(ctx, d)
	return err
}

// Update the entity
func (c *Client) Update(e string, f []pml.Filter, d interface{}) error {
	if d == nil {
		return errors.New("No data provided")
	}
	cl := c.getCollection(e)
	_, err := cl.ReplaceOne(ctx, defaultFilter(f, d), d)
	return err
}

// Upsert date into database
func (c *Client) Upsert(e string, f []pml.Filter, d interface{}) error {
	if d == nil {
		return errors.New("No data provided")
	}
	cl := c.getCollection(e)

	// find the entity, if available update it
	or, err := cl.ReplaceOne(ctx, defaultFilter(f, d), d)
	if err != nil {
		return err
	}
	if or.MatchedCount == 0 {
		_, err := cl.InsertOne(ctx, d)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindOne returns data
func (c *Client) FindOne(e string, f []pml.Filter, out interface{}) error {
	cl := c.getCollection(e)
	res := cl.FindOne(ctx, filter(f))
	if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(out)
}

// Delete by filter
func (c *Client) Delete(e string, f []pml.Filter) (int64, error) {
	cl := c.getCollection(e)
	fo := options.Delete()
	dr, err := cl.DeleteMany(ctx, filter(f), fo)
	if err != nil {
		return -1, err
	}
	return dr.DeletedCount, nil
}

// Count returns available documents count from a collection
func (c *Client) Count(e string, f []pml.Filter) (int64, error) {
	cl := c.getCollection(e)
	bm := filter(f)
	return cl.CountDocuments(ctx, bm)
}

// Distinct returns data
func (c *Client) Distinct(e string, fn string, f []pml.Filter) ([]interface{}, error) {
	cl := c.getCollection(e)
	rs, err := cl.Distinct(ctx, fn, filter(f))
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// Find returns data
func (c *Client) Find(e string, f []pml.Filter, p *pml.Pagination, out interface{}) error {
	p = ut.UpdatePagination(p)
	cl := c.getCollection(e)
	sm := sort(p.SortBy)
	fo := options.Find()
	if p.Limit != -1 {
		fo.SetLimit(p.Limit)
	}
	if p.Offset != -1 {
		fo.SetSkip(p.Offset)
	}
	fo.SetSort(sm)
	cur, err := cl.Find(ctx, filter(f), fo)
	if err != nil {
		return err
	}
	return cur.All(ctx, out)
}

func idFilter(d interface{}) *bson.M {
	var id string
	if reflect.ValueOf(d).Kind() == reflect.Ptr {
		id = reflect.ValueOf(d).Elem().FieldByName("ID").String()
	} else if reflect.ValueOf(d).Kind() == reflect.Struct {
		id = reflect.ValueOf(d).FieldByName("ID").String()
	} else {
		return &bson.M{}
	}
	return &bson.M{"id": id}
}

func defaultFilter(f []pml.Filter, d interface{}) *bson.M {
	if f == nil || len(f) == 0 {
		return idFilter(d)
	}
	return filter(f)
}

func filter(f []pml.Filter) *bson.M {
	bm := bson.M{}
	if f == nil || len(f) == 0 {
		return &bm
	}
	for _, _f := range f {
		fl := strings.ToLower(_f.Key)
		switch strings.ToLower(_f.Operator) {
		case "":
			bm[fl] = _f.Value
		case "eq":
			bm[fl] = bson.M{"$eq": _f.Value}
		case "ne":
			bm[fl] = bson.M{"$ne": _f.Value}
		case "in":
			bm[fl] = bson.M{"$in": _f.Value}
		case "nin":
			bm[fl] = bson.M{"$nin": _f.Value}
		case "gt":
			bm[fl] = bson.M{"$gt": _f.Value}
		case "lt":
			bm[fl] = bson.M{"$lt": _f.Value}
		case "gte":
			bm[fl] = bson.M{"$gte": _f.Value}
		case "lte":
			bm[fl] = bson.M{"$lte": _f.Value}
		case "exists":
			bm[fl] = bson.M{"$exists": _f.Value}
		case "regex":
			bm[fl] = bson.M{"$regex": _f.Value, "$options": "i"}
		}
	}
	return &bm
}

func sort(s []pml.Sort) *bson.M {
	bm := bson.M{}
	if s == nil || len(s) == 0 {
		return &bm
	}
	for _, _s := range s {
		switch strings.ToLower(_s.OrderBy) {
		case "", "asc":
			bm[_s.Field] = 1
		case "desc":
			bm[_s.Field] = -1
		}
	}
	return &bm
}
