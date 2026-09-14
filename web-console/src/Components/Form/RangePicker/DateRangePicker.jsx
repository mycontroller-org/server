import { DatePicker, Split, SplitItem } from "@patternfly/react-core"
import React from "react"

const DateRangePicker = ({ value = {}, onChange = () => {} }) => {
  return (
    <Split>
      <SplitItem>
        <DatePicker
          onChange={(_event, from) => {
            onChange({ ...value, from })
          }}
          value={value.from}
          placeholder="yyyy-mm-dd"
        />
      </SplitItem>
      <SplitItem style={{ padding: "3px 12px 0 12px" }}>to</SplitItem>
      <SplitItem>
        <DatePicker
          onChange={(_event, to) => {
            onChange({ ...value, to })
          }}
          value={value.to}
          placeholder="yyyy-mm-dd"
        />
      </SplitItem>
    </Split>
  )
}

export default DateRangePicker
