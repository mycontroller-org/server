import {
  Chart,
  ChartArea,
  ChartAxis,
  ChartBar,
  ChartGroup,
  ChartLegendTooltip,
  ChartLine,
  ChartPie,
  ChartScatter,
  ChartStack,
  ChartThemeColor,
  ChartTooltip,
  ChartVoronoiContainer,
  createContainer,
} from "@patternfly/react-charts"
import React from "react"
import v from "validator"
import { ChartType } from "../../../Constants/Widgets/ChartsPanel"
import { getValue } from "../../../Util/Util"

const CursorVoronoiContainer = createContainer("voronoi", "cursor")

const defaultValues = {
  padding: {
    bottom: 55,
    left: 70,
    right: 30,
    top: 20,
  },
  colors: [],
}

// sample data format
const _chartConfig = {
  basicConfig: {
    legendPosition: "bottom-left",
    legendOrientation: "horizontal",
    interpolation: "basis",
    height: 300,
    showGridX: true,
    showGridY: true,
    padding: {
      bottom: 60,
      left: 100,
      right: 50,
      top: 30,
    },
  },
  axisConfig: {
    1: {
      offsetY: 0, // in percentage, 0 to 100. left to right
      unit: "°C",
      padding: 0,
    },
    2: {
      offsetY: 100,
      unit: "%",
      padding: 30,
    },
  },
}

const _metrics = [
  {
    name: "Cats",
    yAxis: 0,
    interpolation: "basis",
    color: "",
    strokeWidth: 1, // supports in decimal values, example: 0.1
    roundDecimal: 2,
    data: [
      { x: "2015", y: 30 },
      { x: "2016", y: 40 },
      { x: "2017", y: 80 },
      { x: "2018", y: 60 },
    ],
  },
]

const barWidth = 11

