import { Button } from "@patternfly/react-core"
import { CheckIcon } from "@patternfly/react-icons"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
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
} from "../../../store/entities/system/policy"

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
        r(this.props.history, rMap.settings.policy.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="policies" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

const tableColumns = [
  { title: "id", fieldKey: "id", sortable: true },
  { title: "description", fieldKey: "description", sortable: true },
  { title: "system", fieldKey: "system", sortable: true },
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
              r(history, rMap.settings.policy.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      {
        title: rawData.system ? (
          <div className="align-center">
            <CheckIcon />
          </div>
        ) : (
          ""
        ),
      },
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
    locked: !!rawData.system,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "system", categoryName: "system", fieldType: "enabled", dataType: "boolean" },
]

List.defaultProps = {
  apiGetRecords: api.policy.list,
  apiDeleteRecords: api.policy.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_policy",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.settingsPolicy.loading,
  records: state.entities.settingsPolicy.records,
  pagination: state.entities.settingsPolicy.pagination,
  count: state.entities.settingsPolicy.count,
  lastUpdate: state.entities.settingsPolicy.lastUpdate,
  revision: state.entities.settingsPolicy.revision,
  filters: state.entities.settingsPolicy.filters,
  sortBy: state.entities.settingsPolicy.sortBy,
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
