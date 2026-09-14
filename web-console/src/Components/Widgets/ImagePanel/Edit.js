import objectPath from "object-path"
import { DataType, FieldType } from "../../../Constants/Form"
import { RefreshIntervalType, RefreshIntervalTypeOptions } from "../../../Constants/Metric"
import { getQuickId, ResourceType } from "../../../Constants/ResourcePicker"
import {
  ImageRotationType,
  ImageRotationTypeOptions,
  ImageSourceType,
  ImageSourceTypeOptions,
  ImageType,
  ImageTypeOptions,
} from "../../../Constants/Widgets/ImagePanel"
import { api } from "../../../Service/Api"
import { getDynamicFilter } from "../../../Util/Filter"

// Image Panel items
export const updateFormItemsImagePanel = (rootObject, items = []) => {
  objectPath.set(rootObject, "config.rotation", ImageRotationType.Rotate_0, true)
  items.push(
    {
      label: "image_source",
      fieldId: "!source_type",
      fieldType: FieldType.Divider,
    },
    {
      label: "type",
      fieldId: "config.sourceType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ImageSourceTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { "config.field": {} },
    },
    {
      label: "rotation",
      fieldId: "config.rotation",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ImageRotationTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    }
  )

  const imageSourceType = objectPath.get(rootObject, "config.sourceType", "")

  switch (imageSourceType) {
    case ImageSourceType.Field:
      const simpleCameraItems = getFieldItems(rootObject)
      items.push(...simpleCameraItems)
      break

    case ImageSourceType.URL:
      const urlItems = getURLItems(rootObject)
      items.push(...urlItems)
      break

    case ImageSourceType.Disk:
      const diskItems = getDiskItems(rootObject)
      items.push(...diskItems)
      break

    default:
    // noop
  }
}

const getFieldItems = (rootObject) => {
  const items = [
    {
      label: "field",
      fieldId: "config.field.id",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: getFiltersFunc,
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    },
    {
      label: "name_key",
      fieldId: "config.field.nameKey",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "image_type",
      fieldId: "config.field.type",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ImageTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { "config.field.custom_mapping": {} },
    },
    {
      label: "display_value",
      fieldId: "config.field.displayValue",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
  ]

  const fieldType = objectPath.get(rootObject, "config.field.type", "")

  if (fieldType === ImageType.CustomMapping) {
    items.push(
      {
        label: "threshold_mode",
        fieldId: "config.field.thresholdMode",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      },
      {
        label: "custom_mapping",
        fieldId: "!custom_mapping",
        fieldType: FieldType.Divider,
      },
      {
        label: "",
        fieldId: "config.field.custom_mapping",
        fieldType: FieldType.KeyValueMap,
        dataType: DataType.ArrayObject,
        value: "",
        keyLabel: "value",
        valueLabel: "icon_or_image_location",
      }
    )
  }
  return items
}

const getURLItems = (rootObject) => {
  objectPath.set(rootObject, "config.refreshInterval", RefreshIntervalType.None, true)
  const items = [
    {
      label: "refresh_interval",
      fieldId: "config.refreshInterval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.Integer,
      value: "",
      options: RefreshIntervalTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "url",
      fieldId: "config.imageURL",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_url",
      validated: "default",
      validator: { isURL: {}, isNotEmpty: {} },
    },
  ]
  return items
}

const getDiskItems = (rootObject) => {
  objectPath.set(rootObject, "config.refreshInterval", RefreshIntervalType.None, true)
  const items = [
    {
      label: "refresh_interval",
      fieldId: "config.refreshInterval",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.Integer,
      value: "",
      options: RefreshIntervalTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "location",
      fieldId: "config.imageLocation",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_location",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
  ]
  return items
}

// helper functions

const getFiltersFunc = (value) => {
  return getDynamicFilter("name", value, [])
}

const getOptionsDescriptionFuncImpl = (item) => {
  return item.name
}

const getResourceOptionValueFunc = (item) => {
  return getQuickId(ResourceType.Field, item)
}
