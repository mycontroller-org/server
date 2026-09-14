import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import { getResourceItems } from "../Common/edit_resource"

// example data structure
// const data = {
//   type: "", // circle_size_100
//   resource: {
//     type: "",
//     filterType: "", // quick_id, detailed_filter
//     quickId: "",
//     displayName: "",
//     nameKey: "",
//     roundDecimal: "",
//     valueKey: "",
//     timestampKey: "",
//     unit: "",
//     filter: {},
//   },
//   chart: {
//     maximumValue: 100,
//     minimumValue: 0,
//     thickness: 20,
//     cornerSmoothing: 2,
//     thresholds: {
//       22: "#0066CC",
//       aa: "#0066CC",
//     },
//   },
// }

// Circle items
export const getCircleItems = (rootObject) => {
  // load default values
  objectPath.set(rootObject, "config.chart.minimumValue", 0, true)
  objectPath.set(rootObject, "config.chart.maximumValue", 100, true)
  objectPath.set(rootObject, "config.chart.thickness", 20, true)
  objectPath.set(rootObject, "config.chart.cornerSmoothing", 2, true)
  objectPath.set(rootObject, "config.chart.thresholds", {}, true)

  const items = []

  // include resource items
  const resourceItems = getResourceItems(rootObject)
  items.push(...resourceItems)

  // circle configuration
  items.push(
    {
      label: "configuration",
      fieldId: "!configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "minimum_value",
      fieldId: "config.chart.minimumValue",
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
      fieldId: "config.chart.maximumValue",
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
      label: "thickness",
      fieldId: "config.chart.thickness",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 1,
      max: 99,
      step: 1,
    },
    {
      label: "corner_smoothing",
      fieldId: "config.chart.cornerSmoothing",
      fieldType: FieldType.SliderSimple,
      dataType: DataType.Integer,
      value: "",
      min: 0,
      max: 100,
      step: 1,
    },
    {
      label: "thresholds_color",
      fieldId: "!thresholds",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "config.chart.thresholds",
      fieldType: FieldType.ThresholdsColor,
      dataType: DataType.Object,
      value: "",
    }
  )

  return items
}
