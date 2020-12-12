package influx

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metrics"
	"go.uber.org/zap"
)

// global contants
const (
	DefaultWindow = "5m"
	DefaultStart  = "-1h"
)

func filter(name, value string) string {
	name = strings.ToLower(name)
	return fmt.Sprintf(` |> filter(fn: (r) => r["%s"] == "%s")`, name, value)
}

func aggregateWindowFunc(name, window string) (string, string) {
	return name, fmt.Sprintf(`
		%[1]s = data 
			|> aggregateWindow(every: %[2]s, fn: %[1]s) 
			|> set(key: "aggregation_type", value: "%[1]s")
			|> toFloat()`, name, window)
}

func aggregateWindowPercentileFunc(percentile, window string) (string, string) {
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

func union(name string, fns []string) string {
	return fmt.Sprintf(`
		aColumns = ["_time", "median", "mean", "sum", "count", "min", "max"]
		union(tables: [%s])
			|> pivot(rowKey:["_time"], columnKey: ["aggregation_type"], valueColumn: "_value")
			|> drop(fn: (column) => not contains(value: column, set: aColumns) and not column =~ /percentile*/)
			|> yield(name: "%s")`, strings.Join(fns, ","), name)
}

func buildQuery(metricType, name, bucket, start, stop, window string, filters map[string]string, functions []string) string {
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
		query += filter(n, v)
	}

	// convert to float
	query += `  |> toFloat()`

	switch metricType {
	case mtrml.MetricTypeGaugeFloat, mtrml.MetricTypeGauge:
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
				fnName, definition = aggregateWindowPercentileFunc(fn, window)
			} else {
				fnName, definition = aggregateWindowFunc(fn, window)
			}
			fns = append(fns, fnName)
			query += definition
		}
		// add union
		query += union(name, fns)

	case mtrml.MetricTypeBinary:
		query += fmt.Sprintf(
			`data
				|> toFloat()
				|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
				|> drop(fn: (column) => not contains(value: column, set: ["_time", "%[1]s"]))
				|> yield(name: "%[2]s")`, FieldValue, name)
	}

	// final query
	return query
}

func (c *Client) executeQuery(q mtrml.Query) ([]mtrml.Data, error) {
	// add range
	start := DefaultStart
	if q.Start != "" {
		start = q.Start
	}

	filters := make(map[string]string)

	// add measurement
	measurement, err := measurementName(q.MetricType)
	if err != nil {
		// do some action
	}
	filters["_measurement"] = measurement

	// add tags
	for k, v := range q.Tags {
		filters[k] = v
	}

	// add field value
	filters["_field"] = FieldValue

	// add aggregateWindow
	window := DefaultWindow
	if q.Window != "" {
		window = q.Window
	}

	query := buildQuery(q.MetricType, q.Name, c.Config.Bucket, start, q.Stop, window, filters, q.Functions)

	zap.L().Debug("query", zap.String("query", query))

	api := c.Client.QueryAPI(c.Config.Organization)
	tableResult, err := api.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	metrics := make([]mtrml.Data, 0)

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

		metrics = append(metrics, mtrml.Data{Time: record.Time(), MetricType: q.MetricType, Metric: _metric})

		if tableResult.Err() != nil {
			zap.L().Error("Query error", zap.String("name", q.Name), zap.Error(tableResult.Err()))
		}
	}

	//zap.L().Debug("query response", zap.String("query", query), zap.Any("result", metrics))
	return metrics, nil
}

// Query func implementation
func (c *Client) Query(queryConfig *mtrml.QueryConfig) (map[string][]mtrml.Data, error) {
	metricsMap := make(map[string][]mtrml.Data)

	// fetch metrics details for the given input
	for _, q := range queryConfig.Individual {
		// clone global config
		query := queryConfig.Global.Clone()
		// update individual config
		query.Merge(&q)
		// execute query
		metrics, err := c.executeQuery(query)
		if err != nil {
			return metricsMap, err
		}
		metricsMap[q.Name] = metrics
	}

	return metricsMap, nil
}
