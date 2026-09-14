import {
  ChartArea,
  ChartBar,
  ChartGroup,
  ChartLine,
  ChartThemeColor,
  ChartTooltip,
  ChartVoronoiContainer,
} from "@patternfly/react-charts"
import { Stack, StackItem } from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import v from "validator"
import { InterpolationType, MetricType } from "../../../../Constants/Metric"
import { ChartType } from "../../../../Constants/Widgets/UtilizationPanel"
import { capitalizeFirstLetter, getValue } from "../../../../Util/Util"
import { LastSeen } from "../../../Time/Time"
import "./SparkLine.scss"

const SparkLine = ({
  widgetId = "",
  config = {},
  resource = {},
  metric = {},
  dimensions = {},
  isMetricsLoading,
}) => {
  const chartType = getValue(config, "type", ChartType.SparkLine)
  const chartColor = getValue(config, "chart.color", ChartThemeColor.blue)
  const strokeWidth = getValue(config, "chart.strokeWidth", 2)
  const fillOpacity = getValue(config, "chart.fillOpacity", 2)
  const chartInterpolation = getValue(config, "chart.interpolation", InterpolationType.Basis)
  const metricFunction = getValue(config, "metric.metricFunction", "")
  const unit = getValue(config, "resource.unit", "")
  const roundDecimal = getValue(config, "resource.roundDecimal", 2)
  const chartHeight = getValue(config, "chart.height", 100)
  const yAxisMinValue = getValue(config, "chart.yAxisMinValue", "")
  const yAxisMaxValue = getValue(config, "chart.yAxisMaxValue", "")
  const chartWidth = dimensions.width + 20 // include padding px to put the char till edge of the container

  const { t } = useTranslation()

  const displayValueFloat = parseFloat(resource.value)
  let displayValue = resource.value

  if (v.isFloat(String(displayValue), {})) {
    displayValue = displayValueFloat.toFixed(roundDecimal)
  }

  const isMetricDataAvailable = getValue(metric, "data.length", 0)

  let metricsChart = null

  if (isMetricDataAvailable) {
    let ChartSelected = null

    switch (chartType) {
      case ChartType.SparkArea:
        ChartSelected = ChartArea
        break

      case ChartType.SparkLine:
        ChartSelected = ChartLine
        break

      case ChartType.SparkBar:
        ChartSelected = ChartBar
        // chart = <ChartBar animate={false} data={metric.data} />
        break

      default:
    }

    const chart = (
      <ChartSelected
        key={`${chartType}_${widgetId}`}
        style={{
          data: {
            // fill: chartColor, // this color is handled on chartGroup properties
            fillOpacity: fillOpacity / 100,
            stroke: chartColor,
            strokeWidth: strokeWidth,
          },
        }}
        interpolation={chartInterpolation}
        animate={false}
        data={metric.data}
      />
    )

    const minValue =
      metric.minValue !== undefined && metric.minValue !== Number.POSITIVE_INFINITY
        ? metric.minValue * 0.995
        : 0
    metricsChart = (
      <div
        key={`chart_root_${widgetId}`}
        className="metrics-chart"
        style={{ height: `${chartHeight}px`, width: `${chartWidth}px` }}
      >
        <ChartGroup
          // ariaTitle="hello"
          standalone={true}
          height={chartHeight}
          width={chartWidth}
          domainPadding={{ y: 9 }}
          minDomain={{ y: yAxisMinValue !== "" ? yAxisMinValue : minValue }}
          maxDomain={yAxisMaxValue !== "" ? { y: yAxisMaxValue } : {}}
          containerComponent={
            <ChartVoronoiContainer
              labels={({ datum }) => {
                if (datum.y === null) {
                  return null
                }
                if (datum.y || datum.y === 0) {
                  // round decimal number
                  const yValue = datum.y % 1 === 0 ? datum.y : datum.y.toFixed(roundDecimal)
                  return `${yValue} ${unit} at ${datum.x}`
                }
                return null
              }}
              labelComponent={<ChartTooltip cornerRadius={0} />}
              constrainToVisibleArea
            />
          }
          padding={0}
          color={chartColor}
          scale={{ x: "time", y: "linear" }}
        >
          {chart}
        </ChartGroup>
      </div>
    )
  } else if (isMetricsLoading) {
    metricsChart = (
      <span className="no-metric-data" style={{ color: chartColor }}>
        Loading Metric data
      </span>
    )
  } else {
    metricsChart = (
      <span className="no-metric-data" style={{ color: chartColor }}>
        Metric data not available
      </span>
    )
  }

  let unitText = unit
  let metricFunctionText = ""
  let displayValueText = displayValue

  if (resource.metricType === MetricType.Binary) {
    unitText = ""
    displayValueText = displayValue === "true" ? t("on") : t("off")
  }

  if (resource.metricType !== MetricType.Binary && isMetricDataAvailable) {
    metricFunctionText = `${capitalizeFirstLetter(metricFunction)}`
  }

  return (
    <>
      <Stack className="mc-spark-chart" hasGutter={false}>
        <StackItem>
          <span className="title">{resource.name}</span>
        </StackItem>
        <StackItem>
          <span>
            <span className="value">{displayValueText}</span>
            <span className="value-unit">{unitText}</span>
          </span>
        </StackItem>
        <StackItem>
          <span className="value-timestamp">
            <LastSeen date={resource.timestamp} tooltipPosition="top" />
          </span>
        </StackItem>
        <StackItem>
          <span className="metric-function">{metricFunctionText}</span>
        </StackItem>
        <StackItem>{metricsChart}</StackItem>
      </Stack>
    </>
  )
}

export default SparkLine
