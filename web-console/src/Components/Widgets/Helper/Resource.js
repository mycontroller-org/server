import { ResourceType } from "../../../Constants/ResourcePicker"
import { redirect as rd, routeMap as rMap } from "../../../Service/Routes"

export const navigateToResource = (resourceType, id, history = null) => {
  if (history === null) {
    return
  }

  let route = null

  switch (resourceType) {
    case ResourceType.Gateway:
      route = rMap.resources.gateway.detail
      break

    case ResourceType.Node:
      route = rMap.resources.node.detail
      break

    case ResourceType.Source:
      route = rMap.resources.source.detail
      break

    case ResourceType.Field:
      route = rMap.resources.field.detail
      break

    case ResourceType.Task:
      route = rMap.operations.task.detail
      break

    case ResourceType.Schedule:
      route = rMap.operations.scheduler.detail
      break

    case ResourceType.DataRepository:
      route = rMap.resources.dataRepository.detail
      break

    default:
  }
  if (route !== null) {
    rd(history, route, { id: id })
  }
}
