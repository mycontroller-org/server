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
} from "../../../store/entities/system/user"

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
        r(this.props.history, rMap.settings.user.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="users" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

const tableColumns = [
  { title: "username", fieldKey: "username", sortable: true },
  { title: "full_name", fieldKey: "fullName", sortable: true },
  { title: "email", fieldKey: "email", sortable: true },
  { title: "enabled", fieldKey: "disabled", sortable: true },
  { title: "policies", fieldKey: "policies", sortable: false },
  { title: "modified_on", fieldKey: "modifiedOn", sortable: true },
]

const toRowFuncImpl = (rawData, history) => {
  const policies = Array.isArray(rawData.policies) ? rawData.policies.join(", ") : ""
  return {
    cells: [
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.settings.user.detail, { id: rawData.id })
            }}
          >
            {rawData.username}
          </Button>
        ),
      },
      { title: rawData.fullName },
      { title: rawData.email },
      { title: <div className="align-center">{getStatusBool(!rawData.disabled)}</div> },
      { title: policies },
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "username", categoryName: "username", fieldType: "input", dataType: "string" },
  { category: "fullName", categoryName: "full_name", fieldType: "input", dataType: "string" },
  { category: "email", categoryName: "email", fieldType: "input", dataType: "string" },
  { category: "disabled", categoryName: "disabled", fieldType: "enabled", dataType: "boolean" },
]

List.defaultProps = {
  apiGetRecords: api.user.list,
  apiDeleteRecords: api.user.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_user",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.settingsUser.loading,
  records: state.entities.settingsUser.records,
  pagination: state.entities.settingsUser.pagination,
  count: state.entities.settingsUser.count,
  lastUpdate: state.entities.settingsUser.lastUpdate,
  revision: state.entities.settingsUser.revision,
  filters: state.entities.settingsUser.filters,
  sortBy: state.entities.settingsUser.sortBy,
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
