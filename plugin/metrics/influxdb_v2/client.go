package influx

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	influxdb2log "github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	queryML "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2/query"
	queryV1 "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2/query_v1"
	queryV2 "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2/query_v2"
	"go.uber.org/zap"
)

var ctx = context.TODO()

const (
	defaultFlushInterval = 1 * time.Second
)

// Config of the influxdb_v2
type Config struct {
	Name               string       `yaml:"name"`
	Organization       string       `yaml:"organization"`
	Bucket             string       `yaml:"bucket"`
	URI                string       `yaml:"uri"`
	Token              string       `yaml:"token"`
	Username           string       `yaml:"username"`
	Password           string       `yaml:"password"`
	InsecureSkipVerify bool         `yaml:"insecure_skip_verify"`
	QueryClientVersion string       `yaml:"query_client_version"`
	FlushInterval      string       `yaml:"flush_interval"`
	Logger             LoggerConfig `yaml:"logger"`
}

// LoggerConfig struct
type LoggerConfig struct {
	Mode     string `yaml:"mode"`
	Encoding string `yaml:"encoding"`
	Level    string `yaml:"level"`
}

// Client of the influxdb
type Client struct {
	Client      influxdb2.Client
	queryClient queryML.QueryAPI
	Config      Config
	stop        chan bool
	buffer      []*fml.Field
	logger      *myLogger
	mutex       *sync.RWMutex
}

// global constants
const (
	MeasurementBinary       = "mc_binary_data"
	MeasurementGaugeInteger = "mc_gauge_int_data"
	MeasurementGaugeFloat   = "mc_gauge_float_data"
	MeasurementCounter      = "mc_counter_data"
	MeasurementString       = "mc_string_data"
	MeasurementGeo          = "mc_geo_data"

	QueryClientV1 = "v1"
	QueryClientV2 = "v2"
)

// variables
var (
	TagGateway = strings.ToLower(model.KeyGatewayID)
	TagNode    = strings.ToLower(model.KeyNodeID)
	TagSensor  = strings.ToLower(model.KeySensorID)
	TagField   = strings.ToLower(model.KeyFieldID)
	TagID      = strings.ToLower(model.KeyID)
)

