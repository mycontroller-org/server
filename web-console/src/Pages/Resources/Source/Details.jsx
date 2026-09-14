import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { RouteLink } from "../../../Components/Buttons/Buttons"
import { KeyValueMap, Labels } from "../../../Components/DataDisplay/Label"
import { LastSeen } from "../../../Components/Time/Time"
import InputField from "../../../Components/Widgets/ControlPanel/Common/InputField"
import { getQuickId, ResourceType } from "../../../Constants/ResourcePicker"
import { api } from "../../../Service/Api"
import { routeMap as rMap } from "../../../Service/Routes"
import { getFieldValue, getValue } from "../../../Util/Util"

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.source.get}
      apiListTablesRecord={api.field.list}
      tableTitle="fields"
      getTableFilterFunc={getTableFilterFuncImpl}
      tableColumns={tableColumns}
      getTableRowsFunc={getTableRowsFuncImpl}
      getDetailsFunc={getDetailsFuncImpl}
      cardTitle="details"
    />
  )
}

export default tabDetails

// helper functions

const getDetailsFuncImpl = (data) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "gateway_id", value: data.gatewayId })
  fieldsList1.push({ key: "node_id", value: data.nodeId })
  fieldsList1.push({ key: "source_id", value: data.sourceId })
  fieldsList1.push({ key: "name", value: data.name })
  fieldsList1.push({ key: "last_seen", value: <LastSeen date={data.lastSeen} tooltipPosition="top" /> })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList2.push({ key: "others", value: <KeyValueMap data={data.others} /> })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}

// Properties definition
const tableColumns = [
  { title: "field_id", fieldKey: "fieldId", sortable: true },
  { title: "name", fieldKey: "name", sortable: true },
  { title: "metric_type", fieldKey: "metricType", sortable: true },
  { title: "value", fieldKey: "current.value", sortable: true },
  { title: "previous_value", fieldKey: "previous.value", sortable: true },
  { title: "unit", fieldKey: "unit", sortable: true },
  { title: "last_seen", fieldKey: "lastSeen", sortable: true },
]

const getTableRowsFuncImpl = (rawData, _index, history) => {
  const isReadOnly = getValue(rawData, "labels.read_only", "false") === "true"
  const currentValue = isReadOnly ? (
    getFieldValue(getValue(rawData, "current.value", ""))
  ) : (
    <InputField
      payload={getValue(rawData, "current.value", "")}
      id={rawData.id}
      quickId={getQuickId(ResourceType.Field, rawData)}
      widgetId={rawData.id}
      key={rawData.id}
      sendPayloadWrapper={(callBack) => {
        callBack()
      }}
    />
  )

  return [
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.resources.field.detail}
          id={rawData.id}
          text={rawData.fieldId}
        />
      ),
    },
    {
      title: (
        <RouteLink history={history} path={rMap.resources.field.detail} id={rawData.id} text={rawData.name} />
      ),
    },
    rawData.metricType,
    { title: currentValue },
    { title: getFieldValue(getValue(rawData, "previous.value", "")) },
    rawData.unit,
    { title: <LastSeen date={rawData.lastSeen} /> },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return { gatewayId: data.gatewayId, nodeId: data.nodeId, sourceId: data.sourceId }
}
