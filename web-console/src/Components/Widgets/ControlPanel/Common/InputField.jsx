import React from "react"
import { Button, Popover, Split, SplitItem, TextInput } from "@patternfly/react-core"
import { api } from "../../../../Service/Api"
import "./Common.scss"

class SendPayload extends React.Component {
  state = {
    value: "",
  }

  componentDidMount() {
    this.setState({ value: this.props.payload })
  }

  send = (value) => {
    const { quickId, keyPath } = this.props
    api.action.send({ resource: quickId, payload: value, keyPath: keyPath })
  }

  onChange = (newValue) => {
    this.setState({ value: newValue })
  }

  render() {
    const { value } = this.state
    const { sendPayloadWrapper } = this.props

    return (
      <Split hasGutter>
        <SplitItem isFilled>
          <TextInput value={value} onChange={this.onChange} />
        </SplitItem>
        <SplitItem>
          <Button variant="primary" onClick={() => sendPayloadWrapper(() => this.send(value))}>
            Send
          </Button>
        </SplitItem>
        <SplitItem>
          <Button variant="tertiary" onClick={() => this.onChange("")}>
            Clear
          </Button>
        </SplitItem>
      </Split>
    )
  }
}

const InputField = ({
  id,
  widgetId,
  quickId,
  keyPath = "",
  payload = "",
  minWidth = 70,
  sendPayloadWrapper = () => {},
}) => {
  return (
    <Popover
      zIndex={399}
      aria-label="Input popover"
      position="auto"
      hasAutoWidth
      headerContent={<div>Send Payload</div>}
      bodyContent={
        <SendPayload
          quickId={quickId}
          keyPath={keyPath}
          payload={payload}
          sendPayloadWrapper={sendPayloadWrapper}
        />
      }
      className="mixed-ctl-input-field"
    >
      <div style={{ minWidth: `${minWidth}px`, width: `${minWidth}px` }}>
        <TextInput
          id={`${widgetId}_${id}`}
          key={`${widgetId}_${id}`}
          onClick={() => {}} // acts like readonly
          isDisabled={false}
          value={payload}
        />
      </div>
    </Popover>
  )
}

export default InputField
