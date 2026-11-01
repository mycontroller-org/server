import { Button, Progress } from "@patternfly/react-core"
import { cellWidth, fitContent, Table, TableBody, TableHeader, TableVariant } from "@patternfly/react-table"
import React from "react"
import { useTranslation } from "react-i18next"
import v from "validator"
import { getValue } from "../../../../Util/Util"
import { LastSeen } from "../../../Time/Time"
import { navigateToResource } from "../../Helper/Resource"
import "./TableUtilization.scss"

const TableUtilization = ({ widgetId = "", config = {}, resources = [], history = null }) => {
  const isMixedResources = getValue(config, "resource.isMixedResources", false)
  const globalTable = getValue(config, "table", {})
  const { t } = useTranslation()

  // Mixed rows store "use global" on table.useGlobal. When it is on, the widget
  // table settings apply, including hide status/value. A row that opts out keeps
  // its own table settings. The column is rendered only when some row shows it.
  const resolveResourceConfig = (resource) => {
    const resourceConfig = resource.resourceConfig || {}
    if (!isMixedResources) {
      return {
        ...resourceConfig,
        table: resourceConfig.table || globalTable,
      }
    }
    const useGlobal = getValue(resourceConfig, "table.useGlobal", true)
    const table = useGlobal ? globalTable : resourceConfig.table || globalTable
    return { ...resourceConfig, table }
  }

  // sort by ascending order, by name
  resources.sort((a, b) => (a.sortOrderPriority > b.sortOrderPriority ? 1 : -1))

  const resolvedResources = resources.map((resource) => ({
    resource,
    resourceConfig: resolveResourceConfig(resource),
  }))

  const showValueColumn = resolvedResources.some(
    ({ resourceConfig }) => !getValue(resourceConfig, "table.hideValueColumn", false)
  )
  const showStatusColumn = resolvedResources.some(
    ({ resourceConfig }) => !getValue(resourceConfig, "table.hideStatusColumn", false)
  )

  const columns = [{ title: t("name") }, { title: t("last_seen") }]

  if (showValueColumn) {
    columns.push({ title: t("value") })
  }

  if (showStatusColumn) {
    columns.push({ title: t("status"), transforms: [cellWidth(35), fitContent] })
  }

  const rows = []

  const getResourceValue = (resource = {}, resourceConfig = {}) => {
    const displayValueFloat = parseFloat(resource.value)
    let displayValue = resource.value
    if (!Number.isNaN(displayValueFloat) && v.isFloat(String(displayValue), {})) {
      displayValue = displayValueFloat.toFixed(resourceConfig.roundDecimal)
    }
    if (resourceConfig.unit && resourceConfig.unit !== "") {
      return `${displayValue} ${resourceConfig.unit}`
    }
    return displayValue
  }

  const getStatus = (resource = {}, resourceConfig = {}) => {
    // get color
    let thresholdColor = "#06c"
    getFormattedThresholds(resourceConfig.table.thresholds).forEach((threshold) => {
      if (resource.value >= threshold.value) {
        thresholdColor = threshold.color
      }
    })
    return (
      <Progress
        key={"resource_" + resource.quickId}
        className="table-progress-bar"
        aria-label="utilization_table_status"
        style={{
          "--pf-c-progress__indicator--BackgroundColor": thresholdColor,
          "--pf-c-progress__bar--before--BackgroundColor": thresholdColor,
        }}
        value={resource.value}
        min={resourceConfig.table.minimumValue}
        max={resourceConfig.table.maximumValue}
        measureLocation={resourceConfig.table.displayStatusPercentage ? "inside" : "none"}
      />
    )
  }

  resolvedResources.forEach(({ resource, resourceConfig }) => {
    const rowItems = [
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={() => navigateToResource(resource.resourceType, resource.id, history)}
          >
            {resource.name !== "" ? resource.name : "undefined"}
          </Button>
        ),
      },
      {
        title: (
          <span className="gauge-value-timestamp">
            <LastSeen date={resource.timestamp} tooltipPosition="top" />
          </span>
        ),
      },
    ]

    if (!getValue(resourceConfig, "table.hideValueColumn", false)) {
      rowItems.push({ title: getResourceValue(resource, resourceConfig) })
    } else if (showValueColumn) {
      rowItems.push(null)
    }

    if (!getValue(resourceConfig, "table.hideStatusColumn", false)) {
      rowItems.push({ title: getStatus(resource, resourceConfig) })
    } else if (showStatusColumn) {
      rowItems.push(null)
    }

    rows.push(rowItems)
  })

  if (rows.length === 0) {
    return <span>No resource found</span>
  }

  return (
    <Table
      key={"utilization_table_" + widgetId}
      className="mc-utilization-table ut-table"
      aria-label="Utilization Table"
      variant={TableVariant.compact}
      borders={!config.hideBorder}
      cells={columns}
      rows={rows}
    >
      <TableHeader />
      <TableBody />
    </Table>
  )
}

export default TableUtilization

// helper functions
const getFormattedThresholds = (thresholds = {}) => {
  const thresholdKeys = Object.keys(thresholds)
  return thresholdKeys.map((key) => {
    return { value: parseFloat(key), color: thresholds[key] }
  })
}
