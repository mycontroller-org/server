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
} from "../../../store/entities/operations/handler"
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
        this.actionFuncWithRefresh(api.handler.enable)
      },
    },
    {
      type: "disable",
      onClick: () => {
        this.actionFuncWithRefresh(api.handler.disable)
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
        r(this.props.history, rMap.operations.handler.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="handlers" />
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
  { title: "type", fieldKey: "type", sortable: true },
  { title: "status", fieldKey: "state.status", sortable: true },
  { title: "status_at", fieldKey: "state.since", sortable: true },
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
              r(history, rMap.operations.handler.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      { title: <div className="align-center">{getStatusBool(rawData.enabled)}</div> },
      { title: rawData.type },
      { title: getValue(rawData, "state.status") },
      { title: <LastSeen date={rawData.state.since} /> },
      { title: getValue(rawData, "state.message") },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "enabled", categoryName: "enabled", fieldType: "enabled", dataType: "boolean" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.handler.list,
  apiDeleteRecords: api.handler.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_handler",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.operationHandler.loading,
  records: state.entities.operationHandler.records,
  pagination: state.entities.operationHandler.pagination,
  count: state.entities.operationHandler.count,
  lastUpdate: state.entities.operationHandler.lastUpdate,
  revision: state.entities.operationHandler.revision,
  filters: state.entities.operationHandler.filters,
  sortBy: state.entities.operationHandler.sortBy,
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
