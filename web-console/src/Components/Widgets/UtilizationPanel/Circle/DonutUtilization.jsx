import { ChartDonutUtilization, ChartLabel } from "@patternfly/react-charts"
import React from "react"
import v from "validator"
import { ChartType } from "../../../../Constants/Widgets/UtilizationPanel"
import { getPercentage, getValue } from "../../../../Util/Util"
import { getLastSeen } from "../../../Time/Time"
import "./DonutUtilization.scss"

const DonutUtilization = ({ config = {}, resource = {} }) => {
  const chartType = getValue(config, "type", ChartType.CircleSize50)
  const thresholds = getValue(config, "chart.thresholds", {})
  const thickness = getValue(config, "chart.thickness", 10)
  const cornerSmoothing = getValue(config, "chart.cornerSmoothing", 0)
  const minimumValue = getValue(config, "chart.minimumValue", 0)
  const maximumValue = getValue(config, "chart.maximumValue", 100)
  const unit = getValue(config, "resource.unit", "")
  const displayName = getValue(config, "resource.displayName", false)
  const roundDecimal = getValue(config, "resource.roundDecimal", 2)

  // update inner radius
  const innerRadius = 100 - thickness

  const resourceName = resource.name !== "" ? resource.name : "1"
  const value = getPercentage(resource.value, minimumValue, maximumValue)

  const displayValueFloat = parseFloat(resource.value)
  let displayValue = resource.value

  if (v.isFloat(String(displayValue), {})) {
    displayValue = displayValueFloat.toFixed(roundDecimal)
  }

  const keys = Object.keys(thresholds)

  const tunedThresholds = keys.map((key) => {
    return { value: getPercentage(Number(key), minimumValue, maximumValue), color: thresholds[key] }
  })

  let titleComponent = <ChartLabel />
  let subTitleComponent = <ChartLabel />
  let startAngle = 0
  let endAngle = 360
  let isStandalone = true

  // size calculation
  const definedChartSize = 250

  switch (chartType) {
    case ChartType.CircleSize50:
      titleComponent = <ChartLabel y={0.31 * definedChartSize} />
      subTitleComponent = <ChartLabel y={0.42 * definedChartSize} />
      startAngle = -90
      endAngle = 90
      isStandalone = false
      break

    case ChartType.CircleSize75:
      titleComponent = <ChartLabel y={0.43 * definedChartSize} />
      subTitleComponent = <ChartLabel y={0.54 * definedChartSize} />
      startAngle = -135
      endAngle = 135
      isStandalone = false
      break

    case ChartType.CircleSize100:
      titleComponent = <ChartLabel y={0.43 * definedChartSize} />
      subTitleComponent = <ChartLabel y={0.54 * definedChartSize} />
      isStandalone = false
      break

    default:
  }

  const lastSeenValue = getLastSeen(resource.timestamp)
  const subTitleData = [lastSeenValue, displayName ? resource.name : "-"]
  const chart = (
    <ChartDonutUtilization
      animate={true}
      // ariaDesc="Storage capacity"
      // ariaTitle="Donut utilization chart example"z
      constrainToVisibleArea={true}
      data={{ x: resourceName, y: value }}
      labels={() => {}}
      // labels={({ datum }) => (datum.x ? `${datum.x}: ${datum.y.toFixed(1)}%` : null)}
      title={`${displayValue}${unit}`}
      titleComponent={titleComponent}
      subTitle={subTitleData}
      subTitleComponent={subTitleComponent}
      width={definedChartSize}
      height={definedChartSize}
      cornerRadius={cornerSmoothing}
      startAngle={startAngle}
      endAngle={endAngle}
      innerRadius={innerRadius}
      radius={100}
      thresholds={tunedThresholds}
      standalone={isStandalone}
    />
  )

  switch (chartType) {
    case ChartType.CircleSize50:
      const circle50Height = (definedChartSize * 0.51).toFixed(0)
      return (
        <svg
          viewBox={`23 23 ${definedChartSize - 46} ${circle50Height - 23}`}
          preserveAspectRatio="xMaxYMid meet"
          height={circle50Height}
          width={definedChartSize}
          role="img"
          style={{ height: "100%", width: "100%" }}
        >
          {chart}
        </svg>
      )

    case ChartType.CircleSize75:
      const circle75Height = (definedChartSize * 0.82).toFixed(0)
      return (
        <svg
          // viewBox={`0 0 ${definedChartSize} ${circle75Height}`}
          viewBox={`23 25 ${definedChartSize - 46} ${circle75Height - 35}`}
          preserveAspectRatio="none"
          height={circle75Height}
          width={definedChartSize}
          role="img"
          style={{ height: "100%", width: "100%" }}
        >
          {chart}
        </svg>
      )

    case ChartType.CircleSize100:
      return (
        <svg
          // viewBox={`0 0 ${definedChartSize} ${definedChartSize}`}
          viewBox={`23 23 ${definedChartSize - 46} ${definedChartSize - 46}`}
          preserveAspectRatio="none"
          height={definedChartSize}
          width={definedChartSize}
          role="img"
          style={{ height: "100%", width: "100%" }}
        >
          {chart}
        </svg>
      )

    default:
      return <span>Invalid chart type: {chartType}</span>
  }
}

export default DonutUtilization
