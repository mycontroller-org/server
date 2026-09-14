import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { RouteLink } from "../../../Components/Buttons/Buttons"
import { KeyValueMap, Labels } from "../../../Components/DataDisplay/Label"
import { getStatus } from "../../../Components/Icons/Icons"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"
import { routeMap as rMap } from "../../../Service/Routes"
import { getValue } from "../../../Util/Util"

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.virtualAssistant.get}
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
  fieldsList1.push({ key: "description", value: data.description })
  fieldsList1.push({ key: "enabled", value: data.enabled ? "true" : "false" })
  fieldsList1.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList2.push({ key: "provider", value: getValue(data, "providerType", "") })
  fieldsList2.push({ key: "device_filter", value: <Labels data={data.deviceFilter} /> })
  fieldsList2.push({ key: "status", value: getStatus(data.state ? data.state.status : "unavailable") })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}

const tableColumns = [
  { title: "node_id", fieldKey: "nodeId", sortable: true },
  { title: "name", fieldKey: "name", sortable: true },
  { title: "version", fieldKey: "labels.version", sortable: true },
  { title: "library_version", fieldKey: "labels.library_version", sortable: true },
  { title: "battery", fieldKey: "others.battery_level", sortable: true },
  { title: "status", fieldKey: "state.status", sortable: true },
  { title: "last_seen", fieldKey: "lastSeen", sortable: true },
]

const getTableRowsFuncImpl = (rawData, _index, history) => {
  return [
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.resources.node.detail}
          id={rawData.id}
          text={rawData.nodeId}
        />
      ),
    },
    {
      title: (
        <RouteLink history={history} path={rMap.resources.node.detail} id={rawData.id} text={rawData.name} />
      ),
    },
    rawData.labels.version,
    rawData.labels.library_version,
    { title: getValue(rawData, "others.battery_level", "") },
    { title: getStatus(rawData.state.status) },
    { title: <LastSeen date={rawData.lastSeen} /> },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return { gatewayId: data.id }
}
