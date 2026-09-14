import moment from "moment"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { FIELD_NAME } from "../../../Constants/Common"
import {
  Duration,
  DurationOptions,
  getRecommendedInterval,
  MetricFunctionType,
  MetricType,
  RefreshIntervalType,
} from "../../../Constants/Metric"
import { ResourceFilterType } from "../../../Constants/Resource"
import { getQuickId, ResourceType } from "../../../Constants/ResourcePicker"
import { ChartType } from "../../../Constants/Widgets/UtilizationPanel"
import { api } from "../../../Service/Api"
import { loadData, unloadData } from "../../../store/entities/websocket"
import { getItem, getNodeMetricFieldName, getNodeMetricType, getValue, isEqual } from "../../../Util/Util"
import Loading from "../../Loading/Loading"
import { isShouldComponentUpdateWithWsData } from "../Helper/Common"
import { navigateToResource } from "../Helper/Resource"
import DonutUtilization from "./Circle/DonutUtilization"
import SparkLine from "./Spark/SparkLine"
import TableUtilization from "./Table/TableUtilization"
import "./UtilizationPanel.scss"

const wsKey = "dashboard_utilization_panel"

class UtilizationPanel extends React.Component {
  state = {
    loading: true,
    loadingMetrics: true,
    resources: [],
    metrics: {},
  }

  getWsKey = () => {
    return `${wsKey}_${this.props.widgetId}`
  }

  componentWillUnmount() {
    if (this.interval) {
      clearInterval(this.interval)
    }
    this.props.unloadData({ key: this.getWsKey() })
  }

  updateRefreshInterval = () => {
    // update metrics query on interval
    if (this.interval) {
      clearInterval(this.interval)
    }
    const { config } = this.props
    const refreshInterval = getValue(config, "metric.refreshInterval", RefreshIntervalType.None)
    if (refreshInterval >= 1000) {
      this.interval = setInterval(() => {
        const diffTime = new Date().getTime() - this.lastMetricUpdate + 200 // add 200 milliseconds as offset
        if (diffTime >= refreshInterval) {
          this.updateMetrics()
        }
      }, refreshInterval)
    }
  }

