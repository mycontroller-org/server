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
  const hideValueColumn = getValue(config, "table.hideValueColumn", false)
  const hideStatusColumn = getValue(config, "table.hideStatusColumn", false)
  const { t } = useTranslation()

  const columns = [{ title: t("name") }, { title: t("last_seen") }]

  // value column
  if (!hideValueColumn || isMixedResources) {
    columns.push({ title: t("value") })
  }

  // status column
  if (!hideStatusColumn || isMixedResources) {
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

  // sort by ascending order, by name
  resources.sort((a, b) => (a.sortOrderPriority > b.sortOrderPriority ? 1 : -1))

  resources.forEach((resource) => {
    let _resourceConfig = resource.resourceConfig

    // if asked to use custom table data take it
    if (isMixedResources && !resource.resourceConfig.useGlobal) {
      _resourceConfig.table = { ...resource.resourceConfig.table }
    }

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

    if (!_resourceConfig.table.hideValueColumn) {
      rowItems.push({ title: getResourceValue(resource, _resourceConfig) })
    } else if (isMixedResources) {
      rowItems.push(null)
    }

    if (!_resourceConfig.table.hideStatusColumn) {
      rowItems.push({ title: getStatus(resource, _resourceConfig) })
    } else if (isMixedResources) {
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
