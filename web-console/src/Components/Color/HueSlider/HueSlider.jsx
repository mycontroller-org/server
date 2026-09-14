import { Badge, Split, SplitItem } from "@patternfly/react-core"
import React, { useEffect, useState } from "react"
import { HuePicker } from "react-color"
import "./HueSlider.scss"

const HueSlider = ({ hue = 230, onChange = () => {}, className = "" }) => {
  const [hueColor, setHueColor] = useState(0)

  useEffect(() => {
    setHueColor(hue)
  }, [])

  return (
    <Split className="mc-hue-slider">
      <SplitItem isFilled className={`slider ${className}`}>
        <HuePicker
          className="slider"
          key="hue-picker"
          color={{ h: hueColor, s: 0, l: 0.1 }}
          onChange={(newColor) => {
            setHueColor(newColor.hsl.h.toFixed(0))
          }}
          onChangeComplete={(newColor) => {
            onChange(newColor.hsl.h.toFixed(0))
          }}
        />
      </SplitItem>
      <SplitItem className="badge">
        <Badge key={1}>{hue}</Badge>
      </SplitItem>
    </Split>
  )
}

export default HueSlider
