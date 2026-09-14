import { TextInput } from "@patternfly/react-core"
import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import {
  ResourceFilterType,
  ResourceFilterTypeOptions,
  ResourceType,
  ResourceTypeOptions,
} from "../../../../Constants/Resource"
import { getValue } from "../../../../Util/Util"
import {
  getOptionsDescriptionFunc,
  getResourceFilterFunc,
  getResourceOptionsAPI,
  getResourceOptionValueFunc,
} from "../../../Form/ResourcePicker/ResourceUtils"
import ResourceConfigPicker from "./ResourceConfigPicker"

export const getResourceItems = (rootObject, isTablePanel = false) => {
  // load default values
  objectPath.set(rootObject, "config.resource.type", ResourceType.Field, true)
  objectPath.set(rootObject, "config.resource.filterType", ResourceFilterType.QuickID, true)
  objectPath.set(rootObject, "config.resource.quickId", "", true)
  objectPath.set(rootObject, "config.resource.filter", {}, true)
  objectPath.set(rootObject, "config.resource.displayName", true, true)
  objectPath.set(rootObject, "config.resource.nameKey", "name", true)
  objectPath.set(rootObject, "config.resource.timestampKey", "current.timestamp", true)
  objectPath.set(rootObject, "config.resource.valueKey", "current.value", true)
  objectPath.set(rootObject, "config.resource.roundDecimal", 0, true)
  objectPath.set(rootObject, "config.resource.unit", "", true)

  // resource
  const items = [
    {
      label: "resource_configuration",
      fieldId: "!resource_configuration",
      fieldType: FieldType.Divider,
    },
  ]

  if (isTablePanel) {
    objectPath.set(rootObject, "config.resource.isMixedResources", false, true)
    objectPath.set(rootObject, "config.resource.filterType", ResourceFilterType.DetailedFilter, false)
    items.push({
      label: "mixed_resources",
      fieldId: "config.resource.isMixedResources",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    })
  }

  const isMixedResources = getValue(rootObject, "config.resource.isMixedResources", false)

  if (isMixedResources) {
    items.push({
      label: "resources",
      fieldId: "config.resource.resources",
      fieldType: FieldType.DynamicListGeneric,
      dataType: DataType.ArrayObject,
      value: [],
      isRequired: false,
      showUpdateButton: true,
      validateValueFunc: (value) => value.type !== undefined,
      valueField: getResourceConfigDisplayValue,
      updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
        resourceConfigUpdateButtonCallback(cbIndex, cbItem, cbOnChange, isTablePanel),
    })
  } else {
    items.push({
      label: "type",
      fieldId: "config.resource.type",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ResourceTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { "config.resource.quickId": "" },
    })
    if (!isTablePanel) {
      items.push({
        label: "filter_type",
        fieldId: "config.resource.filterType",
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        value: "",
        options: ResourceFilterTypeOptions,
        isRequired: true,
        validator: { isNotEmpty: {} },
      })
    }

    // resource filter
    const filterType = objectPath.get(rootObject, "config.resource.filterType", ResourceFilterType.QuickID)
    const resourceType = objectPath.get(rootObject, "config.resource.type", "")

    if (filterType === ResourceFilterType.DetailedFilter) {
      items.push({
        label: "filter",
        fieldId: "config.resource.filter",
        fieldType: FieldType.KeyValueMap,
        dataType: DataType.Object,
        value: "",
      })
    } else if (resourceType !== "") {
      const resourceAPI = getResourceOptionsAPI(resourceType)
      const resourceOptionValueFunc = getResourceOptionValueFunc(resourceType)
      const resourceFilterFunc = getResourceFilterFunc(resourceType)
      const resourceDescriptionFunc = getOptionsDescriptionFunc(resourceType)
      items.push({
        label: "resource",
        fieldId: "config.resource.quickId",
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

    if (isTablePanel) {
      objectPath.set(rootObject, "config.resource.displayName", true, false) // force display name
    } else {
      items.push({
        label: "display_name",
        fieldId: "config.resource.displayName",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      })
    }

    const displayName = objectPath.get(rootObject, "config.resource.displayName", false)
    if (displayName) {
      items.push({
        label: "name_key",
        fieldId: "config.resource.nameKey",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_key",
        validated: "default",
        validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
      })
    }

    items.push(
      {
        label: "value_key",
        fieldId: "config.resource.valueKey",
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
        fieldId: "config.resource.timestampKey",
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
        fieldId: "config.resource.roundDecimal",
        fieldType: FieldType.Text,
        dataType: DataType.Integer,
        value: "",
        isRequired: false,
      },
      {
        label: "unit",
        fieldId: "config.resource.unit",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: false,
      }
    )

    if (isTablePanel) {
      items.push({
        label: "limit",
        fieldId: "config.resource.limit",
        fieldType: FieldType.Text,
        dataType: DataType.Integer,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_limit",
        validated: "default",
        validator: { isInteger: {} },
      })
    }
  }
  return items
}

// helper functions
export const getResourceConfigDisplayValue = (index, value, _onChange, validated, _isDisabled) => {
  let displayValue = "undefined"
  if (value.quickId) {
    displayValue = `${value.type}: ${value.quickId}`
  }
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={displayValue}
      isDisabled={true}
      validated={validated}
    />
  )
}

// calls the model to update the resource config details
export const resourceConfigUpdateButtonCallback = (index = 0, item = {}, onChange, isTablePanel = false) => {
  return (
    <ResourceConfigPicker
      key={"picker_" + index}
      value={item}
      index={index}
      onChange={(newValue) => {
        onChange(index, newValue)
      }}
      isTablePanel={isTablePanel}
    />
  )
}
