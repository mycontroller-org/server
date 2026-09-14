import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { getStatusBool } from "../../../Components/Icons/Icons"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { LastSeen } from "../../../Components/Time/Time"
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
} from "../../../store/entities/operations/task"
import { getValue } from "../../../Util/Util"

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

  actions = [
    {
      type: "enable",
      onClick: () => {
        this.actionFuncWithRefresh(api.task.enable)
      },
    },
    {
      type: "disable",
      onClick: () => {
        this.actionFuncWithRefresh(api.task.disable)
      },
    },
    { type: "delete", onClick: this.onDeleteActionClick },
  ]

  toolbar = [
    { type: "refresh", group: "right1" },
    { type: "actions", group: "right1", actions: this.actions, disabled: false },
    {
      type: "addButton",
      group: "right1",
      onClick: () => {
        r(this.props.history, rMap.operations.task.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="tasks" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

// Properties definition

const tableColumns = [
  { title: "id", fieldKey: "id", sortable: true },
  { title: "description", fieldKey: "description", sortable: true },
  { title: "enabled", fieldKey: "enabled", sortable: true },
  { title: "ignore_duplicate", fieldKey: "ignoreDuplicate", sortable: true },
  { title: "auto_disable", fieldKey: "autoDisable", sortable: true },
  { title: "trigger_on_event", fieldKey: "triggerOnEvent", sortable: true },
  { title: "last_evaluation", fieldKey: "state.lastEvaluation", sortable: true },
  { title: "last_success", fieldKey: "state.lastSuccess", sortable: true },
  { title: "message", fieldKey: "state.message", sortable: true },
]

const toRowFuncImpl = (rawData, history) => {
  return {
    cells: [
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.operations.task.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      { title: <div className="align-center">{getStatusBool(rawData.enabled)}</div> },
      { title: <div className="align-center">{getStatusBool(rawData.ignoreDuplicate)}</div> },
      { title: <div className="align-center">{getStatusBool(rawData.autoDisable)}</div> },
      { title: <div className="align-center">{getStatusBool(rawData.triggerOnEvent)}</div> },
      { title: <LastSeen date={getValue(rawData, "state.lastEvaluation", "")} /> },
      { title: <LastSeen date={getValue(rawData, "state.lastSuccess", "")} /> },
      { title: getValue(rawData, "state.message", "") },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "enabled", categoryName: "enabled", fieldType: "enabled", dataType: "boolean" },
  {
    category: "ignoreDuplicate",
    categoryName: "ignore_duplicate",
    fieldType: "enabled",
    dataType: "boolean",
  },
  { category: "autoDisable", categoryName: "auto_disable", fieldType: "enabled", dataType: "boolean" },
  {
    category: "triggerOnEvent",
    categoryName: "trigger_on_event",
    fieldType: "enabled",
    dataType: "boolean",
  },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.task.list,
  apiDeleteRecords: api.task.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_task",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.operationTask.loading,
  records: state.entities.operationTask.records,
  pagination: state.entities.operationTask.pagination,
  count: state.entities.operationTask.count,
  lastUpdate: state.entities.operationTask.lastUpdate,
  revision: state.entities.operationTask.revision,
  filters: state.entities.operationTask.filters,
  sortBy: state.entities.operationTask.sortBy,
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
