import React from "react"
import { ControlType } from "../../../Constants/Widgets/ControlPanel"
import { getValue } from "../../../Util/Util"
import MixedControl from "./MixedControl/MixedControl"
import SwitchPanel from "./SwitchPannel/SwitchPanel"

const ControlPanel = ({ widgetId, config, dimensions, history }) => {
  const controlType = getValue(config, "type", "")

  switch (controlType) {
    case ControlType.SwitchToggle:
    case ControlType.SwitchButton:
      return <SwitchPanel widgetId={widgetId} config={config} dimensions={dimensions} history={history} />

    case ControlType.MixedControl:
      return <MixedControl widgetId={widgetId} config={config} dimensions={dimensions} history={history} />

    default:
      return <span>unknown control type:{controlType}</span>
  }
}

export default ControlPanel
