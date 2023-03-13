package extrav1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	cloneUtils "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"go.uber.org/zap"
)

// QueryV1 struct
type QueryV1 struct {
	client      *httpclient.Client
	headers     map[string]string
	url         string
	queryParams map[string]interface{}
	logger      *zap.Logger
}

func NewQueryClient(logger *zap.Logger, uri string, insecure bool, bucket, username, password string) *QueryV1 {
	headers, newClient := newClient(uri, insecure, username, password)

	queryParams := map[string]interface{}{
		"db":     bucket,
		"pretty": false,
		"epoch":  "ns",
	}

	return &QueryV1{
		client:      newClient,
		url:         fmt.Sprintf("%s/query", uri),
		headers:     headers,
		queryParams: queryParams,
		logger:      logger,
	}
}

func (qv1 *QueryV1) ExecuteQuery(query *metricType.Query, measurement string) ([]metricType.ResponseData, error) {
	queryParams, _ := cloneUtils.Clone(qv1.queryParams).(map[string]interface{})

	queryString := qv1.buildQuery(query, measurement)
	queryParams["q"] = queryString

	qv1.logger.Debug("input", zap.String("query", queryString))

	response, err := qv1.client.ExecuteJson(qv1.url, http.MethodGet, qv1.headers, queryParams, nil, 0)
	if err != nil {
		qv1.logger.Error("error on calling api", zap.Error(err))
		return nil, err
	}

	qv1.logger.Debug("response", zap.String("body", response.StringBody()), zap.Any("qp", queryParams))

	if response.StatusCode != http.StatusOK {
		// call error response
		return nil, fmt.Errorf("invalid status code:%v", response.StatusCode)
	}

	queryResult := QueryResult{}
	err = json.Unmarshal(response.Body, &queryResult)
	if err != nil {
		return nil, err
	}

	metrics := make([]metricType.ResponseData, 0)

	if queryResult.Error != "" {
		return nil, errors.New(queryResult.Error)
	}

	if len(queryResult.Results) == 0 {
		return metrics, nil
	}

	result := queryResult.Results[0]
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}

	if len(result.Series) == 0 {
		return metrics, nil
	}

	series := result.Series[0]
	for index := 0; index < len(series.Values); index++ {
		values := series.Values[index]
		if len(series.Columns) != len(values) {
			continue
		}
		_metric := make(map[string]interface{})
		var _time interface{}
		for vIndex := 0; vIndex < len(values); vIndex++ {
			column := series.Columns[vIndex]
			if column == "time" {
				_time = values[vIndex]
			} else {
				if query.MetricType == metricType.MetricTypeBinary && column == "value" {
					value := converterUtils.ToBool(values[vIndex])
					if value {
						_metric[column] = int64(1)
					} else {
						_metric[column] = int64(0)
					}
				} else {
					_metric[column] = values[vIndex]
				}
			}
		}

		timeNS := converterUtils.ToFloat(_time)
		finalTime := time.Unix(0, int64(timeNS))
		if finalTime.IsZero() {
			return nil, fmt.Errorf("invalid timestamp, type:%T, value:%v, query:%+v", _time, _time, query)
		}
		metrics = append(metrics, metricType.ResponseData{Time: finalTime, MetricType: query.MetricType, Metric: _metric})
	}

	return metrics, nil
}

func (qv1 *QueryV1) buildQuery(query *metricType.Query, measurement string) string {
	if len(query.Functions) == 0 {
		query.Functions = []string{"mean", "min", "max"}
	}

	// update functions
	functions := make([]string, len(query.Functions))
	for index, fn := range query.Functions {
		if strings.HasPrefix(fn, "percentile") {
			p := int64(99)
			tmp := strings.SplitN(fn, "_", 2)
			if len(tmp) == 2 {
				_p, err := strconv.ParseInt(tmp[1], 10, 64)
				if err == nil {
					p = _p
				}
			}
			functions[index] = fmt.Sprintf(`percentile("value", %02d) AS "%s"`, p, fn)
		} else {
			functions[index] = fmt.Sprintf(`%s("value")`, fn)
		}
	}

	// start building query from here
	var qBuilder strings.Builder
	qBuilder.WriteString("SELECT")

	switch query.MetricType {
	case metricType.MetricTypeGauge, metricType.MetricTypeGaugeFloat, metricType.MetricTypeCounter:
		for index, fn := range functions {
			if index != 0 {
				qBuilder.WriteByte(',')
			}
			fmt.Fprintf(&qBuilder, " %s", fn)
		}

	case metricType.MetricTypeBinary:
		fmt.Fprint(&qBuilder, ` "value"`)

	default:
		fmt.Fprint(&qBuilder, ` "value"`)

	}

	fmt.Fprintf(&qBuilder, ` FROM "%s"`, measurement)

	index := 0
	for tag, value := range query.Tags {
		if index == 0 {
			qBuilder.WriteString(" WHERE")
		} else {
			qBuilder.WriteString(" AND")
		}
		fmt.Fprintf(&qBuilder, ` "%s"='%s'`, tag, value)
		index++
	}

	if index == 0 {
		qBuilder.WriteString(" WHERE")
	} else {
		qBuilder.WriteString(" AND")
	}

	_, err := time.ParseDuration(query.Start)
	if err != nil {
		fmt.Fprintf(&qBuilder, " time >= '%s'", query.Start)
	} else {
		fmt.Fprintf(&qBuilder, " time >= now()%s", query.Start)
	}

	if query.Stop != "" {
		_, err := time.ParseDuration(query.Stop)
		if err != nil {
			fmt.Fprintf(&qBuilder, " AND time >= '%s'", query.Stop)
		} else {
			fmt.Fprintf(&qBuilder, " AND time >= now()%s", query.Stop)
		}
	} else {
		qBuilder.WriteString(" AND time <= now()")
	}

	switch query.MetricType {
	case metricType.MetricTypeGauge, metricType.MetricTypeGaugeFloat, metricType.MetricTypeCounter:
		fmt.Fprintf(&qBuilder, " GROUP BY time(%s) fill(null)", query.Window)
	}
	return qBuilder.String()
}
