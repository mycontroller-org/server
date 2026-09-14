import objectPath from "object-path"
import { DataType, FieldType } from "../../../Constants/Form"
import { getQuickId, ResourceType } from "../../../Constants/ResourcePicker"
import {
  LightType,
  LightTypeOptions,
  RGBComponentOptions,
  RGBComponentType,
} from "../../../Constants/Widgets/LightPanel"
import { api } from "../../../Service/Api"
import { getDynamicFilter } from "../../../Util/Filter"

// Light Panel items
export const updateFormItemsLightPanel = (rootObject, items) => {
  items.push(
    {
      label: "light",
      fieldId: "!light",
      fieldType: FieldType.Divider,
    },
    {
      label: "type",
      fieldId: "config.lightType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: LightTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { "config.fieldIds": {}, "config.rgbComponent": "" },
    }
  )

  const lightType = objectPath.get(rootObject, "config.lightType", "")

  if (lightType === LightType.RGB || lightType === LightType.RGBCW || lightType === LightType.RGBCWWW) {
    items.push({
      label: "rgb_component",
      fieldId: "config.rgbComponent",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: RGBComponentOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: { "config.fieldIds.color": "", "config.fieldIds.hue": "", "config.fieldIds.rgb": "" },
    })
  }

  items.push(
    {
      label: "fields",
      fieldId: "!fields",
      fieldType: FieldType.Divider,
    },
    {
      label: "power",
      fieldId: "config.fieldIds.power",
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
      label: "brightness",
      fieldId: "config.fieldIds.brightness",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: getFiltersFunc,
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    }
  )

  // update CW, WW and RGB components
  if (lightType === LightType.CWWW || lightType === LightType.RGBCW || lightType === LightType.RGBCWWW) {
    items.push({
      label: "color_temperature",
      fieldId: "config.fieldIds.colorTemperature",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: getFiltersFunc,
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    })
  }

  if (lightType === LightType.RGB || lightType === LightType.RGBCW || lightType === LightType.RGBCWWW) {
    const rgbComponent = objectPath.get(rootObject, "config.rgbComponent", "")
    const label = rgbComponent === RGBComponentType.ColorPickerQuick ? "rgb" : "hue"
    const fieldId =
      rgbComponent === RGBComponentType.ColorPickerQuick ? "config.fieldIds.rgb" : "config.fieldIds.hue"

    if (rgbComponent !== "") {
      items.push({
        label: label,
        fieldId: fieldId,
        fieldType: FieldType.SelectTypeAheadAsync,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        validator: { isNotEmpty: {} },
        apiOptions: api.field.list,
        getFiltersFunc: getFiltersFunc,
        optionValueFunc: getResourceOptionValueFunc,
        getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
      })
    }

    // add alpha component
    items.push({
      label: "saturation",
      fieldId: "config.fieldIds.saturation",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      // validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: getFiltersFunc,
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    })
  }
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
