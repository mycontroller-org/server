import { api } from "../../../../Service/Api"
import { getQuickId, ResourceType } from "../../../../Constants/ResourcePicker"
import { ControlType } from "../../../../Constants/Widgets/ControlPanel"
import { getValue } from "../../../../Util/Util"

export const getListAPI = (resourceType) => {
  switch (resourceType) {
    case ResourceType.Gateway:
      return api.gateway.list
    case ResourceType.Node:
      return api.node.list
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

export const getResource = (
  resourceType,
  resource,
  resourceNameKey,
  resourceTimestampKey = "",
  controlType = ControlType.SwitchToggle,
  keyPath = ""
) => {
  const quickId = getQuickId(resourceType, resource)
  const defaultLabel = controlType !== ControlType.MixedControl ? "undefined" : resourceNameKey
  const label = getValue(resource, resourceNameKey, defaultLabel)

  switch (resourceType) {
    case ResourceType.Gateway:
    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
      return {
        id: resource.id,
        type: resourceType,
        label: label,
        payload: resource.enabled,
        timestamp:
          resourceTimestampKey === "" ? resource.lastSeen : getValue(resource, resourceTimestampKey, ""),
        quickId: quickId,
        keyPath: keyPath,
      }

    case ResourceType.DataRepository:
      return {
        id: resource.id,
        type: resourceType,
        label: label,
        payload: keyPath !== "" ? getValue(resource, `data.${keyPath}`, "") : "",
        timestamp:
          resourceTimestampKey === "" ? resource.modifiedOn : getValue(resource, resourceTimestampKey, ""),
        quickId: quickId,
        keyPath: keyPath,
      }

    case ResourceType.Node:
      return {
        id: resource.id,
        type: resourceType,
        label: label,
        payload: keyPath !== "" ? getValue(resource, keyPath, "") : "",
        timestamp:
          resourceTimestampKey === "" ? resource.lastSeen : getValue(resource, resourceTimestampKey, ""),
        quickId: quickId,
        keyPath: keyPath,
      }

    case ResourceType.Field:
      return {
        id: resource.id,
        type: resourceType,
        label: label,
        payload: resource.current.value,
        timestamp:
          resourceTimestampKey === "" ? resource.noChangeSince : getValue(resource, resourceTimestampKey, ""),
        quickId: quickId,
        keyPath: keyPath,
      }

    default:
      return { id: "", label: "unknown resource type" }
  }
}
