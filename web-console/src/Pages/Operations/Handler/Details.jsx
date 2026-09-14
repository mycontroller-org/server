import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { KeyValueMap, Labels } from "../../../Components/DataDisplay/Label"
import { api } from "../../../Service/Api"
import { DisplayTrue } from "../../../Components/DataDisplay/Miscellaneous"

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.handler.get}
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
  fieldsList1.push({ key: "description", value: data.description })
  fieldsList1.push({ key: "enabled", value: <DisplayTrue data={data} field="enabled" /> })
  fieldsList1.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList1.push({ key: "type", value: data.type })
  fieldsList1.push({ key: "state", value: <KeyValueMap data={data.state} /> })
  fieldsList2.push({ key: "spec", value: <KeyValueMap data={data.spec} /> })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}
