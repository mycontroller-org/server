import {
  Chart,
  ChartArea,
  ChartAxis,
  ChartGroup,
  ChartThemeColor,
  ChartThemeVariant,
  ChartTooltip,
  ChartVoronoiContainer,
  getCustomTheme,
} from "@patternfly/react-charts"
import React from "react"
import "./Graphs.scss"
import { areaTheme } from "./Themes"

const theme = getCustomTheme(ChartThemeColor.blue, ChartThemeVariant.light, areaTheme)

export const LineChart = ({
  title = "Graph title",
  unit = "",
  interpolation = "natural",
  data = [],
  tickCountX = 2,
  tickCountY = 3,
  height = 200,
  width = 600,
  minDomainY = 0,
}) => {
  return (
    <div>
      <h5 className="graph-title">{title}</h5>
      <Chart
        containerComponent={
          <ChartVoronoiContainer
            mouseFollowTooltips={false}
            constrainToVisibleArea
            labelComponent={<ChartTooltip cornerRadius={3} />}
            labels={({ datum }) => {
              if (datum.y || datum.y === 0) {
                // round decimal number
                const yValue = datum.y % 1 === 0 ? datum.y : datum.y.toFixed(2)
                return `${yValue} ${unit} at ${datum.x}`
              }
              return null
            }}
          />
        }
        animate={false}
        theme={theme}
        height={height}
        width={width}
        domainPadding={{ y: 20 }}
        scale={{ x: "time", y: "linear" }}
        minDomain={{ y: minDomainY }}
        //themeColor={ChartThemeColor.default}
      >
        <ChartAxis showGrid fixLabelOverlap />
        <ChartAxis
          dependentAxis
          showGrid
          fixLabelOverlap
          tickFormat={(tick) => {
            if (tick) {
              if (tick % 1 !== 0) {
                return `${tick.toFixed(2)} ${unit}`
              }
              return `${tick} ${unit}`
            }
            return "-"
          }}
        />
        <ChartGroup>
          <ChartArea interpolation={interpolation} data={data} />
        </ChartGroup>
      </Chart>
    </div>
  )
}
