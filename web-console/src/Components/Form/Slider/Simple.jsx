import { Badge, Split, SplitItem } from "@patternfly/react-core"
import Slider, { createSliderWithTooltip } from "rc-slider"
import React from "react"
import "./Simple.scss"
const SliderWithTooltip = createSliderWithTooltip(Slider)

const SimpleSlider = ({
  className = "",
  id = "",
  min = 0,
  max = 100,
  value = 0,
  step = 1,
  isDisabled = false,
  onChange = () => {},
}) => {
  return (
    <Split className="mc-slider">
      <SplitItem isFilled className={`slider ${className}`}>
        <SliderWithTooltip
          key={id}
          min={min}
          max={max}
          step={step}
          defaultValue={value}
          onAfterChange={(val) => onChange(String(val))}
          disabled={isDisabled}
          handleStyle={{
            backgroundColor: "#06c",
            borderColor: "#06c",
            color: "#06c",
            width: "1rem",
            height: "1rem",
            marginTop: "-6px",
          }}
          trackStyle={[{ backgroundColor: "#2b9af3" }]}
          railStyle={{ backgroundColor: "#d2d2d2" }}
        ></SliderWithTooltip>
      </SplitItem>
      <SplitItem className="badge">
        <Badge key={1}>{value}</Badge>
      </SplitItem>
    </Split>
  )
}

export default SimpleSlider
