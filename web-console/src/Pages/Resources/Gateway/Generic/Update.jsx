import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import { validate } from "../../../../Util/Validator"
import { callBackEndpointConfigUpdateButtonCallback, getEndpointConfigDisplayValue } from "./Endpoints"

// get Generic http config items
export const getGenericNodeEndpointItems = (rootObject) => {
  objectPath.set(rootObject, "provider.protocol.nodes", {}, true)
  const protocolType = objectPath.get(rootObject, "provider.protocol.type", "").toLowerCase()

  const items = []
  if (protocolType === "") {
    return items
  }

  items.push(
    {
      label: "node_endpoints",
      fieldId: "!node_endpoints",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "provider.protocol.nodes",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
      keyLabel: "node_id",
      valueLabel: "endpoint",
      showUpdateButton: true,
      saveButtonText: "update",
      updateButtonText: "update",
      minimapEnabled: true,
      isRequired: false,
      validateKeyFunc: (key) => validate("isID", key),
      validateValueFunc: (value) => value.url !== undefined || value.url !== "",
      valueField: getEndpointConfigDisplayValue,
      updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
        callBackEndpointConfigUpdateButtonCallback(
          cbIndex,
          cbItem,
          cbOnChange,
          true,
          protocolType,
          false,
          "update_endpoint"
        ),
    }
  )

  return items
}

// get script, generic provider
// final script of receive and send
export const getGenericProviderScript = (rootObject) => {
  objectPath.set(rootObject, "provider.script.onReceive", "", true)
  objectPath.set(rootObject, "provider.script.onSend", "", true)

  const items = [
    {
      label: "script",
      fieldId: "!script",
      fieldType: FieldType.Divider,
    },
    {
      label: "on_receive",
      fieldId: "provider.script.onReceive",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.String,
      value: "",
      saveButtonText: "update",
      updateButtonText: "update_script",
      language: "javascript",
      minimapEnabled: true,
      isRequired: false,
    },
    {
      label: "on_send",
      fieldId: "provider.script.onSend",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.String,
      value: "",
      saveButtonText: "update",
      updateButtonText: "update_script",
      language: "javascript",
      minimapEnabled: true,
      isRequired: false,
    },
  ]

  return items
}
