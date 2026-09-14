import { Popover } from "@patternfly/react-core"
import React from "react"
import { CirclePicker } from "react-color"
import "./ColorBox.scss"

const defaultColors = [
  "#06c", // --pf-global--primary-color--100
  "#3e8635", // --pf-global--success-color--100
  "#2b9af3", // --pf-global--info-color--100
  "#f0ab00", // --pf-global--warning-color--100
  "#c9190b", // --pf-global--danger-color--100
  "#151515", // --pf-global--Color--dark-100
  "#f44336", // default color of circle picker
  "#e91e63", // default color of circle picker
  "#9c27b0", // default color of circle picker
  "#673ab7", // default color of circle picker
  "#3f51b5", // default color of circle picker
  "#2196f3", // default color of circle picker
  "#03a9f4", // default color of circle picker
  "#00bcd4", // default color of circle picker
  "#009688", // default color of circle picker
  "#4caf50", // default color of circle picker
  "#8bc34a", // default color of circle picker
  "#cddc39", // default color of circle picker
  "#ffeb3b", // default color of circle picker
  "#ffc107", // default color of circle picker
  "#ff9800", // default color of circle picker
  "#ff5722", // default color of circle picker
  "#795548", // default color of circle picker
  "#607d8b", // default color of circle picker
]
const defaultColor = "#0066CC"

const ColorBox = ({ colors = defaultColors, color = defaultColor, onChange = () => {}, style = {} }) => {
  const selectedColor = color === "" ? defaultColor : color

  return (
    <Popover
      // aria-label="Popover with no header, footer, close button, and padding"
      hasNoPadding={false}
      hasAutoWidth
      showClose={false}
      bodyContent={
        <CirclePicker
          className="mc-color-circle-picker"
          colors={colors}
          color={selectedColor}
          circleSize={21}
          circleSpacing={4}
          onChangeComplete={(newColor, _event) => {
            onChange(newColor.hex)
          }}
        />
      }
    >
      <div className="mc-color-box" style={{ backgroundColor: selectedColor, ...style }}></div>
    </Popover>
  )
}

export default ColorBox
