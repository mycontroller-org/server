import { ChartThemeColor } from "@patternfly/react-charts"

// chart group types
export const ChartGroupType = {
  GroupChart: "group_chart",
  MixedChart: "mixed_chart",
  PieChart: "pie_chart",
}

// chart group type options
export const ChartGroupTypeOptions = [
  {
    value: ChartGroupType.GroupChart,
    label: "opts.chart_group_type.label_group_chart",
    description: "opts.chart_group_type.desc_group_chart",
  },
  {
    value: ChartGroupType.MixedChart,
    label: "opts.chart_group_type.label_mixed_chart",
    description: "opts.chart_group_type.desc_mixed_chart",
  },
  {
    value: ChartGroupType.PieChart,
    label: "opts.chart_group_type.label_pie_chart",
    description: "opts.chart_group_type.desc_pie_chart",
  },
]

// Charts types
export const ChartType = {
  AreaChart: "area_chart",
  BarChart: "bar_chart",
  LineChart: "line_chart",
  PieChart: "pie_chart",
  ScatterChart: "scatter_chart",
}

// charts type options
export const ChartTypeOptions = [
  {
    value: ChartType.AreaChart,
    label: "opts.chart_type.label_area_chart",
    description: "opts.chart_type.desc_area_chart",
  },
  {
    value: ChartType.BarChart,
    label: "opts.chart_type.label_bar_chart",
    description: "opts.chart_type.desc_bar_chart",
  },
  {
    value: ChartType.LineChart,
    label: "opts.chart_type.label_line_chart",
    description: "opts.chart_type.desc_line_chart",
  },
  {
    value: ChartType.PieChart,
    label: "opts.chart_type.label_pie_chart",
    description: "opts.chart_type.desc_pie_chart",
  },
  {
    value: ChartType.ScatterChart,
    label: "opts.chart_type.label_scatter_chart",
    description: "opts.chart_type.desc_scatter_chart",
  },
]

// legend position types
export const LegendPositionType = {
  BottomLeft: "bottom-left",
  Bottom: "bottom",
  Top: "top",
  Right: "right",
}

// legend position types options
export const LegendPositionTypeOptions = [
  {
    value: LegendPositionType.BottomLeft,
    label: "opts.legend_position.label_bottom_left",
    description: "opts.legend_position.desc_bottom_left",
  },
  {
    value: LegendPositionType.Bottom,
    label: "opts.legend_position.label_bottom_center",
    description: "opts.legend_position.desc_bottom_center",
  },
  {
    value: LegendPositionType.Top,
    label: "opts.legend_position.label_top_left",
    description: "opts.legend_position.desc_top_left",
  },
  {
    value: LegendPositionType.Right,
    label: "opts.legend_position.label_right",
    description: "opts.legend_position.desc_right",
  },
]

// legend orientation types
export const LegendOrientationType = {
  Horizontal: "horizontal",
  Vertical: "vertical",
}

// legend orientation types options
export const LegendOrientationTypeOptions = [
  {
    value: LegendOrientationType.Horizontal,
    label: "opts.legend_orientation.label_horizontal",
    description: "opts.legend_orientation.desc_horizontal",
  },
  {
    value: LegendOrientationType.Vertical,
    label: "opts.legend_orientation.label_vertical",
    description: "opts.legend_orientation.desc_vertical",
  },
]

// ThemeColor used on the chart
export const ThemeColor = {
  Blue: ChartThemeColor.blue,
  Cyan: ChartThemeColor.cyan,
  Gold: ChartThemeColor.gold,
  Gray: ChartThemeColor.gray,
  Green: ChartThemeColor.green,
  Orange: ChartThemeColor.orange,
  Purple: ChartThemeColor.purple,
  MultiOrdered: ChartThemeColor.multiOrdered,
  MultiUnordered: ChartThemeColor.multiUnordered,
}

// ThemeColorOptions used on the chart
export const ThemeColorOptions = [
  { value: ThemeColor.Blue, label: "opts.theme_color.label_blue" },
  { value: ThemeColor.Cyan, label: "opts.theme_color.label_cyan" },
  { value: ThemeColor.Gold, label: "opts.theme_color.label_gold" },
  { value: ThemeColor.Green, label: "opts.theme_color.label_green" },
  { value: ThemeColor.Orange, label: "opts.theme_color.label_orange" },
  { value: ThemeColor.Purple, label: "opts.theme_color.label_purple" },
  { value: ThemeColor.MultiOrdered, label: "opts.theme_color.label_multiple_ordered" },
  { value: ThemeColor.MultiUnordered, label: "opts.theme_color.label_multiple_unordered" },
]
