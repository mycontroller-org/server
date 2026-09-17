import { Divider } from "@patternfly/react-core"
import moment from "moment"
import React from "react"
import Measure from "react-measure"
import { LineChart } from "../../Components/Graphs/Graphs"
import {
  AggregationInterval,
  Duration,
  DurationOptions,
  InterpolationType,
  MetricFunctionType,
  MetricType,
} from "../../Constants/Metric"
import { api } from "../../Service/Api"
import { getItem, getValue } from "../../Util/Util"

const graphable = (metricType) =>
  metricType === MetricType.Binary ||
  metricType === MetricType.Gauge ||
  metricType === MetricType.GaugeFloat

const FieldSidebarGraph = ({ field }) => {
  const [points, setPoints] = React.useState(null)
  const [unit, setUnit] = React.useState("")
  const [interpolation, setInterpolation] = React.useState(InterpolationType.Natural)

  React.useEffect(() => {
    if (!field || !field.id || !graphable(field.metricType)) {
      setPoints(null)
      return undefined
    }
    let cancelled = false
    const isBinary = field.metricType === MetricType.Binary
    const start = isBinary ? Duration.Last2Hours : Duration.LastHour
    const duration = getItem(start, DurationOptions)
    const func = MetricFunctionType.Mean
    api.metric
      .fetch({
        global: {
          metricType: field.metricType,
          start,
          window: AggregationInterval.Minute_1,
          functions: [func],
        },
        individual: [{ name: "field_graph", tags: { id: field.id } }],
      })
      .then((res) => {
        if (cancelled) {
          return
        }
        const raw = res.data && res.data.field_graph
        if (!raw || !raw.length) {
          setPoints([])
          return
        }
        const tsFormat = isBinary ? `${duration.tsFormat}:ss` : duration.tsFormat
        const data = []
        raw.forEach((d) => {
          const rawY = isBinary ? getValue(d, "metric.value", undefined) : getValue(d, `metric.${func}`, undefined)
          if (rawY === undefined || rawY === null || rawY === "") {
            return
          }
          const y = Number(rawY)
          if (Number.isNaN(y)) {
            return
          }
          data.push({
            x: moment(d.timestamp).format(tsFormat),
            y,
          })
        })
        if (isBinary && data.length > 0) {
          data.push({
            x: moment().format(tsFormat),
            y: data[data.length - 1].y,
          })
        }
        setUnit(field.unit && field.unit !== "none" ? field.unit : "")
        setInterpolation(isBinary ? InterpolationType.StepAfter : InterpolationType.Natural)
        setPoints(data)
      })
      .catch(() => {
        if (!cancelled) {
          setPoints([])
        }
      })
    return () => {
      cancelled = true
    }
  }, [field && field.id, field && field.metricType, field && field.unit])

  if (!field || !graphable(field.metricType) || !points || !points.length) {
    return null
  }

  const ys = points.map((p) => p.y).filter((y) => y !== undefined && y !== null)
  const dataMin = ys.length ? Math.min(...ys) : 0
  const dataMax = ys.length ? Math.max(...ys) : 0
  const span = dataMax - dataMin
  const pad = span !== 0 ? span * 0.05 : Math.abs(dataMin) * 0.05
  const minDomainY = dataMin < 0 ? dataMin - pad : 0

  return (
    <div className="topology-sidebar-graph">
      <Divider className="topology-sidebar-rule" />
      <div className="topology-sidebar-graph-body">
        <Measure offset>
          {({ measureRef, contentRect }) => (
            <div ref={measureRef}>
              <LineChart
                title=""
                unit={unit}
                data={points}
                interpolation={interpolation}
                type={field.metricType === MetricType.Binary ? "line" : "area"}
                height={140}
                width={contentRect.offset.width ? contentRect.offset.width : 280}
                tickCountX={3}
                tickCountY={3}
                minDomainY={minDomainY}
              />
            </div>
          )}
        </Measure>
      </div>
    </div>
  )
}

export default FieldSidebarGraph
export { graphable }
