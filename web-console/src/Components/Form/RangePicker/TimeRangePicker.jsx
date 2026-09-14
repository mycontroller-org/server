import { Split, SplitItem, TimePicker } from "@patternfly/react-core"
import React from "react"
import "./TimeRangePicker.scss"

const TimeRangePicker = ({ value = {}, onChange = () => {}, is24Hour = true }) => {
  return (
    <Split>
      <SplitItem>
        <TimePicker
          className="mc-time-range-picker"
          onChange={(_event, from) => {
            onChange({ ...value, from })
          }}
          defaultTime={value.from}
          is24Hour={is24Hour}
          placeholder="hh:mm:ss"
        />
      </SplitItem>
      <SplitItem style={{ padding: "6px 12px 0 12px" }}>to</SplitItem>
      <SplitItem>
        <TimePicker
          className="mc-time-range-picker"
          onChange={(_event, to) => {
            onChange({ ...value, to })
          }}
          defaultTime={value.to}
          is24Hour={is24Hour}
          placeholder="hh:mm:ss"
        />
      </SplitItem>
    </Split>
  )
}

export default TimeRangePicker
