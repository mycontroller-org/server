import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { getStatus, getStatusBool } from "../../../Components/Icons/Icons"
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
} from "../../../store/entities/resources/virtualAssistant"
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
        this.actionFuncWithRefresh(api.virtualAssistant.enable)
      },
    },
    {
      type: "disable",
      onClick: () => {
        this.actionFuncWithRefresh(api.virtualAssistant.disable)
      },
    },
    {
      type: "reload",
      onClick: () => {
        this.actionFuncWithRefresh(api.virtualAssistant.reload)
      },
    },
    { type: "separator" },
    { type: "delete", onClick: this.onDeleteActionClick },
  ]

  toolbar = [
    { type: "refresh", group: "right1" },
    { type: "actions", group: "right1", actions: this.actions, disabled: false },
    {
      type: "addButton",
      group: "right1",
      onClick: () => {
        r(this.props.history, rMap.resources.virtualAssistant.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="virtual_assistants" />
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
  { title: "provider", fieldKey: "provider.type", sortable: true },
  { title: "status", fieldKey: "state.status", sortable: true },
  { title: "since", fieldKey: "state.since", sortable: true },
  { title: "message", fieldKey: "state.message", sortable: true },
  { title: "modified_on", fieldKey: "modifiedOn", sortable: true },
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
              r(history, rMap.resources.virtualAssistant.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      { title: <div className="align-center">{getStatusBool(rawData.enabled)}</div> },
      getValue(rawData, "providerType", ""),
      { title: getStatus(getValue(rawData, "state.status", "")) },
      { title: <LastSeen date={getValue(rawData, "state.since", "")} /> },
      getValue(rawData, "state.message", ""),
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "enabled", categoryName: "enabled", fieldType: "enabled", dataType: "boolean" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.virtualAssistant.list,
  apiDeleteRecords: api.virtualAssistant.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_virtual_assistant",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.resourceVirtualAssistant.loading,
  records: state.entities.resourceVirtualAssistant.records,
  pagination: state.entities.resourceVirtualAssistant.pagination,
  count: state.entities.resourceVirtualAssistant.count,
  lastUpdate: state.entities.resourceVirtualAssistant.lastUpdate,
  revision: state.entities.resourceVirtualAssistant.revision,
  filters: state.entities.resourceVirtualAssistant.filters,
  sortBy: state.entities.resourceVirtualAssistant.sortBy,
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
