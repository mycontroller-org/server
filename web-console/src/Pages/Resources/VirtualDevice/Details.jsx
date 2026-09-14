import React from "react"
import TabDetailsBase from "../../../Components/BasePage/TabDetailsBase"
import { Labels } from "../../../Components/DataDisplay/Label"
import { getStatusBool } from "../../../Components/Icons/Icons"
import { LastSeen } from "../../../Components/Time/Time"
import { api } from "../../../Service/Api"
import { getValue } from "../../../Util/Util"

const updateTraits = (resourceId, filter = { offset: 0, limit: 10 }) =>
  new Promise(function (resolve, reject) {
    api.virtualDevice
      .get(resourceId)
      .then((res) => {
        const traitsObj = getValue(res, "data.traits", {})
        const traits = Object.keys(traitsObj)
        traits.sort()
        let start = filter.offset
        let end = start + filter.limit
        if (start > traits.length) {
          start = 0
        }
        if (end > traits.length) {
          end = traits.length
        }
        const response = traits.slice(start, end).map((k) => {
          return { trait: k, resource: traitsObj[k] }
        })
        resolve({ data: { count: traits.length, limit: traits.length, offset: 0, data: response } })
      })
      .catch((e) => {
        reject(e)
      })
  })

const tabDetails = ({ resourceId, history }) => {
  return (
    <TabDetailsBase
      resourceId={resourceId}
      history={history}
      apiGetRecord={api.virtualDevice.get}
      apiListTablesRecord={(filter) => updateTraits(resourceId, filter)}
      tableTitle="traits"
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
  fieldsList1.push({ key: "name", value: data.name })
  fieldsList1.push({ key: "description", value: data.description })
  fieldsList1.push({ key: "enabled", value: getStatusBool(data.enabled) })
  fieldsList2.push({ key: "type", value: data.deviceType })
  fieldsList2.push({ key: "labels", value: <Labels data={data.labels} /> })
  fieldsList1.push({ key: "modified_on", value: <LastSeen date={data.modifiedOn} tooltipPosition="top" /> })

  return {
    "list-1": fieldsList1,
    "list-2": fieldsList2,
  }
}

// Properties definition
const tableColumns = [
  { title: "name", fieldKey: "name", sortable: false },
  { title: "trait", fieldKey: "id", sortable: false },
  { title: "resource", fieldKey: "resource", sortable: false },
  { title: "labels", fieldKey: "labels", sortable: false },
]

const getTableRowsFuncImpl = (rawData, _index, _history) => {
  const resource = getValue(rawData, "resource", {})
  return [
    { title: resource.name },
    { title: resource.traitType },
    { title: `${resource.resourceType}:${resource.quickId}` },
    { title: JSON.stringify(resource.labels) },
  ]
}

const getTableFilterFuncImpl = (data) => {
  return {}
}
