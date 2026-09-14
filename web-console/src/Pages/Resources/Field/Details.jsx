import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { KeyValueMap, Labels } from "../../../Components/DataDisplay/Label"
import { DisplayFieldValue } from "../../../Components/DataDisplay/Miscellaneous"
import { LastSeen } from "../../../Components/Time/Time"
import { MetricTypeOptions } from "../../../Constants/Metric"
import { api } from "../../../Service/Api"
import { getItem, getValue } from "../../../Util/Util"
import { useTranslation } from "react-i18next"

const TabDetails = ({ resourceId, history }) => {
  const { t } = useTranslation()
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.field.get}
      getDetailsFunc={(data) => getDetailsFuncImpl(data, t)}
      showMetrics
    />
  )
}

export default TabDetails

// helper functions

const getDetailsFuncImpl = (data, t) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "gateway_id", value: data.gatewayId })
  fieldsList1.push({ key: "node_id", value: data.nodeId })
  fieldsList1.push({ key: "source_id", value: data.sourceId })
  fieldsList1.push({ key: "field_id", value: data.fieldId })
  fieldsList1.push({ key: "name", value: data.name })
  fieldsList1.push({ key: "last_seen", value: <LastSeen date={data.lastSeen} tooltipPosition="top" /> })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList2.push({ key: "others", value: <KeyValueMap data={data.others} /> })
  fieldsList2.push({ key: "metric_type", value: t(getItem(data.metricType, MetricTypeOptions).label) })
  fieldsList2.push({ key: "unit", value: data.unit })
  fieldsList2.push({
    key: "no_change_since",
    value: <LastSeen date={data.noChangeSince} tooltipPosition="top" />,
  })
  fieldsList2.push({
    key: "value",
    value: (
      <DisplayFieldValue
        value={getValue(data, "current.value", "")}
        timestamp={getValue(data, "current.timestamp", "")}
      />
    ),
  })
  fieldsList2.push({
    key: "previous_value",
    value: (
      <DisplayFieldValue
        value={getValue(data, "previous.value", "")}
        timestamp={getValue(data, "previous.timestamp", "")}
      />
    ),
  })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}
