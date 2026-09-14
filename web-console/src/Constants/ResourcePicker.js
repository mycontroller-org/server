export const CallerType = {
  Variable: "variable",
  Parameter: "parameter",
}

export const FieldDataType = {
  TypeString: "string",
  TypeEmail: "email",
  TypeTelegram: "telegram",
  TypeWebhook: "webhook",
  TypeResourceByQuickId: "resource_by_quick_id",
  TypeResourceByLabels: "resource_by_labels",
  TypeBackup: "backup",
}

export const FieldDataTypeOptions = [
  { value: FieldDataType.TypeString, label: "string" },
  { value: FieldDataType.TypeResourceByQuickId, label: "resource_by_quick_id" },
  { value: FieldDataType.TypeResourceByLabels, label: "resource_by_labels" },
  { value: FieldDataType.TypeWebhook, label: "webhook" },
  { value: FieldDataType.TypeEmail, label: "email" },
  { value: FieldDataType.TypeTelegram, label: "telegram" },
  { value: FieldDataType.TypeBackup, label: "backup" },
]

export const ResourceType = {
  Gateway: "gateway",
  Node: "node",
  Source: "source",
  Field: "field",
  Task: "task",
  Schedule: "schedule",
  Handler: "handler",
  DataRepository: "data_repository",
}

export const ResourceTypeOptions = [
  { value: ResourceType.Gateway, label: "gateway" },
  { value: ResourceType.Node, label: "node" },
  { value: ResourceType.Source, label: "source" },
  { value: ResourceType.Field, label: "field" },
  { value: ResourceType.Task, label: "task" },
  { value: ResourceType.Schedule, label: "schedule" },
  { value: ResourceType.Handler, label: "handler" },
  { value: ResourceType.DataRepository, label: "data_repository" },
]

export const getQuickId = (resourceType = "", resource = {}) => {
  switch (resourceType) {
    case ResourceType.Gateway:
    case ResourceType.Task:
    case ResourceType.Schedule:
    case ResourceType.Handler:
    case ResourceType.DataRepository:
      return `${resourceType}:${resource.id}`

    case ResourceType.Node:
      return `${resourceType}:${resource.gatewayId}.${resource.nodeId}`

    case ResourceType.Source:
      return `${resourceType}:${resource.gatewayId}.${resource.nodeId}.${resource.sourceId}`

    case ResourceType.Field:
      return `${resourceType}:${resource.gatewayId}.${resource.nodeId}.${resource.sourceId}.${resource.fieldId}`

    default:
      return `unknown resource type:${resourceType}`
  }
}

export const TelegramParseMode = {
  Text: "Text",
  Markdown: "Markdown",
  MarkdownV2: "MarkdownV2",
  HTML: "HTML",
}

export const TelegramParseModeOptions = [
  { value: TelegramParseMode.Text, label: "text" },
  { value: TelegramParseMode.Markdown, label: "markdown" },
  { value: TelegramParseMode.MarkdownV2, label: "markdown_v2" },
  { value: TelegramParseMode.HTML, label: "html" },
]

// Exporter type values
export const BackupProviderType = {
  Disk: "disk",
}

// Exporter type options
export const BackupProviderTypeOptions = [
  {
    value: BackupProviderType.Disk,
    label: "opts.backup_provider.label_disk",
    description: "opts.backup_provider.desc_disk",
  },
]

// Storage Export type values
export const StorageExportType = {
  YAML: "yaml",
  JSON: "json",
}

// Storage Export type options
export const StorageExportTypeOptions = [
  { value: StorageExportType.YAML, label: "yaml" },
  { value: StorageExportType.JSON, label: "json" },
]

// Webhook method type values
export const WebhookMethodType = {
  GET: "GET",
  POST: "POST",
  PUT: "PUT",
  DELETE: "DELETE",
}

// Webhook method type options
export const WebhookMethodTypeOptions = [
  { value: WebhookMethodType.GET, label: "opts.http_method.label_get" },
  { value: WebhookMethodType.POST, label: "opts.http_method.label_post" },
  { value: WebhookMethodType.PUT, label: "opts.http_method.label_put" },
  { value: WebhookMethodType.DELETE, label: "opts.http_method.label_delete" },
]
