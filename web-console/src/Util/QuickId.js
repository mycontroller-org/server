import { ResourceType } from "../Constants/ResourcePicker"

export const getQuickId = (resourceType = "", resource) => {
  switch (resourceType) {
    case ResourceType.Gateway:
    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
    case ResourceType.DataRepository:
      return `${resourceType}.${resource.id}`

    case ResourceType.Node:
      return `${resourceType}.${resource.gatewayId}.${resource.nodeId}`

    case ResourceType.Source:
      return `${resourceType}.${resource.gatewayId}.${resource.nodeId}.${resource.sourceId}`

    case ResourceType.Field:
      return `${resourceType}.${resource.gatewayId}.${resource.nodeId}.${resource.sourceId}.${resource.fieldId}`

    default:
      return `unknown resource type. ${resourceType}`
  }
}
