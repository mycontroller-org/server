import { api } from "../../../Service/Api"
import { ResourceType, FieldDataType } from "../../../Constants/ResourcePicker"
import React from "react"
import { TextInput } from "@patternfly/react-core"
import ResourcePicker from "./ResourcePicker"
import { getDynamicFilter } from "../../../Util/Filter"

export const getResourceType = (quickId) => {
  if (quickId === "") {
    return ["", ""]
  }
  const resource = quickId.split(":", 1)[0]
  const id = quickId.substring(resource.length + 1)
  switch (quickId.split(":", 1)[0]) {
    case ResourceType.Gateway:
    case ResourceType.GatewayShort:
      return [ResourceType.Gateway, id]

    case ResourceType.Node:
    case ResourceType.NodeShort:
      return [ResourceType.Node, id]

    case ResourceType.Source:
    case ResourceType.SourceShort:
      return [ResourceType.Source, id]

    case ResourceType.Field:
    case ResourceType.FieldShort:
      return [ResourceType.Field, id]

    case ResourceType.Task:
    case ResourceType.TaskShort:
      return [ResourceType.Task, id]

    case ResourceType.Schedule:
    case ResourceType.ScheduleShort:
      return [ResourceType.Schedule, id]

    case ResourceType.Handler:
    case ResourceType.HandlerShort:
      return [ResourceType.Handler, id]

    case ResourceType.DataRepository:
    case ResourceType.DataRepositoryShort:
      return [ResourceType.DataRepository, id]

    default:
      return ["", quickId]
  }
}

export const getResourceOptionsAPI = (resourceType) => {
  switch (resourceType) {
    case ResourceType.Gateway:
      return api.gateway.list

    case ResourceType.Node:
      return api.node.list

    case ResourceType.Source:
      return api.source.list

    case ResourceType.Field:
      return api.field.list

    case ResourceType.Task:
      return api.task.list

    case ResourceType.Schedule:
      return api.schedule.list

    case ResourceType.Handler:
      return api.handler.list

    case ResourceType.DataRepository:
      return api.dataRepository.list

    default:
      return () => {}
  }
}

export const getResourceOptionValueFunc = (resourceType) => {
  switch (resourceType) {
    case ResourceType.Gateway:
    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
      return (item) => {
        return item.id
      }

    case ResourceType.Node:
      return (item) => {
        return item.gatewayId + "." + item.nodeId
      }

    case ResourceType.Source:
      return (item) => {
        return item.gatewayId + "." + item.nodeId + "." + item.sourceId
      }

    case ResourceType.Field:
      return (item) => {
        return item.gatewayId + "." + item.nodeId + "." + item.sourceId + "." + item.fieldId
      }

    default:
      return null
  }
}

export const getResourceFilterFunc = (resourceType) => {
  switch (resourceType) {
    case ResourceType.Gateway:
    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
    case ResourceType.DataRepository:
      return (value) => {
        return getDynamicFilter("id", value, [])
      }

    case ResourceType.Node:
    case ResourceType.Source:
    case ResourceType.Field:
      return (value) => {
        return getDynamicFilter("name", value, [])
      }

    default:
      return null
  }
}

export const getOptionsDescriptionFunc = (resourceType) => {
  switch (resourceType) {
    case ResourceType.Gateway:
      return (item) => {
        return item.name
      }

    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
    case ResourceType.DataRepository:
      return (item) => {
        return item.description
      }

    case ResourceType.Node:
    case ResourceType.Source:
    case ResourceType.Field:
      return (item) => {
        return item.name
      }

    default:
      return null
  }
}

// returns variable value to display on variables list
export const getResourceDisplayValue = (index, item = {}, _onChange, validated, _isDisabled) => {
  let textValue = `${item.type}`
  switch (item.type) {
    case FieldDataType.TypeString:
      textValue += `: ${item.value}`
      break

    case FieldDataType.TypeTelegram:
      textValue += `: ${item.text}`
      break

    case FieldDataType.TypeResourceByQuickId:
      textValue += `: ${item.resourceType}:${item.quickId}`
      // include payload, if available
      if (item.payload) {
        textValue += ` => ${item.payload}`
      }
      break

    case FieldDataType.TypeResourceByLabels:
      textValue += `: ${item.resourceType}:${Object.keys(item.labels)
        .map((key) => {
          return `${key}=${item.labels[key]}`
        })
        .join(",")}`
      // include payload, if available
      if (item.payload) {
        textValue += ` => ${item.payload}`
      }
      break

    case FieldDataType.TypeBackup:
      textValue += `: ${item.prefix}, ${item.targetDirectory}, ${item.storageExportType}`
      break

    case FieldDataType.TypeEmail:
      textValue += `: ${item.subject}`
      break

    case FieldDataType.TypeWebhook:
      textValue += `: ${item.server}`
      break
  }

  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={textValue}
      isDisabled={true}
      validated={validated}
    />
  )
}

// calls the model to update the resource details
export const resourceUpdateButtonCallback = (callerType, index = 0, item = {}, onChange) => {
  return (
    <ResourcePicker
      key={"picker_" + index}
      value={item.value}
      name={item.key}
      id={"model_" + index}
      callerType={callerType}
      onChange={onChange}
    />
  )
}
