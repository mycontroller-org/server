package gatewaymessageprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	"go.uber.org/zap"
)

func writeFieldMetric(field *fieldTY.Field) error {
	fields := make(map[string]interface{})
	// update fields
	if field.MetricType == metricTY.MetricTypeGEO {
		_f, err := geoData(field.Current.Value)
		if err != nil {
			return err
		}
		fields = _f
	} else {
		fields[metricTY.FieldValue] = field.Current.Value
	}
	// update tags
	tags := map[string]string{
		types.KeyID:        field.ID,
		types.KeyGatewayID: field.GatewayID,
		types.KeyNodeID:    field.NodeID,
		types.KeySourceID:  field.SourceID,
		types.KeyFieldID:   field.FieldID,
	}

	// return data
	metricData := &metricTY.InputData{
		MetricType: field.MetricType,
		Time:       field.Current.Timestamp,
		Tags:       tags,
		Fields:     fields,
	}

	return writeMetric(metricData)
}

func geoData(pl interface{}) (map[string]interface{}, error) {
	// payload should be in this format
	// latitude;longitude;altitude. E.g. "55.722526;13.017972;18"
	d := make(map[string]interface{})
	ds := strings.Split(pl.(string), ";")
	if len(ds) < 2 {
		return nil, fmt.Errorf("invalid geo data: %s", pl)
	}
	lat, err := strconv.ParseFloat(ds[0], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float data: %s", pl)
	}
	lon, err := strconv.ParseFloat(ds[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float data: %s", pl)
	}
	alt := float64(0)
	if len(ds[0]) > 2 {
		alt, err = strconv.ParseFloat(ds[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float data: %s", pl)
		}
	}

	d[metricTY.FieldLatitude] = lat
	d[metricTY.FieldLongitude] = lon
	d[metricTY.FieldAltitude] = alt

	return d, nil
}

func writeNodeMetric(node *nodeTY.Node, suppliedMetricType, fieldName string, value interface{}) error {
	fields := make(map[string]interface{})
	// update fields
	fields[metricTY.FieldValue] = value

	// update tags
	tags := map[string]string{
		types.KeyID:        node.ID,
		types.KeyGatewayID: node.GatewayID,
		types.KeyNodeID:    node.NodeID,
		types.KeyFieldName: fieldName,
	}

	// return data
	metricData := &metricTY.InputData{
		MetricType: suppliedMetricType,
		Time:       node.LastSeen,
		Tags:       tags,
		Fields:     fields,
	}

	return writeMetric(metricData)
}

func writeMetric(metricData *metricTY.InputData) error {
	startTime := time.Now()
	err := store.METRIC.Write(metricData)
	if err != nil {
		zap.L().Error("failed to write into metrics database", zap.Error(err), zap.Any("metricData", metricData))
		return err
	}
	zap.L().Debug("inserted in to metric db", zap.String("timeTaken", time.Since(startTime).String()))
	return nil
}
