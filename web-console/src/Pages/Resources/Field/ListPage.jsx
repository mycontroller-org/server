import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { LastSeen } from "../../../Components/Time/Time"
import InputField from "../../../Components/Widgets/ControlPanel/Common/InputField"
import { getQuickId, ResourceType } from "../../../Constants/ResourcePicker"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import {
  deleteAllFilter,
  deleteFilterCategory,
  deleteFilterValue,
  loading,
  loadingFailed,
  onSortBy,
  updateFilter,
  updateRecords,
} from "../../../store/entities/resources/field"
import { getFieldValue, getValue } from "../../../Util/Util"

class List extends ListBase {
  state = {
    loading: true,
    pagination: {
      limit: 10,
      page: 0,
    },
    rows: [],
  }

  componentDidMount() {
    super.componentDidMount()
  }

  actions = [{ type: "delete", onClick: this.onDeleteActionClick }]

  toolbar = [
    { type: "refresh", group: "right1" },
    { type: "actions", group: "right1", actions: this.actions, disabled: false },
    {
      type: "addButton",
      group: "right1",
      onClick: () => {
        r(this.props.history, rMap.resources.field.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="fields" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

// Properties definition
const tableColumns = [
  { title: "gateway_id", fieldKey: "gatewayId", sortable: true },
  { title: "node_id", fieldKey: "nodeId", sortable: true },
  { title: "source_id", fieldKey: "sourceId", sortable: true },
  { title: "field_id", fieldKey: "fieldId", sortable: true },
  { title: "name", fieldKey: "name", sortable: true },
  { title: "metric_type", fieldKey: "metricType", sortable: true },
  { title: "value", fieldKey: "current.value", sortable: true },
  { title: "previous_value", fieldKey: "previous.value", sortable: true },
  { title: "unit", fieldKey: "unit", sortable: true },
  { title: "last_seen", fieldKey: "lastSeen", sortable: true },
]

const toRowFuncImpl = (rawData, history) => {
  const isReadOnly = getValue(rawData, "labels.read_only", "false") === "true"
  const currentValue = isReadOnly ? (
    getFieldValue(getValue(rawData, "current.value", ""))
  ) : (
    <InputField
      payload={getValue(rawData, "current.value", "")}
      id={rawData.id}
      quickId={getQuickId(ResourceType.Field, rawData)}
      widgetId={rawData.id}
      key={rawData.id}
      sendPayloadWrapper={(callBack) => {
        callBack()
      }}
    />
  )

  return {
    cells: [
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.resources.gateway.detail, { id: rawData.gatewayId })
            }}
          >
            {rawData.gatewayId}
          </Button>
        ),
      },
      rawData.nodeId,
      rawData.sourceId,
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.resources.field.detail, { id: rawData.id })
            }}
          >
            {rawData.fieldId}
          </Button>
        ),
      },
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.resources.field.detail, { id: rawData.id })
            }}
          >
            {rawData.name}
          </Button>
        ),
      },
      rawData.metricType,
      { title: currentValue },
      getFieldValue(getValue(rawData, "previous.value", "")),
      rawData.unit,
      { title: <LastSeen date={rawData.lastSeen} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "name", categoryName: "name", fieldType: "input", dataType: "string" },
  { category: "gatewayId", categoryName: "gateway_id", fieldType: "input", dataType: "string" },
  { category: "nodeId", categoryName: "node_id", fieldType: "input", dataType: "string" },
  { category: "sourceId", categoryName: "source_id", fieldType: "input", dataType: "string" },
  { category: "fieldId", categoryName: "field_id", fieldType: "input", dataType: "string" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.field.list,
  apiDeleteRecords: api.field.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_field",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.resourceField.loading,
  records: state.entities.resourceField.records,
  pagination: state.entities.resourceField.pagination,
  count: state.entities.resourceField.count,
  lastUpdate: state.entities.resourceField.lastUpdate,
  revision: state.entities.resourceField.revision,
  filters: state.entities.resourceField.filters,
  sortBy: state.entities.resourceField.sortBy,
})

const mapDispatchToProps = (dispatch) => ({
  updateRecordsFunc: (data) => dispatch(updateRecords(data)),
  loadingFunc: () => dispatch(loading()),
  loadingFailedFunc: () => dispatch(loadingFailed()),
  updateFilterFunc: (data) => dispatch(updateFilter(data)),
  deleteFilterValueFunc: (data) => dispatch(deleteFilterValue(data)),
  deleteFilterCategoryFunc: (data) => dispatch(deleteFilterCategory(data)),
  deleteAllFilterFunc: () => dispatch(deleteAllFilter()),
  onSortByFunc: (data) => dispatch(onSortBy(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(List))
