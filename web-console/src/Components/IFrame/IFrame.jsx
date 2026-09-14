import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { ExternalLinkSquareAltIcon } from "@patternfly/react-icons"
import React from "react"
import "./IFrame.scss"

class IFrame extends React.Component {
  state = {
    isOpen: false,
  }

  handleToggle = () => {
    this.setState(({ isOpen }) => ({
      isOpen: !isOpen,
    }))
  }

  render() {
    const { isOpen } = this.state
    const { enableInLine = false } = this.props

    let elements = null
    if (enableInLine) {
      elements = (
        <>
          <Button variant="link" isInline onClick={this.handleToggle}>
            {this.props.url}
          </Button>{" "}
          <a href={this.props.url} target="_blank" rel="noopener noreferrer">
            <ExternalLinkSquareAltIcon />
          </a>
        </>
      )
    } else {
      elements = (
        <a href={this.props.url} target="_blank" rel="noopener noreferrer">
          {this.props.url} <ExternalLinkSquareAltIcon />
        </a>
      )
    }

    return (
      <>
        {elements}
        <Modal
          title={this.props.url}
          isOpen={isOpen}
          onClose={this.handleToggle}
          variant={ModalVariant.default}
          className="mc-iframe"
        >
          <div
            className="iframe"
            dangerouslySetInnerHTML={{
              __html: '<iframe width="100%" align-items="stretch" src="' + this.props.url + '"></iframe>',
            }}
          />
        </Modal>
      </>
    )
  }
}

export default IFrame
