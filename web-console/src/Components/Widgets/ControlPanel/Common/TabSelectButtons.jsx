import { SelectVariant } from "@patternfly/react-core"
import React from "react"
import { api } from "../../../../Service/Api"
import Select from "../../../Form/Select"
import ToggleButtonGroup from "../../../Form/ToggleButtonGroup/ToggleButtonGroup"

export const TabButton = ({
  id,
  quickId = "",
  mixedControlCfg = {},
  payload = "",
  keyPath = "",
  sendPayloadWrapper = () => {},
}) => {
  const { options = {} } = mixedControlCfg.config

  return (
    <ToggleButtonGroup
      key={id}
      selected={String(payload)}
      options={getOptions(options)}
      onSelectionFunc={(newValue) => sendPayloadWrapper(() => onChange(quickId, keyPath, newValue))}
    />
  )
}

export const SelectButton = ({
  id,
  quickId = "",
  mixedControlCfg = {},
  keyPath = "",
  payload = "",
  sendPayloadWrapper = () => {},
}) => {
  const { options = {}, dropDownPosition = "" } = mixedControlCfg.config

  return (
    <Select
      key={id}
      variant={SelectVariant.single}
      options={getOptions(options)}
      selected={String(payload)}
      onChange={(newValue) => sendPayloadWrapper(() => onChange(quickId, keyPath, newValue))}
      isDisabled={false}
      disableClear={true}
      direction={dropDownPosition}
    />
  )
}

const getOptions = (options) => {
  const tabOptions = Object.keys(options).map((key) => {
    return {
      label: options[key],
      value: key,
    }
  })
  return tabOptions
}

const onChange = (quickId, keyPath, payload) => {
  api.action.send({ resource: quickId, keyPath: keyPath, payload: payload })
}
