package influx

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go"
	db "github.com/influxdata/influxdb-client-go"
	m2s "github.com/mitchellh/mapstructure"
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
	msg "github.com/mycontroller-org/mycontroller/pkg/model/message"
	"go.uber.org/zap"
)

var ctx = context.TODO()

const (
	bulkInsertInterval = 1 * time.Second
)

// Config of the influxdb_v2
type Config struct {
	Name         string
	Organization string
	Bucket       string
	URI          string
	Token        string
	Username     string
	Password     string
}

// Client of the influxdb
type Client struct {
	Client  *influxdb.Client
	Config  Config
	stop    chan bool
	buffer  []*ml.SensorField
	rwMutex *sync.RWMutex
}

// NewClient of influxdb
func NewClient(config map[string]string) (*Client, error) {
	var cfg Config
	err := m2s.Decode(config, &cfg)
	if err != nil {
		return nil, err
	}
	var ic *influxdb.Client
	if cfg.Token != "" {
		ic, err = influxdb.New(cfg.URI, cfg.Token)
	} else {
		ic, err = influxdb.New(cfg.URI, "", influxdb.WithUserAndPass(cfg.Username, cfg.Password))
	}
	if err != nil {
		return nil, err
	}
	c := &Client{
		Config:  cfg,
		Client:  ic,
		buffer:  make([]*ml.SensorField, 0),
		stop:    make(chan bool),
		rwMutex: &sync.RWMutex{},
	}
	err = c.Ping()
	if err != nil {
		return nil, err
	}
	// start bulk writer
	c.startBulkWriter()
	return c, nil
}

// Ping to target database
func (c *Client) Ping() error {
	return c.Client.Ping(ctx)
}

// Close the influxdb connection
func (c *Client) Close() error {
	// close bulk insert
	close(c.stop)
	return c.Client.Close()
}

// WriteOld definition
func (c *Client) WriteOld(sf *ml.SensorField) error {
	fields := make(map[string]interface{})
	if sf.DataType == msg.DataTypeGeo {
		_f, err := geoData(sf.Payload)
		if err != nil {
			return err
		}
		fields = _f
	} else {
		fields["value"] = sf.Payload
	}
	metrics := []db.Metric{
		db.NewRowMetric(
			fields,
			measurementName(sf.DataType),
			// "gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID,
			map[string]string{"gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID, "id": sf.ID},
			sf.LastSeen,
		),
	}
	_, err := c.Client.Write(ctx, c.Config.Bucket, c.Config.Organization, metrics...)
	return err
}

func (c *Client) Write(sf *ml.SensorField) error {
	//return c.WriteOld(sf)

	c.rwMutex.Lock()
	c.buffer = append(c.buffer, sf)
	c.rwMutex.Unlock()
	return nil

}

func (c *Client) startBulkWriter() {
	ticker := time.NewTicker(bulkInsertInterval)
	go func() {
		for {
			select {
			case <-c.stop:
				ticker.Stop()
				return
			case <-ticker.C:
				c.rwMutex.Lock()
				d := c.buffer
				c.buffer = make([]*ml.SensorField, 0)
				c.rwMutex.Unlock()
				// if there is a data in buffer
				if len(d) > 0 {
					zap.L().Debug("Buffer size", zap.Int("size", len(d)))
					metrics := make([]db.Metric, 0)
					for _, sf := range d {
						fields := make(map[string]interface{})
						if sf.DataType == msg.DataTypeGeo {
							_f, err := geoData(sf.Payload.Value)
							if err != nil {
								zap.L().Error("Failed to parse geo data", zap.Error(err), zap.Any("data", sf))
							}
							fields = _f
						} else {
							fields["value"] = sf.Payload.Value
						}
						metrics = append(metrics,
							db.NewRowMetric(
								fields,
								measurementName(sf.DataType),
								// "gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID,
								map[string]string{"gateway": sf.GatewayID, "node": sf.NodeID, "sensor": sf.SensorID, "id": sf.ID},
								sf.LastSeen,
							),
						)
					}

					start := time.Now()
					_, err := c.Client.Write(ctx, c.Config.Bucket, c.Config.Organization, metrics...)
					if err != nil {
						zap.L().Error("Failed to write into metrics store", zap.Error(err))
					} else {
						zap.L().Debug("Bulk metrics insert completed", zap.String("timeTaken", time.Since(start).String()))
					}
				}

			}
		}
	}()

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

func measurementName(dataType string) string {
	switch dataType {
	case msg.DataTypeBoolean:
		return "binary_data"
	case msg.DataTypeFloat:
		return "float_data"
	case msg.DataTypeInteger:
		return "integer_data"
	case msg.DataTypeString:
		return "string_data"
	case msg.DataTypeGeo:
		return "geo_data"
	default:
		return "unknown_data_type"
	}
}
