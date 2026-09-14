import React from "react"
import { useTranslation } from "react-i18next"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { RouteLink } from "../../../Components/Buttons/Buttons"
import { Labels } from "../../../Components/DataDisplay/Label"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"
import { routeMap as rMap } from "../../../Service/Routes"

const TabDetails = ({ resourceId, history }) => {
  const { t } = useTranslation()
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.user.get}
      apiListTablesRecord={api.serviceAccount.list}
      tableTitle="service_accounts"
      getTableFilterFunc={getTableFilterFuncImpl}
      tableColumns={tableColumns}
      getTableRowsFunc={getTableRowsFuncImpl}
      getDetailsFunc={(data) => getDetailsFuncImpl(data, t)}
      cardTitle="details"
    />
  )
}

export default TabDetails

const getDetailsFuncImpl = (data, t) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "username", value: data.username })
  fieldsList1.push({ key: "full_name", value: data.fullName })
  fieldsList1.push({ key: "email", value: data.email })
  fieldsList2.push({ key: "disabled", value: data.disabled ? t("true") : t("false") })
  fieldsList2.push({
    key: "policies",
    value: Array.isArray(data.policies) ? data.policies.join(", ") : "",
  })
  fieldsList2.push({ key: "modified_on", value: <LastSeen date={data.modifiedOn} tooltipPosition="top" /> })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}

const tableColumns = [
  { title: "name", fieldKey: "name", sortable: true },
  { title: "description", fieldKey: "description", sortable: true },
  { title: "never_expire", fieldKey: "neverExpire", sortable: true },
  { title: "expires_on", fieldKey: "expiresOn", sortable: true },
  { title: "created_on", fieldKey: "createdOn", sortable: true },
]

const getTableRowsFuncImpl = (rawData, _index, history) => {
  return [
    {
      title: (
        <RouteLink
          history={history}
          path={rMap.settings.serviceAccount.detail}
          id={rawData.id}
          text={rawData.name}
        />
      ),
    },
    { title: rawData.description },
    { title: rawData.neverExpire ? "true" : "false" },
    { title: <LastSeen date={rawData.expiresOn} /> },
    { title: <LastSeen date={rawData.createdOn} /> },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return { userId: data.id }
}
