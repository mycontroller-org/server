// Resource type mapping
export const ResourceType = {
  Gateway: "gateway",
  Node: "node",
  Source: "source",
  Field: "field",
}

// Resource type options list
export const ResourceTypeOptions = [
  { value: ResourceType.Gateway, label: "gateway" },
  { value: ResourceType.Node, label: "node" },
  { value: ResourceType.Source, label: "source" },
  { value: ResourceType.Field, label: "field" },
]

// Resource Filter type mapping
export const ResourceFilterType = {
  QuickID: "quick_id",
  DetailedFilter: "detailed_filter",
}

// Resource Filter type options list
export const ResourceFilterTypeOptions = [
  { value: ResourceFilterType.QuickID, label: "quick_id" },
  { value: ResourceFilterType.DetailedFilter, label: "detailed_filter" },
]
