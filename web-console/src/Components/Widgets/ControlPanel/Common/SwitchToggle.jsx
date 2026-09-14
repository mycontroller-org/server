import React from "react"
import { Switch as PfSwitch } from "@patternfly/react-core"
import { api } from "../../../../Service/Api"

const SwitchToggle = ({
  id,
  widgetId,
  quickId,
  payload,
  payloadOn = "true",
  payloadOff = "false",
  keyPath = "",
  sendPayloadWrapper = () => {},
}) => {
  const isChecked = String(payload) === payloadOn
  return (
    <PfSwitch
      id={`${widgetId}_${id}`}
      aria-label={`${widgetId}_${id}`}
      onChange={(newState) =>
        sendPayloadWrapper(() => onChange(newState, quickId, keyPath, payloadOn, payloadOff))
      }
      isChecked={isChecked}
    />
  )
}

const onChange = (isChecked, quickId, keyPath, payloadOn, payloadOff) => {
  api.action.send({ resource: quickId, keyPath, payload: isChecked ? payloadOn : payloadOff })
}

export default SwitchToggle
