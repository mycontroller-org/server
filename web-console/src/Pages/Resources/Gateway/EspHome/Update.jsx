import { TextInput } from "@patternfly/react-core"
import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import { validate } from "../../../../Util/Validator"
import NodeConfigPicker from "./NodeConfigPicker"

// get ESPHome config
export const getESPHomeItems = (rootObject) => {
  objectPath.set(rootObject, "provider.timeout", "5s", true)
  objectPath.set(rootObject, "provider.aliveCheckInterval", "15s", true)
  const items = [
    {
      label: "global_configuration",
      fieldId: "!configuration_sm",
      fieldType: FieldType.Divider,
    },
    {
      label: "connection_timeout",
      fieldId: "provider.timeout",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "alive_check_interval",
      fieldId: "provider.aliveCheckInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "password",
      fieldId: "provider.password",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "encryption_key",
      fieldId: "provider.encryptionKey",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    // {
    //   label: "Auto Discover",
    //   fieldId: "provider.autoDiscover",
    //   fieldType: FieldType.Switch,
    //   dataType: DataType.Boolean,
    //   value: false,
    //   isDisabled: true,
    // },
    {
      label: "nodes",
      fieldId: "!nodes_list",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "provider.nodes",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
      keyLabel: "name",
      valueLabel: "configuration",
      showUpdateButton: true,
      saveButtonText: "update",
      updateButtonText: "update",
      minimapEnabled: true,
      isRequired: false,
      validateKeyFunc: (key) => validate("isID", key),
      validateValueFunc: (value) => value.address !== undefined,
      valueField: getNodeConfigDisplayValue,
      updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
        nodeConfigUpdateButtonCallback(cbIndex, cbItem, cbOnChange),
    },
  ]
  return items
}

// display node config

// returns variable value to display on the nodes list
export const getNodeConfigDisplayValue = (index, value, _onChange, validated, _isDisabled) => {
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={`${value.address}, disabled:${value.disabled}`}
      isDisabled={true}
      validated={validated}
    />
  )
}

// calls the model to update the node config details
export const nodeConfigUpdateButtonCallback = (index = 0, item = {}, onChange) => {
  return (
    <NodeConfigPicker
      key={"picker_" + index}
      value={item.value}
      name={item.key}
      id={"model_" + index}
      onChange={(newValue) => {
        onChange(newValue)
      }}
    />
  )
}
