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
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	extraTY "github.com/mycontroller-org/server/v2/plugin/database/metric/influxdb_v2/extra"
	extraV1 "github.com/mycontroller-org/server/v2/plugin/database/metric/influxdb_v2/extra_v1"
	extraV2 "github.com/mycontroller-org/server/v2/plugin/database/metric/influxdb_v2/extra_v2"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"go.uber.org/zap"
)

var ctx = context.TODO()

// global constants
const (
	PluginInfluxdbV2 = "influxdb"

	MeasurementBinary       = "binary_data"
	MeasurementGaugeInteger = "gauge_int_data"
	MeasurementGaugeFloat   = "gauge_float_data"
	MeasurementCounter      = "counter_data"
	MeasurementString       = "string_data"
	MeasurementGeo          = "geo_data"

	QueryClientV1 = "v1"
	QueryClientV2 = "v2"

	DefaultMeasurementPrefix = "mc"
	loggerName               = "metric_influxdb"
)

const (
	defaultFlushInterval  = 1 * time.Second
	defaultConnectTimeout = 10 * time.Second
)

// Config of the influxdb_v2
type Config struct {
	IsCloudInstance    bool   `yaml:"is_cloud_instance"`
	OrganizationName   string `yaml:"organization_name"`
	BucketName         string `yaml:"bucket_name"`
	MeasurementPrefix  string `yaml:"measurement_prefix"`
	URI                string `yaml:"uri"`
	Token              string `yaml:"token"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Insecure           bool   `yaml:"insecure"`
	QueryClientVersion string `yaml:"query_client_version"`
	FlushInterval      string `yaml:"flush_interval"`
	ConnectTimeout     string `yaml:"connect_timeout"`
}

// LoggerConfig struct
type LoggerConfig struct {
	Mode             string `yaml:"mode"`
	Encoding         string `yaml:"encoding"`
	Level            string `yaml:"level"`
	EnableStacktrace bool   `yaml:"enable_stacktrace"`
}

// Client of the influxdb
type Client struct {
	Client               influxdb2.Client
	queryClient          extraTY.QueryAPI
	adminClient          extraTY.AdminAPI
	Config               Config
	stop                 chan bool
	mutex                *sync.RWMutex
	ctx                  context.Context
	influxInternalLogger *myLogger
	logger               *zap.Logger
}

// client of influxdb
func NewClient(ctx context.Context, config cmap.CustomMap) (metricTY.Plugin, error) {
	// create new logger with custom log level
	logger := metricTY.GetMetricLogger()

	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	namedLogger := logger.Named(loggerName)

	// update default values
	if cfg.MeasurementPrefix == "" {
		cfg.MeasurementPrefix = DefaultMeasurementPrefix
	}

	if cfg.Token == "" && cfg.Username != "" {
		cfg.Token = fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
	}
	flushInterval := defaultFlushInterval

	if cfg.FlushInterval != "" {
		flushInterval, err = time.ParseDuration(cfg.FlushInterval)
		if err != nil {
			namedLogger.Warn("Invalid flush interval", zap.String("flushInterval", cfg.FlushInterval))
			flushInterval = defaultFlushInterval
		}
	}
	if flushInterval.Milliseconds() < 1 {
		namedLogger.Warn("Minimum supported flush interval is 1ms, switching back to default", zap.String("flushInterval", cfg.FlushInterval))
		flushInterval = defaultFlushInterval
	}

	connectTimeout := utils.ToDuration(cfg.ConnectTimeout, defaultConnectTimeout)
	namedLogger.Debug("connect timeout details", zap.String("connectTimeout", cfg.ConnectTimeout), zap.Duration("connectTimeoutDuration", connectTimeout))

	// replace influxdb2 logger with our custom logger
	_logger := getLogger(namedLogger)
	influxdb2log.Log = _logger

	opts := influxdb2.DefaultOptions()
	opts.SetApplicationName("MyController.org")
	opts.SetFlushInterval(uint(flushInterval.Milliseconds()))
	iClient := influxdb2.NewClientWithOptions(cfg.URI, cfg.Token, opts)

	c := &Client{
		Config:               cfg,
		Client:               iClient,
		stop:                 make(chan bool),
		mutex:                &sync.RWMutex{},
		ctx:                  ctx,
		influxInternalLogger: _logger,
		logger:               namedLogger,
	}

	// repeat ping in 5 seconds interval, until timeout
	pingInterval := 5 * time.Second
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(connectTimeout):
			return nil, fmt.Errorf("influxdb connection timeout after %v", connectTimeout)
		case <-ticker.C:
			err = c.Ping()
			if err == nil {
				// ping successful, continue
				c.logger.Debug("influxdb ping successful")
				goto pingSuccess
			}
			c.logger.Warn("influxdb ping failed, retrying...", zap.Error(err))
		}
	}

pingSuccess:

	influxAutoDetectVersion := ""

	if influxAutoDetectVersion == "" {
		// update default route, if non works
		if cfg.OrganizationName != "" {
			influxAutoDetectVersion = QueryClientV2
		} else {
			influxAutoDetectVersion = QueryClientV1
		}

		// cloud instance always 2.x version and will not support for health check
		// https://github.com/influxdata/influxdb-client-go/pull/285
		if cfg.IsCloudInstance {
			influxAutoDetectVersion = QueryClientV2
		} else {
			// get version details with health call
			health, err := c.Client.Health(ctx)
			if err != nil {
				return nil, err
			}
			c.logger.Info("influxdb detected version data", zap.Any("health data", health))
			if health != nil && health.Version != nil {
				detectedVersion := *health.Version
				if strings.HasPrefix(detectedVersion, "1.8") { // 1.8.4
					influxAutoDetectVersion = QueryClientV1
				}
			}
		}
	}

	c.logger.Debug("influx auto detect status", zap.String("version", influxAutoDetectVersion))

	// update admin client
	if influxAutoDetectVersion == QueryClientV1 {
		c.adminClient = extraV1.NewAdminClient(c.logger, cfg.URI, cfg.Insecure, cfg.BucketName, cfg.Username, cfg.Password)
	} else {
		c.adminClient = extraV2.NewAdminClient(c.ctx, c.logger, iClient, cfg.OrganizationName, cfg.BucketName)
	}

	// update autodetect version to user version, if user forced
	if cfg.QueryClientVersion != "" {
		ver := strings.ToLower(cfg.QueryClientVersion)
		if ver == QueryClientV1 || ver == QueryClientV2 {
			influxAutoDetectVersion = ver
		} else {
			c.logger.Warn("invalid query client version, going with auto detection", zap.String("input", cfg.QueryClientVersion))
		}
	}

	// update query client
	if influxAutoDetectVersion == QueryClientV1 {
		c.queryClient = extraV1.NewQueryClient(c.logger, cfg.URI, cfg.Insecure, cfg.BucketName, cfg.Username, cfg.Password)
	} else {
		c.queryClient = extraV2.NewQueryClient(c.logger, iClient.QueryAPI(cfg.OrganizationName), cfg.BucketName)
	}

	// create bucket/database
	err = c.adminClient.CreateBucket()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Name() string {
	return PluginInfluxdbV2
}

// Ping to target database
func (c *Client) Ping() error {
	isReachable, err := c.Client.Ping(ctx)
	if err != nil {
		c.logger.Error("error on getting ping status", zap.Error(err))
		return err
	}
	c.logger.Debug("ping status", zap.Bool("isReachable", isReachable))
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
func (c *Client) Query(queryConfig *metricTY.QueryConfig) (map[string][]metricTY.ResponseData, error) {
	metricsMap := make(map[string][]metricTY.ResponseData)

	// fetch metrics details for the given input
	for _, q := range queryConfig.Individual {
		// clone global config
		query := queryConfig.Global.Clone()
		// update individual config
		query.Merge(&q)

		// add range
		if query.Start == "" {
			query.Start = extraTY.DefaultStart
		}

		// add aggregateWindow
		if query.Window == "" {
			query.Window = extraTY.DefaultWindow
		}

		// get measurement
		measurement, err := c.getMeasurementName(query.MetricType)
		if err != nil {
			c.logger.Error("error on getting measurement name", zap.Error(err))
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
func (c *Client) WriteBlocking(data *metricTY.InputData) error {
	if data.MetricType == metricTY.MetricTypeNone {
		return nil
	}
	p, err := c.getPoint(data)
	if err != nil {
		return err
	}
	wb := c.Client.WriteAPIBlocking(c.Config.OrganizationName, c.Config.BucketName)
	return wb.WritePoint(ctx, p)
}

func (c *Client) Write(data *metricTY.InputData) error {
	if data.MetricType == metricTY.MetricTypeNone {
		return nil
	}
	p, err := c.getPoint(data)
	if err != nil {
		return err
	}
	w := c.Client.WriteAPI(c.Config.OrganizationName, c.Config.BucketName)
	w.WritePoint(p)
	return nil
}

func (c *Client) getPoint(data *metricTY.InputData) (*write.Point, error) {
	measurementName, err := c.getMeasurementName(data.MetricType)
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

func (c *Client) getMeasurementName(suppliedType string) (string, error) {
	measurement := ""
	switch suppliedType {
	case metricTY.MetricTypeBinary:
		measurement = MeasurementBinary

	case metricTY.MetricTypeGauge:
		measurement = MeasurementGaugeInteger

	case metricTY.MetricTypeGaugeFloat:
		measurement = MeasurementGaugeFloat

	case metricTY.MetricTypeCounter:
		measurement = MeasurementCounter

	case metricTY.MetricTypeString:
		measurement = MeasurementString

	case metricTY.MetricTypeGEO:
		measurement = MeasurementGeo

	default:
		return "", fmt.Errorf("unknown metric type: %s", suppliedType)
	}

	return fmt.Sprintf("%s_%s", c.Config.MeasurementPrefix, measurement), nil
}