const MultipleAxes = ({ width = 300, chartConfig = {}, metrics = [], backgroundColor = "transparent" }) => {
  const { basicConfig, axisConfig } = chartConfig

  const paddingConfig = { ...defaultValues.padding, ...basicConfig.padding }
  const chartActualWidth = width - paddingConfig.left - paddingConfig.right

  const charts = []
  const chartAxis = [<ChartAxis key="default_axis_x" showGrid={basicConfig.showGridX} fixLabelOverlap />]

  const yAxisMaxima = []
  const yAxisMinima = []
  Object.keys(chartConfig.axisConfig).forEach(() => {
    yAxisMaxima.push(0)
    yAxisMinima.push(Number.MAX_SAFE_INTEGER)
  })

  // get yAxis maxima index
  const getMaximaIndex = (dataset = {}) => {
    let index = getValue(dataset, "yAxis", 0)
    if (index >= yAxisMaxima.length) {
      index = 0
    }
    return index
  }

  // find maxima for normalizing data
  metrics.forEach((dataset) => {
    const maximaIndex = getMaximaIndex(dataset)
    const currentMax = Math.max(...dataset.data.map((d) => d.y))

    if (yAxisMaxima[maximaIndex] < currentMax) {
      yAxisMaxima[maximaIndex] = currentMax
    }

    const currentMin = Math.min(...dataset.data.map((d) => d.y).filter((val) => val !== null))
    if (yAxisMinima[maximaIndex] > currentMin) {
      yAxisMinima[maximaIndex] = currentMin
    }
  })

  // update chart axis
  const axisKeys = Object.keys(axisConfig).sort()
  axisKeys.forEach((axisCfgKey, index) => {
    const axisCfg = axisConfig[axisCfgKey]
    const tickLabels = {}
    const axis = {}
    if (axisCfg.offsetY !== 0 && axisCfg.offsetY !== 100) {
      // fill with custom color, if supplied
      const customColor = axisCfg.color !== undefined && axisCfg.color !== "" ? axisCfg.color : "#555"
      //tickLabels.fontWeight = 800
      tickLabels.fill = customColor
      axis.stroke = customColor
      axis.strokeWidth = 1
    } else {
      // update for left and right axis
      if (axisCfg.color !== undefined && axisCfg.color !== "") {
        tickLabels.fill = axisCfg.color
        axis.stroke = axisCfg.color
      }
    }

    // get yAxis index
    let maximaIndex = parseInt(axisCfgKey)
    if (maximaIndex > yAxisMaxima.length) {
      maximaIndex = 0
    }
    chartAxis.push(
      <ChartAxis
        key={`chart_axis_index_${index}`}
        dependentAxis
        showGrid={basicConfig.showGridY}
        style={{
          axis: { ...axis },
          ticks: { padding: -axisCfg.padding },
          // tickLabels: { stroke: "red" },
          tickLabels: { ...tickLabels },
        }}
        offsetX={chartActualWidth * (axisCfg.offsetY / 100) + paddingConfig.left}
        // Use normalized tickValues (0 - 1)
        // tickValues={[0.25, 0.5, 0.75, 1]}
        tickCount={10}
        fixLabelOverlap
        // Re-scale ticks by multiplying by correct maxima
        tickFormat={(t) => {
          const value = (t * yAxisMaxima[maximaIndex]).toFixed(axisCfg.roundDecimal)
          if (axisCfg.unit !== "") {
            return `${value} ${axisCfg.unit}`
          }
          return value
        }}
      />
    )
  })

  const legendData = []

  // update chart and chart axis
  let totalBarCharts = 0
  metrics.forEach((metricData, index) => {
    // get style details
    const styleConfig = metricData.useGlobalStyle
      ? {
          interpolation: basicConfig.interpolation,
          fillOpacity: basicConfig.fillOpacity,
          strokeWidth: basicConfig.strokeWidth,
          roundDecimal: basicConfig.roundDecimal,
        }
      : { ...metricData.style }

    // fillOpacity will be in %, convert to 0~1
    styleConfig.fillOpacity = styleConfig.fillOpacity / 100

    const name = metricData.name
    const legend = { name: name, childName: name }
    if (metricData.color !== "") {
      legend.symbol = { fill: metricData.color }
    }
    legendData.push(legend)
    // add name into the data
    metricData.data.forEach((_, index) => {
      metricData.data[index].name = name
      metricData.data[index].unit = metricData.unit
      metricData.data[index].roundDecimal = styleConfig.roundDecimal
    })

    const styleData = {
      fillOpacity: styleConfig.fillOpacity,
      strokeWidth: styleConfig.strokeWidth,
    }

    if (metricData.color !== undefined && metricData.color !== "") {
      styleData.fill = metricData.color
      styleData.stroke = metricData.color
    }

    let ChartSelected = null
    switch (metricData.chartType) {
      case ChartType.LineChart:
        ChartSelected = ChartLine
        delete styleData.fill
        break

      case ChartType.AreaChart:
        ChartSelected = ChartArea
        break

      case ChartType.BarChart:
        ChartSelected = ChartBar
        totalBarCharts++
        break

      case ChartType.PieChart:
        ChartSelected = ChartPie
        break

      case ChartType.ScatterChart:
        ChartSelected = ChartScatter
        break

      default:
        ChartSelected = ChartArea
    }

    const maximaIndex = getMaximaIndex(metricData)
    let axisMinimaValue = null
    const domain = {}
    if (basicConfig.stackCharts === false) {
      // when the minimum and maximum values are zero for a axis, pull all the axis to it's zero level
      // to fix that issue, here is ugly hack
      // take lowest from other axis and use it as lowest also use that value as zero ==> 'axisMinimaValue'
      if (yAxisMinima[maximaIndex] === 0 && yAxisMaxima[maximaIndex] === 0) {
        const minimaTmpValues = []
        yAxisMinima.forEach((_, tmpIndex) => {
          if (!(yAxisMinima[tmpIndex] === 0 && yAxisMaxima[tmpIndex] === 0)) {
            minimaTmpValues.push(yAxisMinima[tmpIndex] / yAxisMaxima[tmpIndex])
          }
        })
        axisMinimaValue = Math.min(...minimaTmpValues)
        domain.y = [axisMinimaValue, 1]
      } else {
        domain.y = [yAxisMinima[maximaIndex] / yAxisMaxima[maximaIndex], 1]
      }
    }
    const chart = (
      <ChartSelected
        key={`chart_index_${index}`}
        style={{
          data: styleData,
        }}
        data={metricData.data}
        interpolation={styleConfig.interpolation}
        y={(datum) => {
          if (datum.y === null) {
            return null
          }
          // if 'axisMinimaValue' indicates that this particular axis zero as all values
          // so for those axis always return 'axisMinimaValue' as zero
          if (axisMinimaValue !== null) {
            return axisMinimaValue
          }
          if (datum.y === 0) {
            return datum.y
          }
          return datum.y / yAxisMaxima[maximaIndex]
        }}
        domain={domain}
        barWidth={barWidth}
        name={name}
      />
    )

    charts.push(chart)
  })

  const getDisplayLabel = (datum = {}, withTime = true, nullValue = null) => {
    let value = datum.y
    if (v.isFloat(String(datum.y), {})) {
      value = datum.y.toFixed(datum.roundDecimal)
    }
    if (datum.y || datum.y === 0) {
      // round decimal number
      const yValue = datum.y % 1 === 0 ? datum.y : datum.y.toFixed(datum.roundDecimal)
      if (withTime) {
        return `${yValue} ${datum.unit} at ${datum.x}`
      }
      return `${yValue} ${datum.unit}`
    }
    return nullValue
  }

  const toolTipComponent =
    basicConfig.cursorTooltip === true ? (
      <CursorVoronoiContainer
        labels={({ datum }) => getDisplayLabel(datum, false, "-")}
        cursorDimension="x" // enable this to show only vertical line
        voronoiDimension="x"
        mouseFollowTooltips
        voronoiPadding={paddingConfig}
        labelComponent={
          <ChartLegendTooltip cornerRadius={3} legendData={legendData} title={(datum) => datum.x} />
        }
        constrainToVisibleArea
        isCursorTooltip={false}
      />
    ) : (
      <ChartVoronoiContainer
        labelComponent={<ChartTooltip cornerRadius={3} />}
        labels={({ datum }) => getDisplayLabel(datum)}
        constrainToVisibleArea
      />
    )

  return (
    <div style={{ backgroundColor: backgroundColor }}>
      <Chart
        standalone={true}
        containerComponent={toolTipComponent}
        legendData={legendData}
        legendOrientation={basicConfig.legendOrientation}
        legendPosition={basicConfig.legendPosition}
        height={basicConfig.height}
        width={width}
        scale={{ x: "time", y: "linear" }}
        padding={paddingConfig}
        // if we change 'y' bottom padding from 0, 'axisMinimaValue' concept will not be perfect!
        domainPadding={{ y: [0, 20] }}
        themeColor={basicConfig.themeColor !== undefined ? basicConfig.themeColor : ChartThemeColor.default}
      >
        {chartAxis}
        <ChartGroup offset={totalBarCharts > 1 ? barWidth : 0}>
          {basicConfig.stackCharts === true ? <ChartStack>{charts}</ChartStack> : charts}
        </ChartGroup>
      </Chart>
    </div>
  )
}

export default MultipleAxes
