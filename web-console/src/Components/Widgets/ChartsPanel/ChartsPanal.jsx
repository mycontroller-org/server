import React from "react"
import {
  Duration,
  DurationOptions,
  getRecommendedInterval,
  MetricFunctionType,
  MetricType,
  RefreshIntervalType,
} from "../../../Constants/Metric"
import { api } from "../../../Service/Api"
import { getItem, getNodeMetricFieldName, getNodeMetricType, getValue, isEqual } from "../../../Util/Util"
import MultipleAxes from "../../Graphs/MultipleAxes/MultipleAxes"
import Loading from "../../Loading/Loading"
import moment from "moment"
import { ChartGroupType } from "../../../Constants/Widgets/ChartsPanel"
import { getListAPI } from "../ControlPanel/Common/Utils"
import { ResourceType } from "../../../Constants/Resource"
import { getQuickId } from "../../../Constants/ResourcePicker"
import { FIELD_NAME } from "../../../Constants/Common"

class ChartsPanel extends React.Component {
  state = {
    loading: true,
    error: "",
    resources: [],
    resourcesRejected: [],
    metricsData: {},
  }

  loadMetricsForGroupChart = () => {
    const { resource: resourceCfg, chart: chartCfg = {} } = this.props.config
    const {
      filters: resourceFilters,
      limit: itemsLimit = 5,
      type: resourceType = ResourceType.Field,
      nameKey = "name",
    } = resourceCfg
    const filterKeys = Object.keys(resourceFilters)
    const filters = filterKeys.map((key) => {
      return { k: key, v: resourceFilters[key] }
    })

    const listAPI = getListAPI(resourceType)

    listAPI({ filter: filters, limit: itemsLimit })
      .then((res) => {
        const resourcesAccepted = []
        const metricQueryIndividuals = []
        res.data.data.forEach((item) => {
          let metricType = item.metricType
          const quickId = getQuickId(resourceType, item)
          const extraTags = {}
          if (quickId.startsWith(ResourceType.Node)) {
            metricType = getNodeMetricType()
            extraTags[FIELD_NAME] = getNodeMetricFieldName()
          }

          switch (metricType) {
            case MetricType.Gauge:
            case MetricType.GaugeFloat:
              const resourceName = getValue(item, nameKey, `undefined_${quickId}`)
              resourcesAccepted.push({
                chartType: chartCfg.chartType,
                color: "",
                disabled: false,
                nameKey: nameKey,
                quickId: quickId,
                resourceType: resourceType,
                unit: resourceCfg.unit,
                useGlobalStyle: true,
                style: {},
                yAxis: "0",
                name: resourceName,
              })
              metricQueryIndividuals.push({
                name: resourceName,
                metricType: metricType,
                tags: { id: item.id, ...extraTags },
              })
              break

            default:
            //noop
          }
        })

        // get metrics
        if (metricQueryIndividuals.length > 0) {
          const metricDuration = getValue(chartCfg, "duration", Duration.LastHour)
          const metricInterval = getValue(chartCfg, "interval", getRecommendedInterval(metricDuration))
          const metricFunction = getValue(chartCfg, "chart.metricFunction", MetricFunctionType.Mean)

          const duration = getItem(metricDuration, DurationOptions)

          const metricQuery = {
            global: {
              start: duration.value,
              window: metricInterval,
              functions: [metricFunction],
            },
            individual: metricQueryIndividuals,
          }

          api.metric
            .fetch(metricQuery)
            .then((metricRes) => {
              // update metrics data
              const names = Object.keys(metricRes.data)
              const metricsData = {}
              names.forEach((name) => {
                const metricsRaw = metricRes.data[name]
                // update data into metrics object
                metricsData[name] = getMetrics(metricsRaw, duration.tsFormat, metricFunction)
              })
              this.setState({
                metricsData: metricsData,
                resources: resourcesAccepted,
                resourcesRejected: [],
                loading: false,
                error: "",
              })
            })
            .catch((_e) => {
              this.setState({ loading: false, error: "error on get metric data" })
            })
        } else {
          this.setState({ loading: false, error: "no resource available to get metric data" })
        }
      })
      .catch((_e) => {
        this.setState({ isLoading: false })
      })
  }

