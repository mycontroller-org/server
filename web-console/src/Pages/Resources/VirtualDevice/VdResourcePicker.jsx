import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { DataType, FieldType } from "../../../Constants/Form"
import { ResourceTypeOptions } from "../../../Constants/ResourcePicker"
import Editor from "../../../Components/Editor/Editor"
import ErrorBoundary from "../../../Components/ErrorBoundary/ErrorBoundary"
import {
  getOptionsDescriptionFunc,
  getResourceFilterFunc,
  getResourceOptionsAPI,
  getResourceOptionValueFunc,
} from "../../../Components/Form/ResourcePicker/ResourceUtils"
import { DeviceTraitOptions } from "../../../Constants/VirtualDevice"

class VdResourcePicker extends React.Component {
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
    const { id, value, onChange, t } = this.props
    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={`${t("update_trait")}`}
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
              onSaveFunc={(data) => {
                onChange(id, data)
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

VdResourcePicker.propTypes = {
  id: PropTypes.string,
  value: PropTypes.object,
  onChange: PropTypes.func,
}

export default withTranslation()(VdResourcePicker)

const getItems = (rootObject = {}) => {
  const items = []

  items.push({
    label: "name",
    fieldId: "name",
    fieldType: FieldType.Text,
    dataType: DataType.String,
    value: "",
    isRequired: false,
    helperText: "",
    helperTextInvalid: "helper_text.invalid_name",
    validated: "default",
    validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
  })

  items.push({
    label: "trait",
    fieldId: "traitType",
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    value: "",
    isRequired: true,
    options: DeviceTraitOptions,
    validator: { isNotEmpty: {} },
  })

  items.push({
    label: "resource_type",
    fieldId: "resourceType",
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    value: "",
    isRequired: true,
    options: ResourceTypeOptions,
    validator: { isNotEmpty: {} },
    resetFields: { quickId: "" },
  })

  const resourceType = objectPath.get(rootObject, "resourceType", "")

  if (resourceType !== "") {
    const resourceAPI = getResourceOptionsAPI(resourceType)
    const resourceOptionValueFunc = getResourceOptionValueFunc(resourceType)
    const resourceFilterFunc = getResourceFilterFunc(resourceType)
    const resourceDescriptionFunc = getOptionsDescriptionFunc(resourceType)
    items.push({
      label: "resource",
      fieldId: "quickId",
      apiOptions: resourceAPI,
      optionValueFunc: resourceOptionValueFunc,
      fieldType: FieldType.SelectTypeAheadAsync,
      getFiltersFunc: resourceFilterFunc,
      getOptionsDescriptionFunc: resourceDescriptionFunc,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      options: [],
      validator: { isNotEmpty: {} },
    })
  }

  items.push(
    {
      label: "labels",
      fieldId: "!labels",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "labels",
      fieldType: FieldType.Labels,
      dataType: DataType.Object,
      value: "",
    }
  )

  return items
}
