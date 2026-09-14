import { Tooltip } from "@patternfly/react-core"
import moment from "moment"
import React from "react"
import Moment from "react-moment"

// pooled timer enabled on App.js
// read more on: https://github.com/headzoo/react-moment#pooled-timer
export const LastSeen = ({ date = "", tooltipPosition = "left" }) => {
  if (date === "" || date === null) {
    return <span></span>
  }
  const lastSeen = moment(date)
  const disabled = lastSeen.year() <= 1 // set "-" of zero year
  const value = disabled ? <span>-</span> : <Moment date={date} fromNow />
  return (
    <Tooltip
      position={tooltipPosition}
      content={lastSeen.format("DD-MMM-YYYY, hh:mm:ss A")}
      disabled={disabled}
    >
      {value}
    </Tooltip>
  )
}

export const getLastSeen = (date) => {
  const lastSeen = moment(date)
  const disabled = lastSeen.year() <= 1 // set "-" of zero year
  return disabled ? "-" : lastSeen.fromNow()
}
