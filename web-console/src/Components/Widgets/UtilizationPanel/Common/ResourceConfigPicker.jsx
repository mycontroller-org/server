import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { DataType, FieldType } from "../../../../Constants/Form"
import { ResourceFilterType, ResourceType, ResourceTypeOptions } from "../../../../Constants/Resource"
import { getValue } from "../../../../Util/Util"
import Editor from "../../../Editor/Editor"
import ErrorBoundary from "../../../ErrorBoundary/ErrorBoundary"
import {
  getOptionsDescriptionFunc,
  getResourceFilterFunc,
  getResourceOptionsAPI,
  getResourceOptionValueFunc
} from "../../../Form/ResourcePicker/ResourceUtils"

class ResourceConfigPicker extends React.Component {
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
    const { value = {}, index = 0, onChange = () => {}, isTablePanel = false } = this.props
    return (
      <>
        <Button key={"edit-btn-" + index} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + index}
          title={"update_resource"}
          variant={ModalVariant.medium}
          position="top"
          isOpen={isOpen}
          onClose={this.onClose}
          onEscapePress={this.onClose}
        >
          <ErrorBoundary>
            <Editor
              key={"editor" + index}
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
              getFormItems={(rootObject) => getItems(rootObject, isTablePanel)}
              saveButtonText="update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

ResourceConfigPicker.propTypes = {
  index: PropTypes.number,
  value: PropTypes.object,
  onChange: PropTypes.func,
  isTablePanel: PropTypes.bool,
}

export default ResourceConfigPicker

const getItems = (rootObject, isTablePanel = false) => {
  objectPath.set(rootObject, "type", ResourceType.Field, true)
  objectPath.set(rootObject, "filterType", ResourceFilterType.QuickID, true)
  objectPath.set(rootObject, "quickId", "", true)
  objectPath.set(rootObject, "displayName", true, false) // force display name
  objectPath.set(rootObject, "nameKey", "name", true)
  objectPath.set(rootObject, "timestampKey", "current.timestamp", true)
  objectPath.set(rootObject, "valueKey", "current.value", true)
  objectPath.set(rootObject, "roundDecimal", 0, true)
  objectPath.set(rootObject, "unit", "", true)

  const items = [
    {
      label: "type",
      fieldId: "type",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ResourceTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { quickId: "" },
    },
  ]

  const resourceType = objectPath.get(rootObject, "type", "")

  if (resourceType !== "") {
    const resourceAPI = getResourceOptionsAPI(resourceType)
    const resourceOptionValueFunc = getResourceOptionValueFunc(resourceType)
    const resourceFilterFunc = getResourceFilterFunc(resourceType)
    const resourceDescriptionFunc = getOptionsDescriptionFunc(resourceType)
    items.push(
      {
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
      },
      {
        label: "name_key",
        fieldId: "nameKey",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_key",
        validated: "default",
        validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
      },
      {
        label: "value_key",
        fieldId: "valueKey",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_key",
        validated: "default",
        validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
      },
      {
        label: "timestamp_key",
        fieldId: "timestampKey",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_key",
        validated: "default",
        validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
      },
      {
        label: "round_decimal",
        fieldId: "roundDecimal",
        fieldType: FieldType.Text,
        dataType: DataType.Integer,
        value: "",
        isRequired: false,
      },
      {
        label: "unit",
        fieldId: "unit",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: false,
      }
    )

    if (isTablePanel) {
      objectPath.set(rootObject, "sortOrderPriority", "1", true)
      items.push({
        label: "sort_order_priority",
        fieldId: "sortOrderPriority",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: false,
      })
    }

    if (isTablePanel) {
      objectPath.set(rootObject, "table.useGlobal", true, true)
      items.push(
        {
          label: "status_configuration",
          fieldId: "!status_configuration",
          fieldType: FieldType.Divider,
        },
        {
          label: "use_global",
          fieldId: "table.useGlobal",
          fieldType: FieldType.Switch,
          dataType: DataType.Boolean,
          value: false,
        }
      )

      const useGlobal = getValue(rootObject, "table.useGlobal", false)
      if (!useGlobal) {
        objectPath.set(rootObject, "table.hideValueColumn", false, true)
        objectPath.set(rootObject, "table.hideStatusColumn", false, true)
        objectPath.set(rootObject, "table.displayStatusPercentage", true, true)
        objectPath.set(rootObject, "table.minimumValue", 0, true)
        objectPath.set(rootObject, "table.maximumValue", 100, true)
        objectPath.set(rootObject, "table.thresholds", {}, true)

        items.push(
          {
            label: "hide_value_column",
            fieldId: "table.hideValueColumn",
            fieldType: FieldType.Switch,
            dataType: DataType.Boolean,
            value: false,
            resetFields: {
              "table.hideStatusColumn": (ro) => {
                const hideValue = getValue(ro, "table.hideValueColumn", false)
                if (hideValue) {
                  return false
                }
                return getValue(ro, "table.hideStatusColumn", false)
              },
            },
          },
          {
            label: "hide_status_column",
            fieldId: "table.hideStatusColumn",
            fieldType: FieldType.Switch,
            dataType: DataType.Boolean,
            value: false,
            resetFields: {
              "table.hideValueColumn": (ro) => {
                const hideStatus = getValue(ro, "table.hideStatusColumn", false)
                if (hideStatus) {
                  return false
                }
                return getValue(ro, "table.hideValueColumn", false)
              },
            },
          }
        )

        const hideStatusColumn = getValue(rootObject, "table.hideStatusColumn", false)
        if (!hideStatusColumn) {
          items.push(
            {
              label: "display_percentage",
              fieldId: "table.displayStatusPercentage",
              fieldType: FieldType.Switch,
              dataType: DataType.Boolean,
              value: false,
            },
            {
              label: "minimum_value",
              fieldId: "table.minimumValue",
              fieldType: FieldType.Text,
              dataType: DataType.Number,
              value: "",
              isRequired: true,
              helperText: "",
              helperTextInvalid: "helper_text.invalid_value",
              validated: "default",
              validator: { isDecimal: { decimal_digits: 2 } },
            },
            {
              label: "maximum_value",
              fieldId: "table.maximumValue",
              fieldType: FieldType.Text,
              dataType: DataType.Number,
              value: "",
              isRequired: true,
              helperText: "",
              helperTextInvalid: "helper_text.invalid_value",
              validated: "default",
              validator: { isDecimal: { decimal_digits: 2 } },
            },
            {
              label: "thresholds",
              fieldId: "table.thresholds",
              fieldType: FieldType.ThresholdsColor,
              dataType: DataType.Object,
              value: "",
            }
          )
        }
      }
    }
  }

  return items
}
