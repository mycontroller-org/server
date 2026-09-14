import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { Labels } from "../../../Components/DataDisplay/Label"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId ? resourceId : "none"}
      history={history}
      apiGetRecord={api.auth.profile}
      getDetailsFunc={getDetailsFuncImpl}
    />
  )
}

export default tabDetails

// helper functions

const getDetailsFuncImpl = (data) => {
  const fieldsList1 = []
  const fieldsList2 = []

  fieldsList1.push({ key: "id", value: data.id })
  fieldsList1.push({ key: "username", value: data.username })
  fieldsList1.push({ key: "email", value: data.email })
  fieldsList1.push({ key: "full_name", value: data.fullName })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList2.push({
    key: "modified_on",
    value: <LastSeen date={data.modifiedOn} tooltipPosition="top" />,
  })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}