  loadMetricsMixedChart = () => {
    const { resources: resourcesCfg, chart: chartCfg = {} } = this.props.config

    // following code for Mixed charts
    const quickIds = []
    resourcesCfg.forEach((r) => {
      if (!r.disabled) {
        quickIds.push(`${r.resourceType}:${r.quickId}`)
      }
    })

    api.quickId
      .getResources({ id: quickIds })
      .then((res) => {
        const resourceMap = res.data
        const resourcesAccepted = []
        const resourcesRejected = []
        const metricQueryIndividuals = []
        // get resources - done
        // get resource name
        // create metric query array
        // get metrics and update
        Object.keys(resourceMap).forEach((quickId) => {
          const resource = resourceMap[quickId]
          if (quickId.startsWith(ResourceType.Field) || quickId.startsWith(ResourceType.Node)) {
            let metricType = getValue(resource, "metricType", MetricType.None)

            // extra tags used to supply node or other resource fields
            const extraTags = {}

            // for node type resources, for now supports only for battery_level
            if (quickId.startsWith(ResourceType.Node)) {
              metricType = getNodeMetricType()
              extraTags[FIELD_NAME] = getNodeMetricFieldName()
            }
            switch (metricType) {
              case MetricType.Gauge:
              case MetricType.GaugeFloat:
                // case MetricType.Binary:
                let found = false
                for (let rIndex = 0; rIndex < resourcesCfg.length; rIndex++) {
                  const resCfg = resourcesCfg[rIndex]
                  const rQuickId = `${resCfg.resourceType}:${resCfg.quickId}`
                  const resourceName = getValue(resource, resCfg.nameKey, `undefined_${quickId}`)
                  // console.log("q1:", quickId, "q2:", rQuickId, "name:", resourceName)
                  if (quickId === rQuickId) {
                    resourcesAccepted.push({
                      ...resCfg,
                      name: resourceName,
                    })
                    // this config used to get metrics data
                    metricQueryIndividuals.push({
                      name: resourceName,
                      metricType: metricType,
                      tags: { id: resource.id, ...extraTags },
                    })
                    found = true
                    break
                  }
                }
                if (!found) {
                  resourcesRejected.push({
                    ...resource,
                  })
                }
                break

              default:
                // add to rejected list
                resourcesRejected.push({
                  ...resource,
                })
            }
          }
        })
        if (metricQueryIndividuals.length > 0) {
          const metricDuration = getValue(chartCfg, "duration", Duration.LastHour)
          const metricInterval = getValue(chartCfg, "interval", getRecommendedInterval(metricDuration))
          const metricFunction = getValue(chartCfg, "chart.metricFunction", MetricFunctionType.Mean)

          const duration = getItem(metricDuration, DurationOptions)

          const metricQuery = {
            global: {
              start: duration.value,
              window: metricInterval,
              functions: [metricFunction],
            },
            individual: metricQueryIndividuals,
          }

          api.metric
            .fetch(metricQuery)
            .then((metricRes) => {
              // update metrics data
              const names = Object.keys(metricRes.data)
              const metricsData = {}
              names.forEach((name) => {
                const metricsRaw = metricRes.data[name]
                // update data into metrics object
                metricsData[name] = getMetrics(metricsRaw, duration.tsFormat, metricFunction)
              })
              this.setState({
                metricsData: metricsData,
                resources: resourcesAccepted,
                resourcesRejected: resourcesRejected,
                loading: false,
                error: "",
              })
            })
            .catch((_e) => {
              this.setState({ loading: false, error: "error on get metric data" })
            })
        } else {
          this.setState({ loading: false, error: "no resource available to get metric data" })
        }
      })
      .catch((_e) => {
        this.setState({ loading: false, error: "error on getting resource details" })
      })
  }

