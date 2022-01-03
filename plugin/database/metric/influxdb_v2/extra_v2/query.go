package extrav2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb-client-go/v2/api"
	queryTY "github.com/mycontroller-org/server/v2/plugin/database/metric/influxdb_v2/extra"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	"go.uber.org/zap"
)

// QueryV2 struct
type QueryV2 struct {
	api    api.QueryAPI
	bucket string
}

func NewQueryClient(api api.QueryAPI, bucket string) *QueryV2 {
	return &QueryV2{api: api, bucket: bucket}
}

func (qv2 *QueryV2) filter(name, value string) string {
	name = strings.ToLower(name)
	return fmt.Sprintf(` |> filter(fn: (r) => r["%s"] == "%s")`, name, value)
}

func (qv2 *QueryV2) aggregateWindowFunc(name, window string) (string, string) {
	return name, fmt.Sprintf(`
		%[1]s = data 
			|> aggregateWindow(every: %[2]s, fn: %[1]s) 
			|> set(key: "aggregation_type", value: "%[1]s")
			|> toFloat()`, name, window)
}

func (qv2 *QueryV2) aggregateWindowPercentileFunc(percentile, window string) (string, string) {
	p := float64(0.99)
	tmp := strings.SplitN(percentile, "_", 2)
	if len(tmp) == 2 {
		_p, err := strconv.ParseFloat(tmp[1], 64)
		if err == nil {
			p = _p / 100.0
		}
	}
	fnName := fmt.Sprintf("percentile_%02d", int64(p*100))
	return fnName, fmt.Sprintf(`
		%[1]s = data 
			|> aggregateWindow(every: %[2]s, fn: (column, tables=<-) => tables 
			|> quantile(q: %.02[3]f, column: column, method: "estimate_tdigest", compression: 1000.0)) 
			|> set(key: "aggregation_type", value: "%[1]s")
			|> toFloat()`, fnName, window, p)
}

func (qv2 *QueryV2) union(name string, fns []string) string {
	if len(fns) == 0 {
		return ""
	}
	finalData := fns[0]
	if len(fns) > 1 {
		finalData = fmt.Sprintf("union(tables: [%s])", strings.Join(fns, ","))
	}
	return fmt.Sprintf(`
		aColumns = ["_time", "median", "mean", "sum", "count", "min", "max"]
		%s
			|> pivot(rowKey:["_time"], columnKey: ["aggregation_type"], valueColumn: "_value")
			|> drop(fn: (column) => not contains(value: column, set: aColumns) and not column =~ /percentile*/)
			|> yield(name: "%s")`, finalData, name)
}

func (qv2 *QueryV2) buildQuery(suppliedMetricType, name, bucket, start, stop, window string, filters map[string]string, functions []string) string {
	// add bucket
	query := fmt.Sprintf(`data = from(bucket: "%s")`, bucket)

	// add range
	if stop != "" {
		query += fmt.Sprintf(` |> range(start: %s, stop: %s)`, start, stop)
	} else {
		query += fmt.Sprintf(` |> range(start: %s)`, start)
	}

	// add filters
	for n, v := range filters {
		query += qv2.filter(n, v)
	}

	// convert to float
	query += `  |> toFloat()`

	switch suppliedMetricType {
	case metricTY.MetricTypeGaugeFloat, metricTY.MetricTypeGauge:
		// add default functions, if none available
		if len(functions) == 0 {
			functions = []string{"mean", "min", "max"}
		}
		// add functions
		fns := make([]string, 0)
		// add function definitions
		for _, fn := range functions {
			fn = strings.ToLower(fn)
			var fnName, definition string
			if strings.HasPrefix(fn, "percentile") {
				fnName, definition = qv2.aggregateWindowPercentileFunc(fn, window)
			} else {
				fnName, definition = qv2.aggregateWindowFunc(fn, window)
			}
			fns = append(fns, fnName)
			query += definition
		}
		// add union
		query += qv2.union(name, fns)

	case metricTY.MetricTypeBinary:
		query += fmt.Sprintf(
			`data
				|> toFloat()
				|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
				|> drop(fn: (column) => not contains(value: column, set: ["_time", "%[1]s"]))
				|> yield(name: "%[2]s")`, queryTY.FieldValue, name)
	}

	// final query
	return query
}

func (qv2 *QueryV2) ExecuteQuery(q *metricTY.Query, measurement string) ([]metricTY.ResponseData, error) {
	filters := make(map[string]string)

	// add measurement
	filters["_measurement"] = measurement

	// add tags
	for k, v := range q.Tags {
		filters[k] = v
	}

	// add field value
	filters["_field"] = queryTY.FieldValue

	query := qv2.buildQuery(q.MetricType, q.Name, qv2.bucket, q.Start, q.Stop, q.Window, filters, q.Functions)

	zap.L().Debug("query", zap.String("query", query))

	tableResult, err := qv2.api.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	metrics := make([]metricTY.ResponseData, 0)

	// Use Next() to iterate over query result lines
	for tableResult.Next() {
		// Observe when there is new grouping key producing new table
		// if tableResult.TableChanged() {
		// 	fmt.Printf("table: %s\n", tableResult.TableMetadata().String())
		// }

		// read result
		record := tableResult.Record()
		//fmt.Printf("row: %s\n", record.String())

		_metric := make(map[string]interface{})

		for k, v := range record.Values() {
			switch k {
			case "_time", "result", "table":
				// do not do anything
			default:
				_metric[k] = v
			}
		}

		metrics = append(metrics, metricTY.ResponseData{Time: record.Time(), MetricType: q.MetricType, Metric: _metric})

		if tableResult.Err() != nil {
			zap.L().Error("Query error", zap.String("name", q.Name), zap.Error(tableResult.Err()))
		}
	}

	//zap.L().Debug("query response", zap.String("query", query), zap.Any("result", metrics))
	return metrics, nil
}
