import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { DisplayTrue } from "../../../Components/DataDisplay/Miscellaneous"
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
} from "../../../store/entities/resources/dataRepository"

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
        r(this.props.history, rMap.resources.dataRepository.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="data_repository" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

// Properties definition

const tableColumns = [
  { title: "id", fieldKey: "id", sortable: true },
  { title: "description", fieldKey: "description", sortable: true },
  { title: "read_only", fieldKey: "readOnly", sortable: true },
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
              r(history, rMap.resources.dataRepository.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      { title: rawData.description },
      { title: <DisplayTrue data={rawData} field="readOnly" /> },
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "name", categoryName: "name", fieldType: "input", dataType: "string" },
  { category: "readOnly", categoryName: "read_only", fieldType: "enabled", dataType: "boolean" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.dataRepository.list,
  apiDeleteRecords: api.dataRepository.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_data_repository",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.resourceDataRepository.loading,
  records: state.entities.resourceDataRepository.records,
  pagination: state.entities.resourceDataRepository.pagination,
  count: state.entities.resourceDataRepository.count,
  lastUpdate: state.entities.resourceDataRepository.lastUpdate,
  revision: state.entities.resourceDataRepository.revision,
  filters: state.entities.resourceDataRepository.filters,
  sortBy: state.entities.resourceDataRepository.sortBy,
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
