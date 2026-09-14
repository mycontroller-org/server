import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import Editor from "../../../../Components/Editor/Editor"
import ErrorBoundary from "../../../../Components/ErrorBoundary/ErrorBoundary"
import { DataType, FieldType } from "../../../../Constants/Form"
import { DataUnitType, DataUnitTypeOptions } from "../../../../Constants/Metric"

class ProcessConfigPicker extends React.Component {
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
    const { value, id, name, onChange } = this.props
    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={"Update process: " + name}
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
              getFormItems={(rootObject) => getItems(rootObject, name)}
              saveButtonText="update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

ProcessConfigPicker.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
}

export default ProcessConfigPicker

const getItems = (rootObject, sourceId) => {
  objectPath.set(rootObject, "disabled", false, true)
  objectPath.set(rootObject, "name", sourceId, true)
  objectPath.set(rootObject, "unit", DataUnitType.MiB, true)
  objectPath.set(rootObject, "filter", {}, true)

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
      label: "name",
      fieldId: "name",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "unit",
      fieldId: "unit",
      fieldType: FieldType.Select,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      options: DataUnitTypeOptions,
    },
    {
      label: "filters",
      fieldId: "!filters_process",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "filter",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: "",
      isRequired: true,
      isDisabled: false,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_filters",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
  ]
  return items
}
