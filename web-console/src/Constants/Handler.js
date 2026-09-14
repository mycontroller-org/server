// Handler type values
export const HandlerType = {
  Resource: "resource",
  Email: "email",
  Telegram: "telegram",
  Webhook: "webhook",
  Backup: "backup",
}

// Handler type options
export const HandlerTypeOptions = [
  {
    value: HandlerType.Resource,
    label: "opts.handler_type.label_resource",
    description: "opts.handler_type.desc_resource",
  },
  {
    value: HandlerType.Email,
    label: "opts.handler_type.label_email",
    description: "opts.handler_type.desc_email",
  },
  {
    value: HandlerType.Telegram,
    label: "opts.handler_type.label_telegram",
    description: "opts.handler_type.desc_telegram",
  },
  {
    value: HandlerType.Webhook,
    label: "opts.handler_type.label_webhook",
    description: "opts.handler_type.desc_webhook",
  },
  {
    value: HandlerType.Backup,
    label: "opts.handler_type.label_backup",
    description: "opts.handler_type.desc_backup",
  },
]
