import React from "react"
import { TextInput } from "@patternfly/react-core"
import NodeConfigPicker from "./NodeConfigPicker"

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
        onChange(index, "value", newValue)
      }}
    />
  )
}
