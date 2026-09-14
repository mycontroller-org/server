import { Spinner as PfSpinner } from "@patternfly/react-core"
import React from "react"
import "./Spinner.scss"

export const HeaderSpinner = () => {
  return <div className="mc-spinner"></div>
}

export const Spinner = () => {
  return <PfSpinner size="lg" aria-valuetext="Loading..." />
}
