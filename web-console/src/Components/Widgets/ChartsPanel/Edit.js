import objectPath from "object-path"
import { DataType, FieldType } from "../../../Constants/Form"
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
} from "../../../Constants/Metric"
import { ResourceType, ResourceTypeOptions } from "../../../Constants/ResourcePicker"
import {
  ChartGroupType,
  ChartGroupTypeOptions,
  ChartType,
  ChartTypeOptions,
  LegendOrientationType,
  LegendOrientationTypeOptions,
  LegendPositionType,
  LegendPositionTypeOptions,
  ThemeColor,
  ThemeColorOptions,
} from "../../../Constants/Widgets/ChartsPanel"
import { getValue } from "../../../Util/Util"

// Charts Panel items
export const updateFormItemsChartsPanel = (rootObject, items = []) => {
  items.push({
    label: "sub_type",
    fieldId: "config.subType",
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    value: "",
    options: ChartGroupTypeOptions.filter((opt) => opt.value !== ChartGroupType.PieChart),
    isRequired: true,
    validator: { isNotEmpty: {} },
    resetFields: { "config.resource": {}, "config.resources": [] },
  })

  const selectedSubType = objectPath.get(rootObject, "config.subType", "")

  switch (selectedSubType) {
    case ChartGroupType.GroupChart:
    case ChartGroupType.MixedChart:
      const metricConfigItems = getMetricConfigItems(rootObject)
      items.push(...metricConfigItems)

      const globalConfigItems = getGlobalConfigItems(rootObject)
      items.push(...globalConfigItems)

      if (selectedSubType === ChartGroupType.MixedChart) {
        const axisConfigItems = getAxisYConfigItems(rootObject)
        items.push(...axisConfigItems)

        const resources = getResourceConfigItems(rootObject)
        items.push(...resources)
      } else {
        const axisConfigItems = getAxisYConfigItems(rootObject, true)
        items.push(...axisConfigItems)

        const resources = getGroupChartResourceConfigItems(rootObject)
        items.push(...resources)
      }

      const chartConfigItems = getChartConfigItems(rootObject)
      items.push(...chartConfigItems)

      break

    default:
    //noop
  }
}

const getGlobalConfigItems = (rootObject) => {
  // update default values
  objectPath.set(rootObject, "config.chart.fillOpacity", 15, true)
  objectPath.set(rootObject, "config.chart.strokeWidth", 1.5, true)
  objectPath.set(rootObject, "config.chart.roundDecimal", 0, true)
  objectPath.set(rootObject, "config.chart.interpolation", InterpolationType.Natural, true)

  const subType = getValue(rootObject, "config.subType", ChartGroupType.GroupChart)

  const items = [
    {
      label: "global_configuration",
      fieldId: "!global_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "fill_opacity_%",
      fieldId: "config.chart.fillOpacity",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 100,
      step: 1,
    },
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
      label: "round_decimal",
      fieldId: "config.chart.roundDecimal",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 0,
      max: 10,
      step: 1,
    },
    {
      label: "interpolation",
      fieldId: "config.chart.interpolation",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: InterpolationTypeLineOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
  ]

  if (subType === ChartGroupType.GroupChart) {
    objectPath.set(rootObject, "config.chart.chartType", ChartType.AreaChart, true)
    objectPath.set(rootObject, "config.chart.themeColor", ThemeColor.MultiOrdered, true)
    items.push(
      {
        label: "chart_type",
        fieldId: "config.chart.chartType",
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        options: ChartTypeOptions.filter((opt) => opt.value !== ChartType.PieChart),
        validator: { isNotEmpty: {} },
      },
      {
        label: "theme_color",
        fieldId: "config.chart.themeColor",
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        options: ThemeColorOptions,
        validator: { isNotEmpty: {} },
      }
    )
  }

  return items
}

