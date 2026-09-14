import { DropDownPositionType } from "../Constants/Common"
import { DataType, FieldType } from "../Constants/Form"
import { Operator } from "../Constants/Filter"
import { api } from "../Service/Api"

export const FieldHandlersList = {
  label: "",
  fieldId: "handlers",
  fieldType: FieldType.DynamicArray,
  dataType: DataType.ArrayString,
  value: [],
  isSelectTypeAheadAsync: true,
  selectOptions: {
    direction: DropDownPositionType.UP,
    isCreatable: true,
    createText: "",
    apiOptions: api.handler.list,
    optionValueKey: "id",
    getFiltersFunc: (value) => {
      return [{ k: "id", o: Operator.Regex, v: value }]
    },
    optionValueFunc: (item) => {
      return item.id
    },
    getOptionsDescriptionFunc: (item) => {
      return item.description
    },
  },
}
