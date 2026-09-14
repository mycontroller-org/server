import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import {
  AggregationIntervalOptions,
  Duration,
  DurationOptions,
  getRecommendedInterval,
  InterpolationType,
  InterpolationTypeLineOptions,
  MetricFunctionType,
  MetricFunctionTypeOptions,
  RefreshIntervalType,
  RefreshIntervalTypeOptions,
} from "../../../../Constants/Metric"
import { ColorsSetBig } from "../../../../Constants/Widgets/Color"
import { ChartType } from "../../../../Constants/Widgets/UtilizationPanel"
import { getValue } from "../../../../Util/Util"
import { getResourceItems } from "../Common/edit_resource"

// example data structure
// const data = {
//   type: "",
//   resource: {
//     type: "",
//     filterType: "", // quick_id, detailed_filter
//     quickId: "",
//     displayName: "",
//     nameKey: "",
//     roundDecimal: "",
//     valueKey: "",
//     timestampKey: "",
//     unit: "",
//     filter: {},
//   },
//   chart: {
//     interpolation: "basis",
//     height: 70,
//     color: "#0066cc",
//     fillOpacity: 15,
//     strokeWidth: 1.5,
//     thickness: 20,
//     yAxisMinValue: 12,
//     yAxisMaxValue: 13,
//   },
//   metric: {
//     duration: "-1h",
//     interval: "1m",
//     refreshInterval: 0,
//     metricFunction: "percentile_99",
//   },
// }

// SparkLine items
export const getSparkItems = (rootObject) => {
  // load default values
  objectPath.set(rootObject, "config.metric.duration", Duration.Last6Hours, true)
  objectPath.set(rootObject, "config.metric.interval", getRecommendedInterval(Duration.Last6Hours), true)
  objectPath.set(rootObject, "config.metric.refreshInterval", RefreshIntervalType.Minutes_5, true)
  objectPath.set(rootObject, "config.metric.metricFunction", MetricFunctionType.Percentile99, true)

  objectPath.set(rootObject, "config.chart.interpolation", InterpolationType.Basis, true)
  objectPath.set(rootObject, "config.chart.height", 70, true)
  objectPath.set(rootObject, "config.chart.color", "#0066cc", true)
  objectPath.set(rootObject, "config.chart.fillOpacity", 15, true)
  objectPath.set(rootObject, "config.chart.strokeWidth", 1.5, true)
  objectPath.set(rootObject, "config.chart.thickness", 20, true)
  objectPath.set(rootObject, "config.chart.yAxisMinValue", "", true)
  objectPath.set(rootObject, "config.chart.yAxisMaxValue", "", true)

  const items = []

  // include resource items
  const resourceItems = getResourceItems(rootObject)
  items.push(...resourceItems)

  // metric configuration
  items.push(
    {
      label: "metric_configuration",
      fieldId: "!metric_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "duration",
      fieldId: "config.metric.duration",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: DurationOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: {
        "config.chart.interval": (ro) => {
          return getRecommendedInterval(getValue(ro, "config.chart.duration", ""))
        },
      },
    },
    {
      label: "interval",
      fieldId: "config.metric.interval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: AggregationIntervalOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "metric_function",
      fieldId: "config.metric.metricFunction",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: MetricFunctionTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "refresh_interval",
      fieldId: "config.metric.refreshInterval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.Integer,
      value: "",
      options: RefreshIntervalTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    }
  )

  // chart configuration
  const chartType = objectPath.get(rootObject, "config.type", "")

  items.push({
    label: "chart_configuration",
    fieldId: "!chart_configuration",
    fieldType: FieldType.Divider,
  })

  if (chartType !== ChartType.SparkBar) {
    items.push({
      label: "interpolation",
      fieldId: "config.chart.interpolation",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: InterpolationTypeLineOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    })
  }

  items.push(
    {
      label: "height_px",
      fieldId: "config.chart.height",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_value",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
    {
      label: "color",
      fieldId: "config.chart.color",
      fieldType: FieldType.ColorBox,
      dataType: DataType.String,
      value: "",
      colors: ColorsSetBig,
    }
  )

  if (chartType !== ChartType.SparkLine) {
    items.push({
      label: "fill_opacity_%",
      fieldId: "config.chart.fillOpacity",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 100,
      step: 1,
    })
  }

  items.push(
    {
      label: "stroke_width_px",
      fieldId: "config.chart.strokeWidth",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Float,
      value: "",
      min: 0,
      max: 20,
      step: 0.5,
    },
    {
      label: "min_y_axis",
      fieldId: "config.chart.yAxisMinValue",
      fieldType: FieldType.Text,
      dataType: DataType.Float,
      value: "",
      isRequired: false,
    },
    {
      label: "max_y_axis",
      fieldId: "config.chart.yAxisMaxValue",
      fieldType: FieldType.Text,
      dataType: DataType.Float,
      value: "",
      isRequired: false,
    }
  )

  return items
}
