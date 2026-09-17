// Hierarchical resource id shown in the UI and CLI, without a type prefix:
//   gateway  -> gw1
//   node     -> gw1.node1
//   source   -> gw1.node1.source1
//   field    -> gw1.node1.source1.field1
export const resourceQuickId = (kind, resource = {}) => {
  if (!resource) {
    return ""
  }
  switch (kind) {
    case "gateway":
      return String(resource.id || "")
    case "node":
      return [resource.gatewayId, resource.nodeId].filter(partPresent).join(".")
    case "source":
      return [resource.gatewayId, resource.nodeId, resource.sourceId].filter(partPresent).join(".")
    case "field":
      return [resource.gatewayId, resource.nodeId, resource.sourceId, resource.fieldId]
        .filter(partPresent)
        .join(".")
    default:
      return ""
  }
}

const partPresent = (value) => value !== undefined && value !== null && String(value) !== ""
