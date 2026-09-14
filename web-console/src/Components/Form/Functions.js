import objectPath from "object-path"
import { DataType, FieldType } from "../../Constants/Form"
import v from "validator"

export const updateRootObject = (rootObject, item, data) => {
  const value = getValue(item, data)
  objectPath.set(rootObject, item.fieldId, value, false)
}

const getValue = (item, data) => {
  switch (item.dataType) {
    case DataType.String:
      return String(data)

    case DataType.Boolean:
      return Boolean(data)

    case DataType.Number:
      if (data === "" || data === "-") {
        return data
      }
      return Number(data)

    case DataType.Integer:
      if (data === "" || data === "-") {
        return data
      }
      return v.toInt(data)

    case DataType.Float:
      if (data === "" || data === "-") {
        return data
      }
      return v.toFloat(data)

    // case DataType.ArrayString:
    //   return data.map((d) => {
    //     String(d)
    //   })

    case DataType.ArrayNumber:
      return data.map((d) => Number(d))

    case DataType.ArrayBoolean:
      return data.map((d) => Boolean(d))

    case DataType.ArrayObject:
      return data

    case DataType.ArrayString:
      return Array.isArray(data) ? data : []

    case DataType.Object:
      return data

    default:
      return data
  }
}

export const updateItems = (rootObject, items) => {
  items.forEach((item) => {
    switch (item.fieldType) {
      case FieldType.Labels:
      case FieldType.KeyValueMap:
      case FieldType.VariablesMap:
      case FieldType.ThresholdsColor:
      case FieldType.DateRangePicker:
      case FieldType.TimeRangePicker:
      case FieldType.ChartYAxisConfigMap:
        item.value = objectPath.get(rootObject, item.fieldId, {})
        if (item.value === null) {
          item.value = {}
        }
        break

      case FieldType.DynamicArray:
      case FieldType.ConditionsArrayMap:
      case FieldType.MixedControlList:
      case FieldType.ChartMixedResourceConfig:
      case FieldType.DynamicListGeneric:
      case FieldType.PolicyStatements:
        item.value = objectPath.get(rootObject, item.fieldId, [])
        if (item.value === null) {
          item.value = []
        }
        break

      case FieldType.Switch:
      case FieldType.Checkbox:
        item.value = objectPath.get(rootObject, item.fieldId, false)
        break

      case FieldType.SelectTypeAhead:
      case FieldType.SelectTypeAheadAsync:
      case FieldType.SelectTypeAheadMultiple:
        if (item.dataType === DataType.ArrayString) {
          item.value = objectPath.get(rootObject, item.fieldId, [])
          if (item.value === null) {
            item.value = []
          }
        } else {
          item.value = String(objectPath.get(rootObject, item.fieldId, ""))
        }
        break

      case FieldType.ScriptEditor:
        if (item.dataType === DataType.Object) {
          item.value = objectPath.get(rootObject, item.fieldId, {})
          if (item.value === null) {
            item.value = {}
          }
        } else {
          item.value = String(objectPath.get(rootObject, item.fieldId, ""))
        }
        break

      // case FieldType.Select:
      // case FieldType.SelectTypeAhead:
      //   item.value = String(objectPath.get(rootObject, item.fieldId, ""))
      //   if (item.isRequired && item.value === "" && item.options.length === 1) {
      //     item.value = item.options[0].value
      //   }
      //   break

      default:
        item.value = String(objectPath.get(rootObject, item.fieldId, ""))
    }
  })
}
