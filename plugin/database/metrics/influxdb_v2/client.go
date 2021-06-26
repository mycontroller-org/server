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
	fieldML "github.com/mycontroller-org/server/v2/pkg/model/field"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	metricsML "github.com/mycontroller-org/server/v2/plugin/database/metrics"
	extraML "github.com/mycontroller-org/server/v2/plugin/database/metrics/influxdb_v2/extra"
	extraV1 "github.com/mycontroller-org/server/v2/plugin/database/metrics/influxdb_v2/extra_v1"
	extraV2 "github.com/mycontroller-org/server/v2/plugin/database/metrics/influxdb_v2/extra_v2"
	"go.uber.org/zap"
)

var ctx = context.TODO()

const (
	defaultFlushInterval = 1 * time.Second
)

// Config of the influxdb_v2
type Config struct {
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
	queryClient extraML.QueryAPI
	adminClient extraML.AdminAPI
	Config      Config
	stop        chan bool
	buffer      []*fieldML.Field
	logger      *myLogger
	mutex       *sync.RWMutex
	ctx         context.Context
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

// NewClient of influxdb
func NewClient(config map[string]interface{}) (*Client, error) {
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
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
		buffer: make([]*fieldML.Field, 0),
		stop:   make(chan bool),
		mutex:  &sync.RWMutex{},
		logger: _logger,
		ctx:    context.TODO(),
	}

	err = c.Ping()
	if err != nil {
		return nil, err
	}

	influxAutoDetectVersion := ""

	if influxAutoDetectVersion == "" {
		influxAutoDetectVersion = QueryClientV1 // update default route, if non works
		// get version
		health, err := c.Client.Health(ctx)
		if err != nil {
			return nil, err
		}
		zap.L().Info("influxdb detected version data", zap.Any("health data", health))
		if health != nil && health.Version != nil {
			detectedVersion := *health.Version
			if strings.HasPrefix(detectedVersion, "1.8") { // 1.8.4
				influxAutoDetectVersion = QueryClientV1
			}
		}
	}

	zap.L().Debug("influx auto detect status", zap.String("version", influxAutoDetectVersion))

	// update admin client
	if influxAutoDetectVersion == QueryClientV1 {
		c.adminClient = extraV1.NewAdminClient(cfg.URI, cfg.InsecureSkipVerify, cfg.Bucket, cfg.Username, cfg.Password)
	} else {
		c.adminClient = extraV2.NewAdminClient(c.ctx, iClient.BucketsAPI(), cfg.Organization, cfg.Bucket)
	}

	// update autodetect version to user version, if user forced
	if cfg.QueryClientVersion != "" {
		ver := strings.ToLower(cfg.QueryClientVersion)
		if ver == QueryClientV1 || ver == QueryClientV2 {
			influxAutoDetectVersion = ver
		} else {
			zap.L().Warn("invalid query client version, going with auto detection", zap.String("input", cfg.QueryClientVersion))
		}
	}

	// update query client
	if influxAutoDetectVersion == QueryClientV1 {
		c.queryClient = extraV1.NewQueryClient(cfg.URI, cfg.InsecureSkipVerify, cfg.Bucket, cfg.Username, cfg.Password)
	} else {
		c.queryClient = extraV2.NewQueryClient(iClient.QueryAPI(cfg.Organization), cfg.Bucket)
	}

	// create bucket/database
	err = c.adminClient.CreateBucket()
	if err != nil {
		return nil, err
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
func (c *Client) Query(queryConfig *metricsML.QueryConfig) (map[string][]metricsML.ResponseData, error) {
	metricsMap := make(map[string][]metricsML.ResponseData)

	// fetch metrics details for the given input
	for _, q := range queryConfig.Individual {
		// clone global config
		query := queryConfig.Global.Clone()
		// update individual config
		query.Merge(&q)

		// add range
		if query.Start == "" {
			query.Start = extraML.DefaultStart
		}

		// add aggregateWindow
		if query.Window == "" {
			query.Window = extraML.DefaultWindow
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
func (c *Client) WriteBlocking(data *metricsML.InputData) error {
	if data.MetricType == metricsML.MetricTypeNone {
		return nil
	}
	p, err := getPoint(data)
	if err != nil {
		return err
	}
	wb := c.Client.WriteAPIBlocking(c.Config.Organization, c.Config.Bucket)
	return wb.WritePoint(ctx, p)
}

func (c *Client) Write(data *metricsML.InputData) error {
	if data.MetricType == metricsML.MetricTypeNone {
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

func getPoint(data *metricsML.InputData) (*write.Point, error) {
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
	case metricsML.MetricTypeBinary:
		return MeasurementBinary, nil

	case metricsML.MetricTypeGauge:
		return MeasurementGaugeInteger, nil

	case metricsML.MetricTypeGaugeFloat:
		return MeasurementGaugeFloat, nil

	case metricsML.MetricTypeCounter:
		return MeasurementCounter, nil

	case metricsML.MetricTypeString:
		return MeasurementString, nil

	case metricsML.MetricTypeGEO:
		return MeasurementGeo, nil

	default:
		return "", fmt.Errorf("unknown metric type: %s", metricType)
	}
}