  loadMetrics = () => {
    this.lastMetricUpdate = new Date().getTime() // reference point for auto refresh job
    const { subType } = this.props.config

    if (subType === ChartGroupType.GroupChart) {
      this.loadMetricsForGroupChart()
    } else {
      this.loadMetricsMixedChart()
    }
  }

  updateRefreshInterval = () => {
    // update metrics query on interval
    if (this.interval) {
      clearInterval(this.interval)
    }
    const { config } = this.props
    const refreshInterval = getValue(config, "chart.refreshInterval", RefreshIntervalType.None)
    if (refreshInterval >= 1000) {
      this.interval = setInterval(() => {
        const diffTime = new Date().getTime() - this.lastMetricUpdate + 200 // add 200 milliseconds as offset
        if (diffTime >= refreshInterval) {
          this.loadMetrics()
        }
      }, refreshInterval)
    }
  }

  componentWillUnmount() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  }

  componentDidMount() {
    this.loadMetrics()
    this.updateRefreshInterval()
  }

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.config, this.props.config)) {
      this.loadMetrics()
      this.updateRefreshInterval()
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return !isEqual(this.state, nextState) || !isEqual(this.props, nextProps)
  }

  render() {
    const { loading, metricsData, resources: resourcesCfg = [], error } = this.state

    // console.log(JSON.stringify(metricsData))

    if (error !== "") {
      return <span>Error: {error}</span>
    }

    if (loading) {
      return <Loading />
    }

    const { config, widgetId } = this.props

    const { axisY: axisCfg, chart: basicConfig } = config

    const { dimensions } = this.props
    const width = dimensions.width
    const chartConfig = {
      basicConfig: basicConfig,
      axisConfig: axisCfg,
    }

    const newMetrics = []
    Object.keys(metricsData).forEach((metricName) => {
      const metricConfig = getItem(metricName, resourcesCfg, -1, "name")
      if (metricConfig !== null) {
        newMetrics.push({
          ...metricConfig,
          data: this.state.metricsData[metricName].data,
        })
      }
    })
    //console.log(newMetrics)
    return (
      <MultipleAxes key={`cp_${widgetId}`} width={width} chartConfig={chartConfig} metrics={newMetrics} />
    )
  }
}

export default ChartsPanel

const getMetrics = (metricsRaw = [], tsFormat, metricFunction) => {
  if (metricsRaw.length === 0) {
    return { data: [] }
  }

  const metricType = metricsRaw[0].metricType
  switch (metricType) {
    case MetricType.Binary:
      return getBinaryMetrics(metricsRaw, `${tsFormat}:ss`)

    case MetricType.Gauge:
    case MetricType.GaugeFloat:
      return getGaugeMetrics(metricsRaw, tsFormat, metricFunction)

    default:
      return { data: [] }
  }
}

const getBinaryMetrics = (metricsRaw = [], tsFormat) => {
  const data = []
  metricsRaw.forEach((d) => {
    const ts = moment(d.timestamp).format(tsFormat)
    data.push({
      x: ts,
      y: d.metric.value,
    })
  })
  return { data: data }
}

const getGaugeMetrics = (metricsRaw = [], tsFormat, metricFunction) => {
  const data = []
  let minValue = Infinity
  let maxValue = -Number.MAX_VALUE * 2
  metricsRaw.forEach((d) => {
    const ts = moment(d.timestamp).format(tsFormat)
    // update data
    const yValue = d.metric[metricFunction]
    data.push({
      x: ts,
      y: yValue,
    })
    if (yValue) {
      if (minValue > yValue) {
        minValue = yValue
      }
      if (maxValue < yValue) {
        maxValue = yValue
      }
    }
  })
  return { data: data, minValue: minValue, maxValue: maxValue }
}
