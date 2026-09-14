import { TextInput } from "@patternfly/react-core"
import React from "react"
import ChartYAxisConfigMap from "./ChartYAxisConfigMap"

// displayYAxisConfig displays Y Axis config details in readable way
export const displayYAxisConfig = (index, value = {}, _onChange, validated, _isDisabled) => {
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={`offsetY=${value.offsetY}, color=${value.color}, unit=${value.unit}`}
      isDisabled={true}
      validated={validated}
    />
  )
}

// calls the model to update the axis config details
export const chartYAxisConfigUpdateButtonCallback = (index = 0, item = {}, onChange) => {
  return (
    <ChartYAxisConfigMap
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