  componentDidMount() {
    this.updateResourceData()
  }

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.config, this.props.config)) {
      this.updateResourceData()
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return isShouldComponentUpdateWithWsData(this.getWsKey(), this.props, this.state, nextProps, nextState)
  }

  updateMetrics = () => {
    this.lastMetricUpdate = new Date().getTime() // reference point for auto refresh job

    const { config } = this.props

    // if config empty, stop here
    if (config === undefined) {
      return
    }

    const chartType = getValue(config, "type", ChartType.CircleSize50)
    if (!chartType.startsWith("spark")) {
      this.setState({ loadingMetrics: false })
      return
    }

    const chartDuration = getValue(config, "metric.duration", Duration.LastHour)
    const chartInterval = getValue(config, "metric.interval", getRecommendedInterval(chartDuration))
    const metricFunction = getValue(config, "metric.metricFunction", MetricFunctionType.Mean)
    const resourceType = getValue(config, "resource.type", "")

    // if it is spark chart, get metrics data
    const duration = getItem(chartDuration, DurationOptions)

    const extraTags = {}

    if (resourceType === ResourceType.Node) {
      const valueField = getValue(config, "resource.valueKey", "")
      const fieldName = getNodeMetricFieldName(valueField)
      extraTags[FIELD_NAME] = fieldName
    }

    const metricQuery = {
      global: {
        start: duration.value,
        window: chartInterval,
        functions: [metricFunction],
      },
      individual: [], // add id and metric type
    }
    this.setState((prevState) => {
      const { resources } = prevState
      // add id and metricType
      resources.forEach((r) => {
        // supports only for the below types
        if (
          r.metricType === MetricType.Gauge ||
          r.metricType === MetricType.GaugeFloat ||
          r.metricType === MetricType.Binary
        ) {
          metricQuery.individual.push({
            name: r.id,
            metricType: r.metricType,
            tags: { id: r.id, ...extraTags },
          })
        }
      })
      if (metricQuery.individual.length > 0) {
        api.metric
          .fetch(metricQuery)
          .then((metricRes) => {
            // update metrics data
            const keys = Object.keys(metricRes.data)
            const metrics = {}
            keys.forEach((key) => {
              const metricsRaw = metricRes.data[key]
              // update data into metrics object
              metrics[key] = getMetrics(metricsRaw, duration.tsFormat, metricFunction)
            })
            this.setState({ metrics: metrics, loadingMetrics: false })
          })
          .catch((_e) => {
            this.setState({ loadingMetrics: false })
          })
      } else {
        this.setState({ loadingMetrics: false })
      }
      return {}
    })
  }

  // updates resources data
  updateResourceData = () => {
    this.updateRefreshInterval()
    const { config } = this.props

    // if config empty, stop here
    if (config === undefined) {
      this.setState({ loading: false })
      return
    }

    const isMixedResources = getValue(config, "resource.isMixedResources", false)
    const resourceFilterType = getValue(config, "resource.filterType", ResourceFilterType.DetailedFilter)

    if (isMixedResources || resourceFilterType === ResourceFilterType.QuickID) {
      this.updateMixedResources(config)
    } else if (resourceFilterType === ResourceFilterType.DetailedFilter) {
      this.updateNonMixedResources(config)
    } else {
      this.setState({ loading: false })
    }
  }

  // update data for non mixed resources, detailed filter
  updateNonMixedResources = (config = {}) => {
    const resourceType = getValue(config, "resource.type", "")
    let itemsLimit = getValue(config, "resource.limit", 1)
    const resourceFilter = getValue(config, "resource.filter", {})

    const displayName = getValue(config, "resource.displayName", false)
    const resourceNameKey = getValue(config, "resource.nameKey", "undefined")
    const resourceValueKey = getValue(config, "resource.valueKey", "undefined")
    const resourceTimestampKey = getValue(config, "resource.timestampKey", "")

    const chartType = getValue(config, "type", ChartType.CircleSize50)

    if (chartType !== ChartType.Table) {
      itemsLimit = 1
    }

    // detailed filter
    const filters = []
    if (resourceFilter) {
      const keys = Object.keys(resourceFilter)
      keys.forEach((key) => {
        filters.push({ k: key, v: resourceFilter[key] })
      })
    }
    let resourceApi = null
    switch (resourceType) {
      case ResourceType.Gateway:
        resourceApi = api.gateway.list
        break

      case ResourceType.Node:
        resourceApi = api.node.list
        break

      case ResourceType.Source:
        resourceApi = api.source.list
        break

      case ResourceType.Field:
        resourceApi = api.field.list
        break

      default:
        this.setState({ loading: false })
        return
    }

    // update resource config
    const _resource = getValue(config, "resource", {})
    const _table = getValue(config, "table", {})
    const resourceConfig = { ..._resource, table: { ..._table } }

    resourceApi({ filter: filters, limit: itemsLimit })
      .then((res) => {
        const resourcesRaw = {}
        const formattedResources = res.data.data.map((resource, index) => {
          const formattedResource = this.getFormattedResource(
            resourceType,
            displayName,
            resourceNameKey,
            resourceValueKey,
            resourceTimestampKey,
            resource,
            resourceConfig
          )
          formattedResource.sortOrderPriority = `${formattedResource.name}-${index}`
          resourcesRaw[formattedResource.quickId] = resource
          return formattedResource
        })

        // push raw resources into redux
        this.props.loadData({ key: this.getWsKey(), resources: resourcesRaw })
        this.setState({ loading: false, resources: formattedResources }, () => {
          this.updateMetrics()
        })
      })
      .catch((_e) => {
        this.setState({ loading: false })
      })
  }

  // updates mixed resource and quick id data
  updateMixedResources = (config = {}) => {
    const isMixedResources = getValue(config, "resource.isMixedResources", false)

    const quickIds = []
    if (isMixedResources) {
      const resources = getValue(config, "resource.resources", [])
      resources.forEach((resource) => {
        const resourceType = getValue(resource, "type", "")
        const quickId = getValue(resource, "quickId", "")
        const qID = `${resourceType}:${quickId}`
        if (quickIds.indexOf(qID) === -1) {
          // add if it is not available
          quickIds.push(qID)
        }
      })
    } else {
      // quick id based filter
      const resourceType = getValue(config, "resource.type", "")
      const quickId = getValue(config, "resource.quickId", "")
      quickIds.push(`${resourceType}:${quickId}`)
    }

    api.quickId
      .getResources({ id: quickIds })
      .then((res) => {
        const resourcesRaw = {}
        const resourcesMap = res.data
        const resourceKeys = Object.keys(resourcesMap)
        const formattedResources = []
        const resourcesConfig = []

        if (isMixedResources) {
          const _resourcesConfig = getValue(config, "resource.resources", [])
          resourcesConfig.push(..._resourcesConfig)
        } else {
          // for quick id based resource
          const _resource = getValue(config, "resource", {})
          const _table = getValue(config, "table", {})
          resourcesConfig.push({ ..._resource, table: { ..._table } })
        }

        resourcesConfig.forEach((rc, index) => {
          const rType = rc.type
          const dName = getValue(rc, "displayName", false)
          const rNameKey = getValue(rc, "nameKey", "undefined")
          const rValueKey = getValue(rc, "valueKey", "undefined")
          const rTimestampKey = getValue(rc, "timestampKey", "")
          const sortOrderPriority = getValue(rc, "sortOrderPriority", `${index}`)

          for (let index = 0; index < resourceKeys.length; index++) {
            const resourceQuickID = resourceKeys[index]
            if (resourceQuickID === `${rc.type}:${rc.quickId}`) {
              const resource = resourcesMap[resourceQuickID]
              const formattedResource = this.getFormattedResource(
                rType,
                dName,
                rNameKey,
                rValueKey,
                rTimestampKey,
                resource,
                rc
              )

              formattedResource.sortOrderPriority = sortOrderPriority
              resourcesRaw[formattedResource.quickId] = resource
              formattedResources.push(formattedResource)
            }
          }
        })

        // push raw resources into redux
        this.props.loadData({ key: this.getWsKey(), resources: resourcesRaw })
        this.setState({ loading: false, resources: formattedResources }, () => {
          this.updateMetrics()
        })
      })
      .catch((_e) => {
        this.setState({ loading: false })
      })
  }

  getFormattedResource = (
    resourceType,
    displayName,
    resourceNameKey,
    resourceValueKey,
    resourceValueTimestampKey,
    resource,
    resourceConfig = {}
  ) => {
    const name = displayName
      ? resourceNameKey
        ? objectPath.get(resource, resourceNameKey, resourceNameKey)
        : ""
      : ""

    const quickId = getQuickId(resourceType, resource)

    // update metric type
    let metricType = objectPath.get(resource, "metricType", "")
    // for node resource type always fix metric type as gaugeFloat
    if (quickId.startsWith(ResourceType.Node)) {
      metricType = getNodeMetricType()
    }

    // formatted resource
    return {
      id: resource.id,
      resourceType: resourceType,
      quickId: quickId,
      name: name,
      metricType: metricType,
      value: String(objectPath.get(resource, resourceValueKey, "")),
      timestamp:
        resourceValueTimestampKey !== "" ? objectPath.get(resource, resourceValueTimestampKey, "") : "",
      key: {
        value: resourceValueKey,
        timestamp: resourceValueTimestampKey,
      },
      resourceConfig: { ...resourceConfig },
    }
  }

  render() {
    const { loading, loadingMetrics, resources, metrics } = this.state
    const { widgetId, dimensions, config, history } = this.props
    const columnDisplay = getValue(config, "chart.columnDisplay", false)
    const chartType = getValue(config, "type", ChartType.CircleSize75)
    const resourceType = getValue(config, "resource.type", "")

    if (loading) {
      return <Loading />
    }

    if (resources.length === 0) {
      return <span>No data available</span>
    }

    // update resource value
    const resourcesRaw = getValue(this.props.wsData, this.getWsKey(), {})
    Object.keys(resourcesRaw).forEach((qId) => {
      const res = resourcesRaw[qId]
      for (let index = 0; index < resources.length; index++) {
        const resource = resources[index]
        if (resource.quickId === qId) {
          resources[index].resourceType = resource.resourceType
          const rValueKey = getValue(resource, "key.value", "")
          const rTimestampKey = getValue(resource, "key.timestamp", "")
          if (rValueKey !== "") {
            resources[index].value = String(objectPath.get(res, rValueKey, ""))
          }
          if (rTimestampKey !== "") {
            resources[index].timestamp = objectPath.get(res, rTimestampKey, "")
          }
        }
      }
    })

    if (chartType === ChartType.Table) {
      return (
        <div className="mc-utilization-panel-item">
          <TableUtilization widgetId={widgetId} config={config} resources={resources} history={history} />
        </div>
      )
    }

    // if chartType not equal to table
    const charts = resources.map((resource, index) => {
      let chart = null
      switch (chartType) {
        case ChartType.SparkLine:
        case ChartType.SparkArea:
        case ChartType.SparkBar:
          chart = (
            <SparkLine
              key={"chart_" + index}
              widgetId={widgetId}
              dimensions={dimensions}
              config={config}
              resource={resource}
              metric={metrics[resource.id]}
              isMetricsLoading={loadingMetrics}
            />
          )
          break

        case ChartType.CircleSize50:
        case ChartType.CircleSize75:
        case ChartType.CircleSize100:
          chart = (
            <DonutUtilization
              key={"chart_" + index}
              widgetId={widgetId}
              config={config}
              resource={resource}
            />
          )
          break

        default:
          return <span>Type not implemented: {chartType}</span>
      }

      return (
        <div
          className="mc-utilization-panel-item"
          key={"chart_" + index}
          onClick={() => navigateToResource(resourceType, resource.id, history)}
          style={{ height: "100%", width: "100%", cursor: "pointer" }}
        >
          {chart}
        </div>
      )
    })

    const displayType = columnDisplay ? "grid" : "flex"

    return <div style={{ display: displayType }}>{charts}</div>
  }
}

UtilizationPanel.propTypes = {
  config: PropTypes.object,
}

// const config = {
//   resourceType: "",
//   resourceId: "",
//   resourceLabels: {},
//   resourceNameKey: "",
//   resourceValueKey: "",
//   resourceUnit: "",
//   thresholds: [],
//   columnGrid: false,
// }

const mapStateToProps = (state) => ({
  wsData: state.entities.websocket.data,
})

const mapDispatchToProps = (dispatch) => ({
  loadData: (data) => dispatch(loadData(data)),
  unloadData: (data) => dispatch(unloadData(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(UtilizationPanel)

// helper functions

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
