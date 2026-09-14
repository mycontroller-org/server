import React from "react"
import { api } from "../../../../Service/Api"
import SimpleSlider from "../../../Form/Slider/Simple"

const SliderControl = ({
  id,
  widgetId,
  quickId,
  payload = "",
  keyPath = "",
  min = 0,
  max = 100,
  step = 1,
  minWidth = 70,
  sendPayloadWrapper = () => {},
}) => {
  return (
    <div style={{ minWidth: `${minWidth}px`, marginBottom: "-6px" }}>
      <SimpleSlider
        id={`${widgetId}_${id}`}
        min={min}
        max={max}
        step={step}
        value={payload}
        onChange={(newValue) => sendPayloadWrapper(() => onChange(quickId, keyPath, newValue))}
      />
    </div>
  )
}

const onChange = (quickId, keyPath, payload) => {
  api.action.send({ resource: quickId, payload: payload, keyPath: keyPath })
}

export default SliderControl
