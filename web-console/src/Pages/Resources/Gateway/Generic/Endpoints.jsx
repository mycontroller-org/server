import { TextInput } from "@patternfly/react-core"
import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import { Protocol } from "../../../../Constants/Gateway"
import { getValue } from "../../../../Util/Util"
import { validate } from "../../../../Util/Validator"
import EndpointConfigPicker from "./EndpointPicker"

// get http generic protocol items
export const getHttpGenericProtocolItems = (rootObject) => {
  objectPath.set(rootObject, "provider.protocol.headers", {}, true)
  objectPath.set(rootObject, "provider.protocol.queryParameters", {}, true)
  objectPath.set(rootObject, "provider.protocol.endpoints", {}, true)

  const protocolType = getValue(objectPath, "provider.protocol.type", Protocol.HTTP)

  const items = [
    {
      label: "global_configuration",
      fieldId: "!global_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "headers",
      fieldId: "provider.protocol.headers",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.ArrayObject,
      value: {},
      keyLabel: "header_name",
      valueLabel: "value",
      validateKeyFunc: (key) => validate("isNotEmpty", key, {}),
      validateValueFunc: (value) => validate("isNotEmpty", value, {}),
    },
    {
      label: "query_parameters",
      fieldId: "provider.protocol.queryParameters",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.Object,
      language: "yaml",
      updateButtonText: "update_query_parameters",
      value: {},
      isRequired: false,
    },
    {
      label: "endpoints",
      fieldId: "!endpoints",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "provider.protocol.endpoints",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
      keyLabel: "name",
      valueLabel: "endpoint",
      showUpdateButton: true,
      saveButtonText: "update",
      updateButtonText: "update",
      minimapEnabled: true,
      isRequired: false,
      validateKeyFunc: (key) => validate("isVariableKey", key),
      validateValueFunc: (value) => value.url !== undefined || value.url !== "",
      valueField: getEndpointConfigDisplayValue,
      updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
        callBackEndpointConfigUpdateButtonCallback(
          cbIndex,
          cbItem,
          cbOnChange,
          false,
          protocolType,
          false,
          "update_endpoint"
        ),
    },
  ]
  return items
}

// display endpoint config button

// returns variable value to display value
export const getEndpointConfigDisplayValue = (index, value, _onChange, validated, _isDisabled) => {
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={`${value.url ? value.url : value.topic}`}
      isDisabled={true}
      validated={validated}
    />
  )
}

// calls the model to update the endpoint config details
export const callBackEndpointConfigUpdateButtonCallback = (
  index = 0,
  item = {},
  onChange,
  isNodeEndpoint = false,
  protocolType = Protocol.HTTP,
  isPreRun = false,
  title = "update_endpoint"
) => {
  return (
    <EndpointConfigPicker
      key={"picker_" + index}
      value={item.value}
      name={item.key}
      id={"model_" + index}
      onChange={(newValue) => {
        onChange(newValue)
      }}
      isNodeEndpoint={isNodeEndpoint}
      isPreRun={isPreRun}
      protocolType={protocolType}
      title={title}
    />
  )
}
