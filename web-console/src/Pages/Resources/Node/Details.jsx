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
      apiGetRecord={api.node.get}
      apiListTablesRecord={api.source.list}
      tableTitle={"sources"}
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
  fieldsList1.push({ key: "name", value: data.name })
  fieldsList1.push({ key: "status", value: getStatus(getValue(data, "state.status", "")) })
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
  { title: "source_id", fieldKey: "sourceId", sortable: true },
  { title: "name", fieldKey: "name", sortable: true },
  { title: "last_seen", fieldKey: "lastSeen", sortable: true },
]

const getTableRowsFuncImpl = (rawData, _index, history) => {
  return [
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.resources.source.detail}
          id={rawData.id}
          text={rawData.sourceId}
        />
      ),
    },
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.resources.source.detail}
          id={rawData.id}
          text={rawData.name}
        />
      ),
    },
    { title: <LastSeen date={rawData.lastSeen} /> },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return { gatewayId: data.gatewayId, nodeId: data.nodeId }
}