const getMetricConfigItems = (rootObject) => {
  // update default values
  objectPath.set(rootObject, "config.chart.duration", Duration.Last6Hours, true)
  objectPath.set(
    rootObject,
    "config.chart.interval",
    getRecommendedInterval(getValue(rootObject, "config.chart.duration", "")),
    true
  )
  objectPath.set(rootObject, "config.chart.metricFunction", MetricFunctionType.Percentile99, true)
  objectPath.set(rootObject, "config.chart.refreshInterval", RefreshIntervalType.None, true)

  const items = [
    {
      label: "metric_configuration",
      fieldId: "!metric_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "duration",
      fieldId: "config.chart.duration",
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
      fieldId: "config.chart.interval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: AggregationIntervalOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "metric_function",
      fieldId: "config.chart.metricFunction",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: MetricFunctionTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },

    {
      label: "refresh_interval",
      fieldId: "config.chart.refreshInterval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.Integer,
      value: "",
      options: RefreshIntervalTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
  ]

  return items
}

const getChartConfigItems = (rootObject) => {
  // update default values
  objectPath.set(rootObject, "config.chart.showGridX", false, true)
  objectPath.set(rootObject, "config.chart.showGridY", false, true)
  objectPath.set(rootObject, "config.chart.cursorTooltip", true, true)
  objectPath.set(rootObject, "config.chart.stackCharts", false, true)
  objectPath.set(rootObject, "config.chart.height", 250, true)
  objectPath.set(rootObject, "config.chart.legendPosition", LegendPositionType.BottomLeft, true)
  objectPath.set(rootObject, "config.chart.legendOrientation", LegendOrientationType.Horizontal, true)
  objectPath.set(rootObject, "config.chart.padding.top", 5, true)
  objectPath.set(rootObject, "config.chart.padding.right", 5, true)
  objectPath.set(rootObject, "config.chart.padding.bottom", 55, true)
  objectPath.set(rootObject, "config.chart.padding.left", 55, true)

  const items = [
    {
      label: "chart_configuration",
      fieldId: "!chart_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "show_grid_x",
      fieldId: "config.chart.showGridX",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "show_grid_y",
      fieldId: "config.chart.showGridY",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "cursor_tooltip",
      fieldId: "config.chart.cursorTooltip",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "stack_charts",
      fieldId: "config.chart.stackCharts",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "height_px",
      fieldId: "config.chart.height",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_height",
      validated: "default",
      validator: { isNotEmpty: {}, isInteger: { min: 1 } },
    },
    {
      label: "legend_position",
      fieldId: "config.chart.legendPosition",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: LegendPositionTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "legend_orientation",
      fieldId: "config.chart.legendOrientation",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: LegendOrientationTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "padding_top_px",
      fieldId: "config.chart.padding.top",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 250,
      step: 1,
    },
    {
      label: "padding_right_px",
      fieldId: "config.chart.padding.right",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 250,
      step: 1,
    },
    {
      label: "padding_bottom_px",
      fieldId: "config.chart.padding.bottom",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 250,
      step: 1,
    },
    {
      label: "padding_left_px",
      fieldId: "config.chart.padding.left",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 250,
      step: 1,
    },
  ]

  return items
}

const getAxisYConfigItems = (rootObject, isSingleAxis = false) => {
  updateYAxisConfig(rootObject, isSingleAxis)

  const items = [
    {
      label: "y_axis_configuration",
      fieldId: "!y_axis_configuration",
      fieldType: FieldType.Divider,
    },
  ]

  if (!isSingleAxis) {
    items.push({
      label: "number_of_axis",
      fieldId: "config.axisYCount",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 10,
      step: 1,
      onChangeFunc: updateYAxisConfig,
    })
  }
  items.push({
    label: "",
    fieldId: "config.axisY",
    fieldType: FieldType.ChartYAxisConfigMap,
    dataType: DataType.Object,
    value: "",
    saveButtonText: "update",
    updateButtonText: "update",
    language: "yaml",
    minimapEnabled: true,
    isRequired: false,
  })
  return items
}

const getResourceConfigItems = (rootObject) => {
  const yAxisConfig = getValue(rootObject, "config.axisY", {})
  const items = [
    {
      label: "resources",
      fieldId: "!resources",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "config.resources",
      fieldType: FieldType.ChartMixedResourceConfig,
      dataType: DataType.ArrayObject,
      value: "",
      saveButtonText: "update",
      updateButtonText: "update",
      language: "yaml",
      minimapEnabled: true,
      isRequired: false,
      yAxisConfig: yAxisConfig,
    },
  ]

  return items
}

const getGroupChartResourceConfigItems = (rootObject) => {
  objectPath.set(rootObject, "config.resource.type", ResourceType.Field, true)
  objectPath.set(rootObject, "config.resource.nameKey", "name", true)
  objectPath.set(rootObject, "config.resource.unit", "", true)
  objectPath.set(rootObject, "config.resource.limit", 5, true)
  objectPath.set(rootObject, "config.resource.filters", { "": "" }, true)

  const items = [
    {
      label: "resource_configuration",
      fieldId: "!resource_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "resource_type",
      fieldId: "config.resource.type",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      options: ResourceTypeOptions.filter(
        (r) => r.value === ResourceType.Node || r.value === ResourceType.Field
      ),
      validator: { isNotEmpty: {} },
    },
    {
      label: "filter",
      fieldId: "config.resource.filters",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
    },
    {
      label: "name_key",
      fieldId: "config.resource.nameKey",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_key",
      validated: "default",
      validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "unit",
      fieldId: "config.resource.unit",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "limit",
      fieldId: "config.resource.limit",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_limit",
      validated: "default",
      validator: { isInteger: {} },
    },
  ]
  return items
}

// helper functions

const axisColors = [
  "#4f5255", // 0  color: rgb(79, 82, 85)
  "#4f5255", // 1
  "#4f5255", // 2
  "#4f5255", // 3
  "#4f5255", // 4
  "#4f5255", // 5
  "#4f5255", // 6
  "#4f5255", // 7
  "#4f5255", // 8
  "#4f5255", // 9
]
const axisOffset = [0, 100, 10, 20, 30, 40, 50, 60, 70, 80]

const updateYAxisConfig = (rootObject, isSingleAxis = false) => {
  objectPath.set(rootObject, "config.axisYCount", 1, isSingleAxis ? false : true)
  objectPath.set(rootObject, "config.axisY", {}, true)

  const configMap = getValue(rootObject, "config.axisY", {})
  const axisYCount = getValue(rootObject, "config.axisYCount", 1)

  const keys = Object.keys(configMap).sort()
  let newMap = {}

  if (keys.length > axisYCount) {
    keys.forEach((key) => {
      if (key < axisYCount) {
        newMap[key] = { ...configMap[key] }
      }
    })
  } else if (keys.length < axisYCount) {
    for (let index = 0; index < axisYCount; index++) {
      if (index >= keys.length) {
        newMap[index] = {
          color: axisColors[index],
          offsetY: axisOffset[index],
          roundDecimal: 0,
          unit: "",
        }
      } else {
        newMap[index] = { ...configMap[index] }
      }
    }
  }

  if (Object.keys(newMap).length > 0) {
    objectPath.set(rootObject, "config.axisY", newMap, false)
  }
}
