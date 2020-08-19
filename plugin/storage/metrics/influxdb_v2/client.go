package influx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api/write"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	sml "github.com/mycontroller-org/mycontroller-v2/pkg/model/sensor"
	"github.com/mycontroller-org/mycontroller-v2/pkg/util"
	"go.uber.org/zap"
)

var ctx = context.TODO()

const (
	defaultFlushInterval = 1 * time.Second
)

// Config of the influxdb_v2
type Config struct {
	Name          string
	Organization  string
	Bucket        string
	URI           string
	Token         string
	Username      string
	Password      string
	FlushInterval string `yaml:"flush_interval"`
}

// Client of the influxdb
type Client struct {
	Client  influxdb2.Client
	Config  Config
	stop    chan bool
	buffer  []*sml.SensorField
	rwMutex *sync.RWMutex
}

// NewClient of influxdb
func NewClient(config map[string]interface{}) (*Client, error) {
	cfg := Config{}
	err := util.MapToStruct(util.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}
	token := cfg.Token
	if token == "" {
		token = fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
	}
	flushInterval := defaultFlushInterval

	if cfg.FlushInterval != "" {
		flushInterval, err = time.ParseDuration(cfg.FlushInterval)
		if err != nil {
			zap.L().Warn("Invalid flush interval", zap.String("flushInterval", cfg.FlushInterval))
			flushInterval = defaultFlushInterval
		}
	}
	if flushInterval.Milliseconds() < 1 {
		zap.L().Warn("Minimum supported flush interval is 1ms, switching back to default", zap.String("flushInterval", cfg.FlushInterval))
		flushInterval = defaultFlushInterval
	}
	opts := influxdb2.DefaultOptions()
	opts.SetFlushInterval(uint(flushInterval.Milliseconds()))
	iClient := influxdb2.NewClient(cfg.URI, cfg.Token)
	c := &Client{
		Config:  cfg,
		Client:  iClient,
		buffer:  make([]*sml.SensorField, 0),
		stop:    make(chan bool),
		rwMutex: &sync.RWMutex{},
	}
	err = c.Ping()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Ping to target database
func (c *Client) Ping() error {
	s, err := c.Client.Ready(ctx)
	if err != nil {
		return err
	}
	if !s {
		return errors.New("Influx server not ready yet")
	}
	return nil
}

// Close the influxdb connection
func (c *Client) Close() error {
	// close bulk insert
	close(c.stop)
	c.Client.Close()
	return nil
}

// WriteBlocking implementation
func (c *Client) WriteBlocking(sf *sml.SensorField) error {
	p, err := getPoint(sf)
	if err != nil {
		return err
	}
	wb := c.Client.WriteApiBlocking(c.Config.Organization, c.Config.Bucket)
	return wb.WritePoint(ctx, p)
}

func (c *Client) Write(sf *sml.SensorField) error {
	p, err := getPoint(sf)
	if err != nil {
		return err
	}
	w := c.Client.WriteApi(c.Config.Organization, c.Config.Bucket)
	w.WritePoint(p)
	return nil
}

func getPoint(sf *sml.SensorField) (*write.Point, error) {
	fields := make(map[string]interface{})
	if sf.PayloadType == msg.PayloadTypeGeo {
		_f, err := geoData(sf.Payload)
		if err != nil {
			return nil, err
		}
		fields = _f
	} else {
		fields["value"] = sf.Payload
	}
	p := influxdb2.NewPoint(measurementName(sf.PayloadType),
		// "gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID,
		map[string]string{"gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID, "id": sf.ID},
		fields,
		sf.LastSeen,
	)
	return p, nil
}

func geoData(pl interface{}) (map[string]interface{}, error) {
	// payload should be in this format
	// latitude;longitude;altitude. E.g. "55.722526;13.017972;18"
	d := make(map[string]interface{})
	ds := strings.Split(pl.(string), ";")
	if len(ds) < 2 {
		return nil, fmt.Errorf("Invalid geo data: %s", pl)
	}
	lat, err := strconv.ParseFloat(ds[0], 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid float data: %s", pl)
	}
	lon, err := strconv.ParseFloat(ds[1], 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid float data: %s", pl)
	}
	alt := float64(0)
	if len(ds[0]) > 2 {
		alt, err = strconv.ParseFloat(ds[2], 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid float data: %s", pl)
		}
	}

	d["latitude"] = lat
	d["longitude"] = lon
	d["altitude"] = alt

	return d, nil
}

func measurementName(payloadType string) string {
	switch payloadType {
	case msg.PayloadTypeBoolean:
		return "binary_data"
	case msg.PayloadTypeFloat:
		return "float_data"
	case msg.PayloadTypeInteger:
		return "integer_data"
	case msg.PayloadTypeString:
		return "string_data"
	case msg.PayloadTypeGeo:
		return "geo_data"
	default:
		return "unknown_data_type"
	}
}
