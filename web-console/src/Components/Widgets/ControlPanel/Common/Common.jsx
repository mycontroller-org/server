import React from "react"
import { ControlType, MixedControlType } from "../../../../Constants/Widgets/ControlPanel"
import { LastSeen } from "../../../Time/Time"
import SwitchButton from "./SwitchButton"
import SwitchToggle from "./SwitchToggle"
import PushButton from "./PushButton"
import { SelectButton, TabButton } from "./TabSelectButtons"
import { Table, TableBody, TableHeader, TableVariant } from "@patternfly/react-table"
import "./Common.scss"
import InputField from "./InputField"
import { navigateToResource } from "../../Helper/Resource"
import SliderControl from "./SliderControl"
import { Button, Divider, Split, SplitItem, Stack, StackItem } from "@patternfly/react-core"

const columns = [{ title: "Name" }, "Last Update", ""]

const ControlObjects = ({
  widgetId = "",
  resources = [],
  config = {},
  history = {},
  sendPayloadWrapper = () => {},
}) => {
  const rows = resources.map((resource, index) => {
    const mixedControlCfg = resource.config && resource.config.control ? resource.config.control : {}

    let controlElement = null
    const { type: controlType = "" } = config

    let askConfirmation = false
    let confirmationMessage = ""

    let ctlObjectType = ""

    switch (controlType) {
      case ControlType.SwitchToggle:
        ctlObjectType = MixedControlType.ToggleSwitch
        break

      case ControlType.SwitchButton:
        ctlObjectType = MixedControlType.ButtonSwitch
        break

      case ControlType.MixedControl:
        ctlObjectType = mixedControlCfg.type
        askConfirmation = mixedControlCfg.askConfirmation
        confirmationMessage = mixedControlCfg.confirmationMessage
        break

      default:
    }

    switch (ctlObjectType) {
      case MixedControlType.ToggleSwitch:
        controlElement = (
          <SwitchToggle
            id={`toggle_switch_${resource.id}_${index}`}
            widgetId={widgetId}
            quickId={resource.quickId}
            payload={resource.payload}
            keyPath={resource.keyPath}
            payloadOn="true"
            payloadOff="false"
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.ButtonSwitch:
        let btnConfig = config
        if (mixedControlCfg.config) {
          btnConfig = mixedControlCfg.config.button
        }
        const { onText, offText, minWidth, onButtonType } = btnConfig
        controlElement = (
          <SwitchButton
            id={`button_switch_${resource.id}_${index}`}
            widgetId={widgetId}
            quickId={resource.quickId}
            payload={resource.payload}
            keyPath={resource.keyPath}
            payloadOn="true"
            payloadOff="false"
            onButtonType={onButtonType}
            onText={onText}
            offText={offText}
            minWidth={minWidth}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.PushButton:
        const { button = {}, payload = "" } = mixedControlCfg.config
        controlElement = (
          <PushButton
            id={`push_button_${resource.id}_${index}`}
            widgetId={widgetId}
            quickId={resource.quickId}
            payload={payload}
            keyPath={resource.keyPath}
            buttonType={button.buttonType}
            buttonText={button.text}
            minWidth={button.minWidth}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.SelectOptions:
        controlElement = (
          <SelectButton
            id={`tab_options_${resource.id}_${index}`}
            mixedControlCfg={mixedControlCfg}
            payload={resource.payload}
            quickId={resource.quickId}
            keyPath={resource.keyPath}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.TabOptions:
        controlElement = (
          <TabButton
            id={`tab_options_${resource.id}_${index}`}
            mixedControlCfg={mixedControlCfg}
            payload={resource.payload}
            quickId={resource.quickId}
            keyPath={resource.keyPath}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.Input:
        controlElement = (
          <InputField
            id={`input_filed_${resource.id}_${index}`}
            mixedControlCfg={mixedControlCfg}
            payload={resource.payload}
            quickId={resource.quickId}
            keyPath={resource.keyPath}
            minWidth={mixedControlCfg.config.input.minWidth}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      case MixedControlType.Slider:
        const { slider = {} } = mixedControlCfg.config
        controlElement = (
          <SliderControl
            id={`push_button_${resource.id}_${index}`}
            widgetId={widgetId}
            quickId={resource.quickId}
            payload={resource.payload}
            keyPath={resource.keyPath}
            min={slider.min}
            max={slider.max}
            step={slider.step}
            minWidth={slider.minWidth}
            sendPayloadWrapper={(callBack) =>
              sendPayloadWrapper(askConfirmation, confirmationMessage, callBack)
            }
          />
        )
        break

      default:
        controlElement = <span>unknown type:{ctlObjectType}</span>
    }

    return [
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={() => navigateToResource(resource.type, resource.id, history)}
          >
            {resource.label}
          </Button>
        ),
      },
      {
        title: (
          <span className="value-timestamp">
            <LastSeen date={resource.timestamp} tooltipPosition="top" />
          </span>
        ),
      },
      { title: <span style={{ float: "right" }}>{controlElement}</span> },
    ]
  })

  const { tableView, hideHeader, hideBorder } = config

  if (tableView) {
    return (
      <Table
        key={"mixed_panel_table_" + widgetId}
        className="mc-control-panel"
        aria-label="Mixed Panel Table"
        variant={TableVariant.compact}
        borders={!hideBorder}
        cells={columns}
        rows={rows}
      >
        <TableHeader hidden={hideHeader} />
        <TableBody />
      </Table>
    )
  }

  const divider = hideBorder ? null : <Divider style={{ margin: "4px 0px" }} />

  const switches = rows.map((row, index) => {
    const dividerComponent = resources.length > 1 ? divider : null
    return (
      <StackItem key={`si_${index}`} style={{ marginBottom: "6px" }}>
        <Split hasGutter>
          <SplitItem>{row[0].title}</SplitItem>
          <SplitItem isFilled>{row[1].title}</SplitItem>
          <SplitItem>
            <span style={{ float: "right" }}>{row[2].title}</span>
          </SplitItem>
        </Split>
        {dividerComponent}
      </StackItem>
    )
  })
  return (
    <Stack key={`cp_${widgetId}`} className="mc-control-panel">
      {switches}
    </Stack>
  )
}

export default ControlObjects
