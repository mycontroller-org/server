import { DataType, FieldType } from "../../../Constants/Form"
import { ChartType, ChartTypeOptions } from "../../../Constants/Widgets/UtilizationPanel"
import { getValue } from "../../../Util/Util"
import { getCircleItems } from "./Circle/Edit"
import { getSparkItems } from "./Spark/Edit"
import { getTableItems } from "./Table/Edit"

// UtilizationPanel items
export const getUtilizationPanelItems = (rootObject) => {
  // update chart config items
  const subType = getValue(rootObject, "config.type", "")

  const items = [
    {
      label: "sub_type",
      fieldId: "config.type",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: ChartTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
      resetFields: {
        "config.resource.isMixedResources": false,
        "config.resource.resources": [],
        "config.table": {},
      },
    },
  ]

  switch (subType) {
    case ChartType.CircleSize50:
    case ChartType.CircleSize75:
    case ChartType.CircleSize100:
      const circleItems = getCircleItems(rootObject)
      items.push(...circleItems)
      break

    case ChartType.Table:
      const tableItems = getTableItems(rootObject)
      items.push(...tableItems)
      break

    case ChartType.SparkArea:
    case ChartType.SparkLine:
    case ChartType.SparkBar:
      const sparkItems = getSparkItems(rootObject)
      items.push(...sparkItems)
      break

    default:
    // noop
  }

  return items
}
