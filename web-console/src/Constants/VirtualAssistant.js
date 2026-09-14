// Provider keeps supported providers list, this value will be used in backend
export const Provider = {
  GoogleAssistant: "google_assistant",
  Alexa: "alexa_assistant",
}

// Providers options list
export const ProviderOptions = [
  {
    value: Provider.GoogleAssistant,
    label: "opts.va_provider.label_google_assistant",
    description: "opts.va_provider.desc_google_assistant",
  },
  {
    value: Provider.Alexa,
    label: "opts.va_provider.label_alexa",
    description: "opts.va_provider.desc_alexa",
  },
]