// NewClient of influxdb
func NewClient(config map[string]interface{}) (*Client, error) {
	cfg := Config{}
	err := ut.MapToStruct(ut.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Token == "" && cfg.Username != "" {
		cfg.Token = fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
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

	// replace influxdb2 logger with our custom logger
	_logger := getLogger(cfg.Logger.Mode, cfg.Logger.Level, cfg.Logger.Encoding)
	influxdb2log.Log = _logger

	opts := influxdb2.DefaultOptions()
	opts.SetFlushInterval(uint(flushInterval.Milliseconds()))
	iClient := influxdb2.NewClient(cfg.URI, cfg.Token)

	c := &Client{
		Config: cfg,
		Client: iClient,
		buffer: make([]*fml.Field, 0),
		stop:   make(chan bool),
		mutex:  &sync.RWMutex{},
		logger: _logger,
	}

	err = c.Ping()
	if err != nil {
		return nil, err
	}

	selectedVersion := ""

	if cfg.QueryClientVersion != "" {
		ver := strings.ToLower(cfg.QueryClientVersion)
		if ver == QueryClientV1 || ver == QueryClientV2 {
			selectedVersion = ver
		} else {
			zap.L().Warn("invalid query client version, going with auto detection", zap.String("input", cfg.QueryClientVersion))
		}
	}

	if selectedVersion == "" {
		selectedVersion = QueryClientV2 // update default route, if non works

		// get version
		health, err := c.Client.Health(ctx)
		if err != nil {
			return nil, err
		}
		zap.L().Info("influxdb detected version data", zap.String("name", cfg.Name), zap.Any("health data", health))

		detectedVersion := *health.Version
		if strings.HasPrefix(detectedVersion, "1.8") { // 1.8.4
			selectedVersion = QueryClientV1
		}
	}

	zap.L().Debug("selected query client", zap.String("query client version", selectedVersion))

	// update query client
	if selectedVersion == QueryClientV1 {
		c.queryClient = queryV1.InitClientV1(cfg.URI, cfg.InsecureSkipVerify, cfg.Bucket, cfg.Username, cfg.Password)
	} else {
		c.queryClient = queryV2.InitQueryV2(iClient.QueryAPI(cfg.Organization), cfg.Bucket)
	}
	return c, nil
}

// Ping to target database
func (c *Client) Ping() error {
	health, err := c.Client.Health(ctx)
	if err != nil {
		zap.L().Error("error on getting ready status", zap.Error(err))
		return err
	}
	zap.L().Debug("health status", zap.Any("health", health))
	return nil
}

// Close the influxdb connection
func (c *Client) Close() error {
	// close bulk insert
	close(c.stop)
	c.Client.Close()
	return nil
}

// Query func implementation
func (c *Client) Query(queryConfig *mtsml.QueryConfig) (map[string][]mtsml.ResponseData, error) {
	metricsMap := make(map[string][]mtsml.ResponseData)

	// fetch metrics details for the given input
	for _, q := range queryConfig.Individual {
		// clone global config
		query := queryConfig.Global.Clone()
		// update individual config
		query.Merge(&q)

		// add range
		if query.Start == "" {
			query.Start = queryML.DefaultStart
		}

		// add aggregateWindow
		if query.Window == "" {
			query.Window = queryML.DefaultWindow
		}

		// get measurement
		measurement, err := getMeasurementName(query.MetricType)
		if err != nil {
			zap.L().Error("error on getting measurement name", zap.Error(err))
			return nil, err
		}

		// execute query
		metrics, err := c.queryClient.ExecuteQuery(&query, measurement)
		if err != nil {
			return metricsMap, err
		}
		metricsMap[q.Name] = metrics
	}

	return metricsMap, nil
}

// WriteBlocking implementation
func (c *Client) WriteBlocking(data *mtsml.InputData) error {
	if data.MetricType == mtsml.MetricTypeNone {
		return nil
	}
	p, err := getPoint(data)
	if err != nil {
		return err
	}
	wb := c.Client.WriteAPIBlocking(c.Config.Organization, c.Config.Bucket)
	return wb.WritePoint(ctx, p)
}

func (c *Client) Write(data *mtsml.InputData) error {
	if data.MetricType == mtsml.MetricTypeNone {
		return nil
	}
	p, err := getPoint(data)
	if err != nil {
		return err
	}
	w := c.Client.WriteAPI(c.Config.Organization, c.Config.Bucket)
	w.WritePoint(p)
	return nil
}

func getPoint(data *mtsml.InputData) (*write.Point, error) {
	measurementName, err := getMeasurementName(data.MetricType)
	if err != nil {
		return nil, err
	}
	// convert tags to lowercase
	tags := make(map[string]string)
	for name, value := range data.Tags {
		formattedName := strings.ToLower(name)
		tags[formattedName] = value
	}

	p := influxdb2.NewPoint(
		measurementName,
		tags,
		data.Fields,
		data.Time,
	)
	return p, nil
}

func getMeasurementName(metricType string) (string, error) {
	switch metricType {
	case mtsml.MetricTypeBinary:
		return MeasurementBinary, nil

	case mtsml.MetricTypeGauge:
		return MeasurementGaugeInteger, nil

	case mtsml.MetricTypeGaugeFloat:
		return MeasurementGaugeFloat, nil

	case mtsml.MetricTypeCounter:
		return MeasurementCounter, nil

	case mtsml.MetricTypeString:
		return MeasurementString, nil

	case mtsml.MetricTypeGEO:
		return MeasurementGeo, nil

	default:
		return "", fmt.Errorf("unknown metric type: %s", metricType)
	}
}
