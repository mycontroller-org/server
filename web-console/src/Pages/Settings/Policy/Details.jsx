import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { RouteLink } from "../../../Components/Buttons/Buttons"
import { Labels, Statements } from "../../../Components/DataDisplay/Label"
import { DisplayTrue } from "../../../Components/DataDisplay/Miscellaneous"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"
import { routeMap as rMap } from "../../../Service/Routes"

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.policy.get}
      apiListTablesRecord={api.user.list}
      tableTitle="users"
      getTableFilterFunc={getTableFilterFuncImpl}
      tableColumns={tableColumns}
      getTableRowsFunc={getTableRowsFuncImpl}
      getDetailsFunc={getDetailsFuncImpl}
      cardTitle="details"
    />
  )
}

export default tabDetails

const getDetailsFuncImpl = (data) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "description", value: data.description })
  fieldsList1.push({ key: "system", value: <DisplayTrue data={data} field="system" /> })
  fieldsList1.push({ key: "labels", value: <Labels data={data.labels} /> })

  fieldsList2.push({ key: "modified_on", value: <LastSeen date={data.modifiedOn} tooltipPosition="top" /> })

  fieldsList2.push({ key: "statements", value: <Statements statements={data.statements} /> })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}

const tableColumns = [
  { title: "username", fieldKey: "username", sortable: true },
  { title: "full_name", fieldKey: "fullName", sortable: true },
  { title: "email", fieldKey: "email", sortable: true },
  { title: "disabled", fieldKey: "disabled", sortable: true },
  { title: "modified_on", fieldKey: "modifiedOn", sortable: true },
]

const getTableRowsFuncImpl = (rawData, _index, history) => {
  return [
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.settings.user.detail}
          id={rawData.id}
          text={rawData.username}
        />
      ),
    },
    { title: rawData.fullName },
    { title: rawData.email },
    { title: rawData.disabled ? "true" : "false" },
    { title: <LastSeen date={rawData.modifiedOn} /> },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return { policies: data.id }
}
