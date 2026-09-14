// Provider keeps supported providers list, this value will be used in backend
export const Provider = {
  ESPHome: "esphome",
  Generic: "generic",
  MySensorsV2: "mysensors_v2",
  PhilipsHue: "philips_hue",
  SystemMonitoring: "system_monitoring",
  Tasmota: "tasmota",
}

// Providers options list
export const ProviderOptions = [
  {
    value: Provider.ESPHome,
    label: "opts.gw_provider.label_esphome",
    description: "opts.gw_provider.desc_esphome",
  },
  {
    value: Provider.Generic,
    label: "opts.gw_provider.label_generic",
    description: "opts.gw_provider.desc_generic",
  },
  {
    value: Provider.MySensorsV2,
    label: "opts.gw_provider.label_mysensors_v2",
    description: "opts.gw_provider.desc_mysensors_v2",
  },
  {
    value: Provider.PhilipsHue,
    label: "opts.gw_provider.label_philips_hue",
    description: "opts.gw_provider.desc_philips_hue",
  },
  {
    value: Provider.SystemMonitoring,
    label: "opts.gw_provider.label_system_monitoring",
    description: "opts.gw_provider.desc_system_monitoring",
  },
  {
    value: Provider.Tasmota,
    label: "opts.gw_provider.label_tasmota",
    description: "opts.gw_provider.desc_tasmota",
  },
]

// Protocol values
export const Protocol = {
  MQTT: "mqtt",
  Serial: "serial",
  Ethernet: "ethernet",
  HTTP: "http",
}

// Protocol options list
export const ProtocolOptions = [
  {
    value: Protocol.MQTT,
    label: "opts.gw_protocol.label_mqtt",
    description: "opts.gw_protocol.desc_mqtt",
  },
  {
    value: Protocol.Serial,
    label: "opts.gw_protocol.label_serial",
    description: "opts.gw_protocol.desc_serial",
  },
  {
    value: Protocol.Ethernet,
    label: "opts.gw_protocol.label_ethernet",
    description: "opts.gw_protocol.desc_ethernet",
  },
  {
    value: Protocol.HTTP,
    label: "opts.gw_protocol.label_http",
    description: "opts.gw_protocol.desc_http",
  },
]

export const filterProtocolOptions = (providerType) => {
  const protocols = []
  switch (providerType) {
    case Provider.MySensorsV2:
      protocols.push(Protocol.MQTT, Protocol.Serial, Protocol.Ethernet)
      break

    case Provider.Tasmota:
      protocols.push(Protocol.MQTT)
      break

    case Provider.Generic:
      protocols.push(Protocol.MQTT, Protocol.HTTP)
      break

    default:
      break
  }
  const protocolOptions = []
  ProtocolOptions.forEach((option) => {
    if (protocols.includes(option.value)) {
      protocolOptions.push(option)
    }
  })

  return protocolOptions
}

// Message logger values
export const MessageLogger = {
  None: "none",
  FileLogger: "file_logger",
}

// Message logger options list
export const MessageLoggerOptions = [
  {
    value: MessageLogger.None,
    label: "opts.gw_msg_logger.label_none",
    description: "opts.gw_msg_logger.desc_none",
  },
  {
    value: MessageLogger.FileLogger,
    label: "opts.gw_msg_logger.label_file_logger",
    description: "opts.gw_msg_logger.desc_file_logger",
  },
]
