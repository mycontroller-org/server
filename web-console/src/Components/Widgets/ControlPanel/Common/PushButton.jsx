import React from "react"
import { Button } from "@patternfly/react-core"
import { api } from "../../../../Service/Api"

const PushButton = ({
  id,
  widgetId,
  quickId,
  payload = "",
  keyPath = "",
  buttonText = "NO TEXT",
  buttonType = "secondary",
  minWidth = 70,
  sendPayloadWrapper = () => {},
}) => {
  return (
    <Button
      id={`${widgetId}_${id}`}
      onClick={() => sendPayloadWrapper(() => onChange(quickId, keyPath, payload))}
      variant={buttonType}
      isSmall
      style={{ minWidth: `${minWidth}px` }}
    >
      <span style={{ fontWeight: 500 }}>{buttonText}</span>
    </Button>
  )
}

const onChange = (quickId, keyPath, payload) => {
  api.action.send({ resource: quickId, payload: payload, keyPath: keyPath })
}

export default PushButton
