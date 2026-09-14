import { TextInput } from "@patternfly/react-core"
import { DataType, FieldType } from "../../../../Constants/Form"
import { validate } from "../../../../Util/Validator"
import NodeConfigPicker from "./NodeConfigPicker"

// get System Monitoring config
export const getSystemMonitoringItems = (_rootObject) => {
  const items = [
    {
      label: "configuration",
      fieldId: "!configuration_sm",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "provider.hostConfigMap",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
      keyLabel: "node_id",
      valueLabel: "configuration",
      showUpdateButton: true,
      saveButtonText: "update",
      updateButtonText: "update",
      minimapEnabled: true,
      isRequired: false,
      validateKeyFunc: (key) => validate("isID", key),
      validateValueFunc: (value) => value.disabled !== undefined,
      valueField: getNodeConfigDisplayValue,
      updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
        nodeConfigUpdateButtonCallback(cbIndex, cbItem, cbOnChange),
    },
  ]
  return items
}

// display node config button

// returns variable value to display on the nodes list
export const getNodeConfigDisplayValue = (index, value, _onChange, validated, _isDisabled) => {
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={`disabled:${value.disabled}`}
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
