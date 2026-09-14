import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"
import { getValue } from "../../../../Util/Util"
import { getResourceItems } from "../Common/edit_resource"

// Table items
export const getTableItems = (rootObject) => {
  // load default values
  objectPath.set(rootObject, "config.table.minimumValue", 0, true)
  objectPath.set(rootObject, "config.table.maximumValue", 100, true)
  objectPath.set(rootObject, "config.table.displayStatusPercentage", true, true)
  objectPath.set(rootObject, "config.table.hideValueColumn", false, true)
  objectPath.set(rootObject, "config.table.hideStatusColumn", false, true)
  objectPath.set(rootObject, "config.table.hideBorder", false, true)
  objectPath.set(rootObject, "config.table.thresholds", {}, true)

  const items = []

  // include resource items
  const resourceItems = getResourceItems(rootObject, true)
  items.push(...resourceItems)

  // table configuration
  items.push(
    {
      label: "table_configuration",
      fieldId: "!table_configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "hide_border",
      fieldId: "config.hideBorder",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "hide_value_column",
      fieldId: "config.table.hideValueColumn",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
      resetFields: {
        "config.table.hideStatusColumn": (ro) => {
          const hideValue = getValue(ro, "config.table.hideValueColumn", false)
          if (hideValue) {
            return false
          }
          return getValue(ro, "config.table.hideStatusColumn", false)
        },
      },
    },
    {
      label: "hide_status_column",
      fieldId: "config.table.hideStatusColumn",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
      resetFields: {
        "config.table.hideValueColumn": (ro) => {
          const hideStatus = getValue(ro, "config.table.hideStatusColumn", false)
          if (hideStatus) {
            return false
          }
          return getValue(ro, "config.table.hideValueColumn", false)
        },
      },
    }
  )
  const hideStatusColumn = getValue(rootObject, "config.table.hideStatusColumn", false)
  if (!hideStatusColumn) {
    items.push(
      {
        label: "status_configuration",
        fieldId: "!status_configuration",
        fieldType: FieldType.Divider,
      },
      {
        label: "display_percentage",
        fieldId: "config.table.displayStatusPercentage",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      },
      {
        label: "minimum_value",
        fieldId: "config.table.minimumValue",
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
        fieldId: "config.table.maximumValue",
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
        fieldId: "config.table.thresholds",
        fieldType: FieldType.ThresholdsColor,
        dataType: DataType.Object,
        value: "",
      }
    )
  }

  return items
}
