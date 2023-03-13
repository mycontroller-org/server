package gatewaymessageprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	metricPluginTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"go.uber.org/zap"
)

func (svc *MessageProcessor) writeFieldMetric(field *fieldTY.Field) error {
	fields := make(map[string]interface{})
	// update fields
	if field.MetricType == metricPluginTY.MetricTypeGEO {
		_f, err := svc.geoData(field.Current.Value)
		if err != nil {
			return err
		}
		fields = _f
	} else {
		fields[metricPluginTY.FieldValue] = field.Current.Value
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
	metricData := &metricPluginTY.InputData{
		MetricType: field.MetricType,
		Time:       field.Current.Timestamp,
		Tags:       tags,
		Fields:     fields,
	}

	return svc.writeMetric(metricData)
}

func (svc *MessageProcessor) geoData(pl interface{}) (map[string]interface{}, error) {
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

	d[metricPluginTY.FieldLatitude] = lat
	d[metricPluginTY.FieldLongitude] = lon
	d[metricPluginTY.FieldAltitude] = alt

	return d, nil
}

func (svc *MessageProcessor) writeNodeMetric(node *nodeTY.Node, suppliedMetricType, fieldName string, value interface{}) error {
	fields := make(map[string]interface{})
	// update fields
	fields[metricPluginTY.FieldValue] = value

	// update tags
	tags := map[string]string{
		types.KeyID:        node.ID,
		types.KeyGatewayID: node.GatewayID,
		types.KeyNodeID:    node.NodeID,
		types.KeyFieldName: fieldName,
	}

	// return data
	metricData := &metricPluginTY.InputData{
		MetricType: suppliedMetricType,
		Time:       node.LastSeen,
		Tags:       tags,
		Fields:     fields,
	}

	return svc.writeMetric(metricData)
}

func (svc *MessageProcessor) writeMetric(metricData *metricPluginTY.InputData) error {
	startTime := time.Now()
	err := svc.metric.Write(metricData)
	if err != nil {
		svc.logger.Error("failed to write into metrics database", zap.Error(err), zap.Any("metricData", metricData))
		return err
	}
	svc.logger.Debug("inserted in to metric db", zap.String("timeTaken", time.Since(startTime).String()))
	return nil
}
