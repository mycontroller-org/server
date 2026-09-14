import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { DataType, FieldType } from "../../../../Constants/Form"
import { ColorsSetBig } from "../../../../Constants/Widgets/Color"
import Editor from "../../../Editor/Editor"
import ErrorBoundary from "../../../ErrorBoundary/ErrorBoundary"
import { withTranslation } from "react-i18next"

class ChartYAxisConfigMap extends React.Component {
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
          title={`${t("update_y_axis")}: ${name}`}
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

ChartYAxisConfigMap.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
}

export default withTranslation()(ChartYAxisConfigMap)

const getItems = (rootObject) => {
  // set default values
  objectPath.set(rootObject, "offsetY", 0, true)
  objectPath.set(rootObject, "color", "#4f5255", true)
  objectPath.set(rootObject, "roundDecimal", 0, true)
  objectPath.set(rootObject, "unit", "", true)

  const items = [
    {
      label: "offset_y_%",
      fieldId: "offsetY",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 0,
      max: 100,
      step: 1,
    },
    {
      label: "color",
      fieldId: "color",
      fieldType: FieldType.ColorBox,
      dataType: DataType.String,
      colors: ColorsSetBig,
      isRequired: true,
      value: "",
      validator: { isNotEmpty: {} },
    },
    {
      label: "round_decimal",
      fieldId: "roundDecimal",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 0,
      max: 10,
      step: 1,
    },
    {
      label: "unit",
      fieldId: "unit",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      isRequired: false,
      value: "",
    },
  ]

  return items
}
