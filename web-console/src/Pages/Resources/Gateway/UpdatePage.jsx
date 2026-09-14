import objectPath from "object-path"
import React from "react"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import {
  filterProtocolOptions,
  MessageLogger,
  MessageLoggerOptions,
  Protocol,
  Provider,
  ProviderOptions,
} from "../../../Constants/Gateway"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import { getESPHomeItems } from "./EspHome/Update"
import { getPhilipsHueItems } from "./PhilipsHue/Update"
import { getSystemMonitoringItems } from "./SystemMonitoring/Update"
import { getGenericNodeEndpointItems, getGenericProviderScript } from "./Generic/Update"
import { getHttpGenericProtocolItems } from "./Generic/Endpoints"

class UpdatePage extends React.Component {
  render() {
    const { id } = this.props.match.params
    const { cancelFn = () => {} } = this.props

    const isNewEntry = id === undefined || id === ""

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.gateway.get}
        apiSaveRecord={api.gateway.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.gateway.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.gateway.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_gateway" />
          <PageContent hasNoPaddingTop>{editor} </PageContent>
        </>
      )
    }
    return editor
  }
}

export default UpdatePage

// support functions

const getFormItems = (rootObject, id) => {
  objectPath.set(rootObject, "reconnectDelay", "30s", true)
  objectPath.set(rootObject, "queueFailedMessage", false, true)
  objectPath.set(rootObject, "labels", { location: "server" }, true)
  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: id ? true : false,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_id",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isID: {}, isNotEmpty: {} },
    },
    {
      label: "description",
      fieldId: "description",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "enabled",
      fieldId: "enabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "reconnect_delay",
      fieldId: "reconnectDelay",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "queue_failed_messages",
      fieldId: "queueFailedMessage",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "labels",
      fieldId: "!labels",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "labels",
      fieldType: FieldType.Labels,
      dataType: DataType.Object,
      value: {},
      validator: { isLabel: {} },
    },
    {
      label: "provider",
      fieldId: "!provider.type",
      fieldType: FieldType.Divider,
    },
    {
      label: "type",
      fieldId: "provider.type",
      isRequired: true,
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: ProviderOptions,
      validated: "error",
      value: "",
    },
  ]

  const providerType = objectPath.get(rootObject, "provider.type", "").toLowerCase()
  const protocolType = objectPath.get(rootObject, "provider.protocol.type", "").toLowerCase()

  switch (providerType) {
    case Provider.MySensorsV2:
      const mySensorsItems = getMySensorsItems(rootObject)
      items.push(...mySensorsItems)
      break

    case Provider.Tasmota: // set tasmota defaults
      objectPath.set(rootObject, "provider.protocol.type", Protocol.MQTT, true)
      break

    default:
      break
  }

  // if there is provider selected, common details
  if (
    providerType !== "" &&
    providerType !== Provider.SystemMonitoring &&
    providerType !== Provider.PhilipsHue &&
    providerType !== Provider.ESPHome
  ) {
    items.push(
      {
        label: "protocol",
        fieldId: "!provider.protocol.type",
        fieldType: FieldType.Divider,
      },
      {
        label: "type",
        fieldId: "provider.protocol.type",
        isRequired: true,
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        options: filterProtocolOptions(providerType),
        value: "",
        resetFields: {
          resetProtocol: (ro, newValue) => {
            objectPath.set(ro, "provider.protocol", { type: newValue })
          },
        },
      }
    )

    if (protocolType !== "" && protocolType !== Protocol.HTTP) {
      objectPath.set(rootObject, "provider.protocol.transmitPreDelay", "10ms", true)
      items.push({
        label: "transmit_pre_delay",
        fieldId: "provider.protocol.transmitPreDelay",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
      })
    }

    // protocol fields
    switch (protocolType) {
      case Protocol.MQTT:
        const protocolMqttItems = getProtocolMqttItems(rootObject, providerType)
        items.push(...protocolMqttItems)
        // include generic provider script
        if (providerType === Provider.Generic) {
          const scriptItems = getGenericProviderScript(rootObject)
          items.push(...scriptItems)
        }
        break

      case Protocol.Serial:
        const protocolSerialItems = getProtocolSerialItems(rootObject)
        items.push(...protocolSerialItems)
        break

      case Protocol.Ethernet:
        const protocolEthernetItems = getProtocolEthernetItems(rootObject)
        items.push(...protocolEthernetItems)
        break

      case Protocol.HTTP:
        if (providerType === Provider.Generic) {
          // include generic provider script
          const scriptItems = getGenericProviderScript(rootObject)
          items.push(...scriptItems)
          // include http specific items
          const protocolHttpGenericItems = getHttpGenericProtocolItems(rootObject)
          items.push(...protocolHttpGenericItems)
        }
        break

      default:
        break
    }
  }

  if (
    providerType !== "" &&
    providerType !== Provider.SystemMonitoring &&
    providerType !== Provider.PhilipsHue &&
    providerType !== Provider.ESPHome &&
    providerType !== Provider.Generic
  ) {
    // message logger
    objectPath.set(rootObject, "messageLogger.type", MessageLogger.None, true)
    items.push(
      {
        label: "message_logger",
        fieldId: "!messageLogger.type",
        fieldType: FieldType.Divider,
      },
      {
        label: "type",
        fieldId: "messageLogger.type",
        isRequired: true,
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        options: MessageLoggerOptions,
        direction: "up",
      }
    )

    // message logger properties
    const msgLoggerType = objectPath.get(rootObject, "messageLogger.type", "").toLowerCase()
    switch (msgLoggerType) {
      case MessageLogger.FileLogger:
        const fileLoggerItems = getLoggerFileItems(rootObject)
        items.push(...fileLoggerItems)
        break

      default:
        break
    }
  }

  switch (providerType) {
    case Provider.SystemMonitoring:
      const systemMonitoringItems = getSystemMonitoringItems(rootObject)
      items.push(...systemMonitoringItems)
      break

    case Provider.PhilipsHue:
      const philipsHueItems = getPhilipsHueItems(rootObject)
      items.push(...philipsHueItems)
      break

    case Provider.ESPHome:
      const espHomeItems = getESPHomeItems(rootObject)
      items.push(...espHomeItems)
      break

    case Provider.Generic:
      // include node endpoints
      const genericHttpItems = getGenericNodeEndpointItems(rootObject)
      items.push(...genericHttpItems)
      break

    default:
    // NOOP
  }

  return items
}

