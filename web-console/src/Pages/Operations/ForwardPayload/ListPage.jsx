import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { getStatusBool } from "../../../Components/Icons/Icons"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
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
} from "../../../store/entities/operations/forwardPayload"

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
        this.actionFuncWithRefresh(api.forwardPayload.enable)
      },
    },
    {
      type: "disable",
      onClick: () => {
        this.actionFuncWithRefresh(api.forwardPayload.disable)
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
        r(this.props.history, rMap.operations.forwardPayload.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="forward_payload" />
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
  { title: "source_field", fieldKey: "srcFieldId", sortable: true },
  { title: "destination_field", fieldKey: "dstFieldId", sortable: true },
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
              r(history, rMap.operations.forwardPayload.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      { title: <div className="align-center">{getStatusBool(rawData.enabled)}</div> },
      rawData.srcFieldId,
      rawData.dstFieldId,
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "enabled", categoryName: "enabled", fieldType: "enabled", dataType: "boolean" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
  { category: "srcFieldId", categoryName: "source_field", fieldType: "input", dataType: "string" },
  { category: "dstFieldId", categoryName: "destination_field", fieldType: "input", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.forwardPayload.list,
  apiDeleteRecords: api.forwardPayload.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_forward_payload",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.operationForwardPayload.loading,
  records: state.entities.operationForwardPayload.records,
  pagination: state.entities.operationForwardPayload.pagination,
  count: state.entities.operationForwardPayload.count,
  lastUpdate: state.entities.operationForwardPayload.lastUpdate,
  revision: state.entities.operationForwardPayload.revision,
  filters: state.entities.operationForwardPayload.filters,
  sortBy: state.entities.operationForwardPayload.sortBy,
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
