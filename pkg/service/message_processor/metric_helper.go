package messageprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"go.uber.org/zap"
)

func writeFieldMetric(field *fml.Field) error {
	fields := make(map[string]interface{})
	// update fields
	if field.MetricType == mtsml.MetricTypeGEO {
		_f, err := geoData(field.Current.Value)
		if err != nil {
			return err
		}
		fields = _f
	} else {
		fields[mtsml.FieldValue] = field.Current.Value
	}
	// update tags
	tags := map[string]string{
		model.KeyID:        field.ID,
		model.KeyGatewayID: field.GatewayID,
		model.KeyNodeID:    field.NodeID,
		model.KeySourceID:  field.SourceID,
		model.KeyFieldID:   field.FieldID,
	}

	// return data
	metricData := &mtsml.InputData{
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

	d[mtsml.FieldLatitude] = lat
	d[mtsml.FieldLongitude] = lon
	d[mtsml.FieldAltitude] = alt

	return d, nil
}

func writeNodeMetric(node *nml.Node, metricType, fieldName string, value interface{}) error {
	fields := make(map[string]interface{})
	// update fields
	fields[mtsml.FieldValue] = value

	// update tags
	tags := map[string]string{
		model.KeyID:        node.ID,
		model.KeyGatewayID: node.GatewayID,
		model.KeyNodeID:    node.NodeID,
		model.KeyFieldName: fieldName,
	}

	// return data
	metricData := &mtsml.InputData{
		MetricType: metricType,
		Time:       node.LastSeen,
		Tags:       tags,
		Fields:     fields,
	}

	return writeMetric(metricData)
}

func writeMetric(metricData *mtsml.InputData) error {
	startTime := time.Now()
	err := mts.SVC.Write(metricData)
	if err != nil {
		zap.L().Error("failed to write into metrics database", zap.Error(err), zap.Any("metricData", metricData))
		return err
	}
	zap.L().Debug("inserted in to metric db", zap.String("timeTaken", time.Since(startTime).String()))
	return nil
}
