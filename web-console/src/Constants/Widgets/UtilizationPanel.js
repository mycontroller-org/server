// utilization panel chart types
export const ChartType = {
  CircleSize50: "circle_size_50",
  CircleSize75: "circle_size_75",
  CircleSize100: "circle_size_100",
  SparkArea: "spark_area",
  SparkLine: "spark_line",
  SparkBar: "spark_bar",
  Table: "table",
}

// utilization panel chart type options
export const ChartTypeOptions = [
  { value: ChartType.CircleSize50, label: "opts.utilization_chart.label_circle_050" },
  { value: ChartType.CircleSize75, label: "opts.utilization_chart.label_circle_075" },
  { value: ChartType.CircleSize100, label: "opts.utilization_chart.label_circle_100" },
  { value: ChartType.SparkArea, label: "opts.utilization_chart.label_spark_area" },
  { value: ChartType.SparkLine, label: "opts.utilization_chart.label_spark_line" },
  { value: ChartType.SparkBar, label: "opts.utilization_chart.label_spark_bar" },
  { value: ChartType.Table, label: "opts.utilization_chart.label_table" },
]

// table sub types
export const TableSubType = {
  Grouped: "grouped",
  Mixed: "mixed",
}

// table sub type options
export const TableSubTypeOptions = [
  {
    value: TableSubType.Grouped,
    label: "opts.utilization_table_sub_type.label_grouped",
    description: "opts.utilization_table_sub_type.desc_grouped",
  },
  {
    value: TableSubType.Mixed,
    label: "opts.utilization_table_sub_type.label_mixed",
    description: "opts.utilization_table_sub_type.desc_mixed",
  },
]
