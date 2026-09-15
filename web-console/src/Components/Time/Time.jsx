import { Tooltip } from "@patternfly/react-core"
import moment from "moment"
import React from "react"

const FROM_NOW_INTERVAL_MS = 30000
const tickListeners = new Set()
let tickTimer = null

const subscribeFromNowTick = (fn) => {
  tickListeners.add(fn)
  if (tickTimer == null) {
    tickTimer = setInterval(() => {
      tickListeners.forEach((listener) => listener())
    }, FROM_NOW_INTERVAL_MS)
  }
  return () => {
    tickListeners.delete(fn)
    if (tickListeners.size === 0 && tickTimer != null) {
      clearInterval(tickTimer)
      tickTimer = null
    }
  }
}

export const LastSeen = ({ date = "", tooltipPosition = "left" }) => {
  const [, setTick] = React.useState(0)
  const hasDate = date !== "" && date !== null
  React.useEffect(() => {
    if (!hasDate) {
      return undefined
    }
    return subscribeFromNowTick(() => setTick((n) => n + 1))
  }, [hasDate])

  if (!hasDate) {
    return <span></span>
  }
  const lastSeen = moment(date)
  const disabled = lastSeen.year() <= 1 // set "-" of zero year
  const value = disabled ? <span>-</span> : <span>{lastSeen.fromNow()}</span>
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
