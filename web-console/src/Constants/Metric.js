export const MetricType = {
  None: "none",
  String: "string",
  Binary: "binary",
  Gauge: "gauge",
  GaugeFloat: "gauge_float",
  Counter: "counter",
  GEO: "geo",
}

export const MetricTypeOptions = [
  { value: MetricType.None, label: "opts.metric_type.label_none" },
  { value: MetricType.String, label: "opts.metric_type.label_string" },
  { value: MetricType.Binary, label: "opts.metric_type.label_binary" },
  { value: MetricType.Gauge, label: "opts.metric_type.label_gauge" },
  { value: MetricType.GaugeFloat, label: "opts.metric_type.label_gauge_float" },
  { value: MetricType.Counter, label: "opts.metric_type.label_counter" },
  { value: MetricType.GEO, label: "opts.metric_type.label_geo" },
]

export const Duration = {
  Last30Minutes: "-30m",
  LastHour: "-1h",
  Last2Hours: "-2h",
  Last3Hours: "-3h",
  Last6Hours: "-6h",
  Last12Hours: "-12h",
  Last24Hours: "-24h",
  Last2Days: "-48h",
  Last7Days: "-168h",
  Last30Days: "-720h",
  Last60Days: "-1440h",
  Last90Days: "-2160h",
  Last180Days: "-4320h",
  Last365Days: "-8760h",
}

