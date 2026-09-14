import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { DataType, FieldType } from "../../../../Constants/Form"
import Editor from "../../../Editor/Editor"
import ErrorBoundary from "../../../ErrorBoundary/ErrorBoundary"

class NodeConfigPicker extends React.Component {
  state = {
    isOpen: false,
  }

  onClose = () => {
    this.setState({ isOpen: false })
  }

  onOpen = () => {
    this.setState({ isOpen: true })
  }

  render() {
    const { isOpen } = this.state
    const { value, id, name, onChange, t } = this.props
    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={`${t("update_node")}: ${name}`}
          variant={ModalVariant.medium}
          position="top"
          isOpen={isOpen}
          onClose={this.onClose}
          onEscapePress={this.onClose}
        >
          <ErrorBoundary>
            <Editor
              key={"editor" + id}
              disableEditor={false}
              language="yaml"
              rootObject={value}
              onSaveFunc={(rootObject) => {
                onChange(rootObject)
                this.onClose()
              }}
              onChangeFunc={() => {}}
              minimapEnabled={false}
              onCancelFunc={this.onClose}
              isWidthLimited={false}
              getFormItems={getItems}
              saveButtonText="update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

NodeConfigPicker.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
}

export default withTranslation()(NodeConfigPicker)

const getItems = (rootObject) => {
  objectPath.set(rootObject, "disabled", false, true)
  objectPath.set(rootObject, "address", "", true)
  objectPath.set(rootObject, "password", "", true)
  objectPath.set(rootObject, "timeout", "5s", true)
  objectPath.set(rootObject, "aliveCheckInterval", "15s", true)
  objectPath.set(rootObject, "reconnectDelay", "", true)

  const items = [
    {
      label: "disabled",
      fieldId: "disabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
      isRequired: false,
    },
    {
      label: "address",
      fieldId: "address",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: false,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_address_esphome",
      validated: "default",
      validator: {
        isURL: {
          require_protocol: false,
          require_host: true,
          require_port: true,
          allow_underscores: true,
        },
      },
    },
    {
      label: "password",
      fieldId: "password",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "connection_timeout",
      fieldId: "timeout",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "alive_check_interval",
      fieldId: "aliveCheckInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "reconnect_delay",
      fieldId: "reconnectDelay",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
  ]
  return items
}