// get provider items
// get MySensor Provider items
const getMySensorsItems = (rootObject) => {
  objectPath.set(rootObject, "provider.retryCount", 3, true)
  objectPath.set(rootObject, "provider.timeout", "500ms", true)
  const items = [
    {
      label: "internal_message_ack",
      fieldId: "provider.enableInternalMessageAck",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
    {
      label: "stream_message_ack",
      fieldId: "provider.enableStreamMessageAck",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
    {
      label: "retry_count",
      fieldId: "provider.retryCount",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      validator: { isInt: { min: 0 } },
      helperTextInvalid: "helper_text.invalid_retry_count",
    },
    {
      label: "timeout",
      fieldId: "provider.timeout",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
  ]
  return items
}

// get protocol items
// get protocol mqtt items
const getProtocolMqttItems = (rootObject, providerType) => {
  const items = [
    {
      label: "broker",
      fieldId: "provider.protocol.broker",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      validator: {
        isURL: {
          protocols: ["ws", "wss", "unix", "mqtt", "tcp", "ssl", "tls", "mqtts", "mqtt+ssl", "tcps"],
        },
      },
      helperTextInvalid: "helper_text.invalid_broker_url",
    },
    {
      label: "insecure",
      fieldId: "provider.protocol.insecure",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
    {
      label: "username",
      fieldId: "provider.protocol.username",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "password",
      fieldId: "provider.protocol.password",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "subscribe",
      fieldId: "provider.protocol.subscribe",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      helperTextInvalid: "helper_text.invalid_topic",
      validator: { isNotEmpty: {} },
    },
  ]

  if (providerType !== Provider.Generic) {
    objectPath.set(rootObject, "provider.protocol.qos", 0, true)
    items.push(
      {
        label: "publish",
        fieldId: "provider.protocol.publish",
        isRequired: true,
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        helperTextInvalid: "helper_text.invalid_topic",
        validator: { isNotEmpty: {} },
      },
      {
        label: "qos",
        fieldId: "provider.protocol.qos",
        isRequired: true,
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.Integer,
        value: "",
        options: [
          { value: "0", label: "At most once (0)" },
          { value: "1", label: "At least once (1)" },
          { value: "2", label: "Exactly once (2)" },
        ],
      }
    )
  }

  return items
}

// get protocol serial items
const getProtocolSerialItems = (rootObject) => {
  objectPath.set(rootObject, "provider.protocol.baudrate", 115200, true)
  const items = [
    {
      label: "port_name",
      fieldId: "provider.protocol.portname",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "baud_rate",
      fieldId: "provider.protocol.baudrate",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      validator: { isBaudRate: {} },
      helperTextInvalid: "helper_text.invalid_baud_rate",
    },
  ]
  return items
}

// get protocol ethernet items
const getProtocolEthernetItems = (_rootObject) => {
  const items = [
    {
      label: "server",
      fieldId: "provider.protocol.server",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      validator: {
        isURL: {
          protocols: ["tcp", "tcps", "ssl", "tls"],
        },
      },
      helperTextInvalid: "helper_text.invalid_server_url",
    },
    {
      label: "insecure",
      fieldId: "provider.protocol.insecure",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
  ]
  return items
}

// logger items
// get file logger items
const getLoggerFileItems = (rootObject) => {
  objectPath.set(rootObject, "messageLogger.flushInterval", "1s", true)
  objectPath.set(rootObject, "messageLogger.logRotateInterval", "12h", true)
  objectPath.set(rootObject, "messageLogger.maxSize", "2MiB", true)
  objectPath.set(rootObject, "messageLogger.maxAge", "72h", true)
  objectPath.set(rootObject, "messageLogger.maxBackup", 7, true)
  const items = [
    {
      label: "flush_interval",
      fieldId: "messageLogger.flushInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "log_rotate_interval",
      fieldId: "messageLogger.logRotateInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "maximum_size",
      fieldId: "messageLogger.maxSize",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "maximum_age",
      fieldId: "messageLogger.maxAge",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "maximum_backup",
      fieldId: "messageLogger.maxBackup",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      validator: { isInt: { min: 0 } },
      helperTextInvalid: "helper_text.invalid_maximum_backup",
    },
  ]
  return items
}
