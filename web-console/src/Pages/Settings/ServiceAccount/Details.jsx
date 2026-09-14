import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { Labels, Statements } from "../../../Components/DataDisplay/Label"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"

const TabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.serviceAccount.get}
      getDetailsFunc={getDetailsFuncImpl}
      cardTitle="details"
    />
  )
}

export default TabDetails

// helper functions

const getDetailsFuncImpl = (data) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "name", value: data.name })
  fieldsList1.push({ key: "username", value: data.username })
  fieldsList1.push({ key: "description", value: data.description })
  fieldsList1.push({ key: "never_expire", value: data.neverExpire ? "true" : "false" })
  fieldsList2.push({ key: "user_id", value: data.userId })
  fieldsList2.push({ key: "expires_on", value: <LastSeen date={data.expiresOn} tooltipPosition="top" /> })
  fieldsList2.push({ key: "created_on", value: <LastSeen date={data.createdOn} tooltipPosition="top" /> })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })
  if (Array.isArray(data.statements) && data.statements.length > 0) {
    fieldsList2.push({ key: "statements", value: <Statements statements={data.statements} /> })
  }

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}