export const DurationOptions = [
  { value: Duration.Last30Minutes, tsFormat: "HH:mm", label: "opts.metric_duration.label_minute_30" },
  { value: Duration.LastHour, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_01" },
  { value: Duration.Last2Hours, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_02" },
  { value: Duration.Last3Hours, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_03" },
  { value: Duration.Last6Hours, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_06" },
  { value: Duration.Last12Hours, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_12" },
  { value: Duration.Last24Hours, tsFormat: "HH:mm", label: "opts.metric_duration.label_hour_24" },
  { value: Duration.Last2Days, tsFormat: "Do HH:mm", label: "opts.metric_duration.label_day_002" },
  { value: Duration.Last7Days, tsFormat: "Do HH:mm", label: "opts.metric_duration.label_day_007" },
  { value: Duration.Last30Days, tsFormat: "MMM Do HH:mm", label: "opts.metric_duration.label_day_030" },
  { value: Duration.Last60Days, tsFormat: "MMM Do HH:mm", label: "opts.metric_duration.label_day_060" },
  { value: Duration.Last90Days, tsFormat: "MMM Do HH:mm", label: "opts.metric_duration.label_day_090" },
  { value: Duration.Last180Days, tsFormat: "MMM Do HH:mm", label: "opts.metric_duration.label_day_180" },
  { value: Duration.Last365Days, tsFormat: "YYYY MMM Do HH:mm", label: "opts.metric_duration.label_day_365" },
]

export const AggregationInterval = {
  Minute_1: "1m",
  Minute_2: "2m",
  Minute_3: "3m",
  Minute_5: "5m",
  Minute_10: "10m",
  Minute_15: "15m",
  Minute_30: "30m",
  Hour_1: "1h",
  Hour_3: "3h",
  Hour_6: "6h",
  Day_1: "24h",
}

export const AggregationIntervalOptions = [
  { value: AggregationInterval.Minute_1, label: "opts.interval.label_minute_01" },
  { value: AggregationInterval.Minute_2, label: "opts.interval.label_minute_02" },
  { value: AggregationInterval.Minute_3, label: "opts.interval.label_minute_03" },
  { value: AggregationInterval.Minute_5, label: "opts.interval.label_minute_05" },
  { value: AggregationInterval.Minute_10, label: "opts.interval.label_minute_10" },
  { value: AggregationInterval.Minute_15, label: "opts.interval.label_minute_15" },
  { value: AggregationInterval.Minute_30, label: "opts.interval.label_minute_30" },
  { value: AggregationInterval.Hour_1, label: "opts.interval.label_hour_01" },
  { value: AggregationInterval.Hour_3, label: "opts.interval.label_hour_03" },
  { value: AggregationInterval.Hour_6, label: "opts.interval.label_hour_06" },
  { value: AggregationInterval.Day_1, label: "opts.interval.label_day_01" },
]

export const getRecommendedInterval = (durationValue) => {
  switch (durationValue) {
    case Duration.Last30Minutes:
    case Duration.LastHour:
    case Duration.Last2Hours:
      return AggregationInterval.Minute_1

    case Duration.Last3Hours:
      return AggregationInterval.Minute_2

    case Duration.Last6Hours:
      return AggregationInterval.Minute_3

    case Duration.Last12Hours:
      return AggregationInterval.Minute_5

    case Duration.Last24Hours:
      return AggregationInterval.Minute_10

    case Duration.Last2Days:
      return AggregationInterval.Minute_15

    case Duration.Last7Days:
      return AggregationInterval.Hour_1

    case Duration.Last30Days:
      return AggregationInterval.Hour_3

    case Duration.Last60Days:
    case Duration.Last90Days:
      return AggregationInterval.Hour_6

    case Duration.Last180Days:
    case Duration.Last365Days:
      return AggregationInterval.Day_1

    default:
      return AggregationInterval.Minute_15
  }
}

export const MetricFunctionType = {
  Mean: "mean",
  Max: "max",
  Min: "min",
  Median: "median",
  Sum: "sum",
  Count: "count",
  Percentile50: "percentile_50",
  Percentile75: "percentile_75",
  Percentile95: "percentile_95",
  Percentile99: "percentile_99",
}

export const MetricFunctionTypeOptions = [
  { value: MetricFunctionType.Mean, label: "opts.metric_function.label_mean" },
  { value: MetricFunctionType.Min, label: "opts.metric_function.label_minimum" },
  { value: MetricFunctionType.Max, label: "opts.metric_function.label_maximum" },
  { value: MetricFunctionType.Median, label: "opts.metric_function.label_median" },
  { value: MetricFunctionType.Percentile50, label: "opts.metric_function.label_percentile_50" },
  { value: MetricFunctionType.Percentile75, label: "opts.metric_function.label_percentile_75" },
  { value: MetricFunctionType.Percentile95, label: "opts.metric_function.label_percentile_95" },
  { value: MetricFunctionType.Percentile99, label: "opts.metric_function.label_percentile_99" },
  { value: MetricFunctionType.Sum, label: "opts.metric_function.label_sum" },
  { value: MetricFunctionType.Count, label: "opts.metric_function.label_count" },
]

export const InterpolationType = {
  Basis: "basis",
  Cardinal: "cardinal",
  CatmullRom: "catmullRom",
  Linear: "linear",
  MonotoneX: "monotoneX",
  MonotoneY: "monotoneY",
  Natural: "natural",
  Step: "step",
  StepAfter: "stepAfter",
  StepBefore: "stepBefore",
}

export const InterpolationTypeLineOptions = [
  { value: InterpolationType.Basis, label: "opts.chart_interpolation.label_basis" },
  { value: InterpolationType.Cardinal, label: "opts.chart_interpolation.label_cardinal" },
  { value: InterpolationType.CatmullRom, label: "opts.chart_interpolation.label_catmull_rom" },
  { value: InterpolationType.Linear, label: "opts.chart_interpolation.label_linear" },
  { value: InterpolationType.MonotoneX, label: "opts.chart_interpolation.label_monotone_x" },
  { value: InterpolationType.MonotoneY, label: "opts.chart_interpolation.label_monotone_y" },
  { value: InterpolationType.Natural, label: "opts.chart_interpolation.label_natural" },
  { value: InterpolationType.Step, label: "opts.chart_interpolation.label_step" },
  { value: InterpolationType.StepAfter, label: "opts.chart_interpolation.label_step_after" },
  { value: InterpolationType.StepBefore, label: "opts.chart_interpolation.label_step_before" },
]

export const RefreshIntervalType = {
  None: "0",
  Seconds_5: "5000",
  Seconds_10: "10000",
  Seconds_15: "15000",
  Seconds_30: "30000",
  Seconds_45: "45000",
  Minutes_1: "60000",
  Minutes_5: "300000",
  Minutes_15: "900000",
  Minutes_30: "1800000",
}

export const RefreshIntervalTypeOptions = [
  { value: RefreshIntervalType.None, label: "opts.interval.label_never" },
  { value: RefreshIntervalType.Seconds_5, label: "opts.interval.label_second_05" },
  { value: RefreshIntervalType.Seconds_10, label: "opts.interval.label_second_10" },
  { value: RefreshIntervalType.Seconds_15, label: "opts.interval.label_second_15" },
  { value: RefreshIntervalType.Seconds_30, label: "opts.interval.label_second_30" },
  { value: RefreshIntervalType.Seconds_45, label: "opts.interval.label_second_45" },
  { value: RefreshIntervalType.Minutes_1, label: "opts.interval.label_minute_01" },
  { value: RefreshIntervalType.Minutes_5, label: "opts.interval.label_minute_05" },
  { value: RefreshIntervalType.Minutes_15, label: "opts.interval.label_minute_15" },
  { value: RefreshIntervalType.Minutes_30, label: "opts.interval.label_minute_30" },
]

export const DataUnitType = {
  Bytes: "",
  KiB: "KiB",
  MiB: "MiB",
  GiB: "GiB",
  TiB: "TiB",
  PiB: "PiB",
  EiB: "EiB",
}

export const DataUnitTypeOptions = [
  { value: DataUnitType.Bytes, label: "opts.data_unit.label_bytes" },
  { value: DataUnitType.KiB, label: "opts.data_unit.label_kib" },
  { value: DataUnitType.MiB, label: "opts.data_unit.label_mib" },
  { value: DataUnitType.GiB, label: "opts.data_unit.label_gib" },
  { value: DataUnitType.TiB, label: "opts.data_unit.label_tib" },
  { value: DataUnitType.PiB, label: "opts.data_unit.label_pib" },
  { value: DataUnitType.EiB, label: "opts.data_unit.label_eib" },
]
