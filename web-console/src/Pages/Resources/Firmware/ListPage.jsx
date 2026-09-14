import { Button } from "@patternfly/react-core"
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
} from "../../../store/entities/resources/firmware"

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
        r(this.props.history, rMap.resources.firmware.add)
      },
    },
  ]

  render() {
    return (
      <>
        <PageTitle title="firmwares" />
        <PageContent>{super.render()}</PageContent>
      </>
    )
  }
}

// Properties definition
const tableColumns = [
  { title: "id", fieldKey: "id", sortable: true },
  { title: "description", fieldKey: "description", sortable: true },
  { title: "filename", fieldKey: "file.name", sortable: true },
  { title: "last_modified", fieldKey: "lastModified", sortable: true },
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
              r(history, rMap.resources.firmware.detail, { id: rawData.id })
            }}
          >
            {rawData.id}
          </Button>
        ),
      },
      rawData.description,
      rawData.file.name,
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "id", categoryName: "id", fieldType: "input", dataType: "string" },
  { category: "description", categoryName: "description", fieldType: "input", dataType: "string" },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.firmware.list,
  apiDeleteRecords: api.firmware.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_firmware",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.resourceFirmware.loading,
  records: state.entities.resourceFirmware.records,
  pagination: state.entities.resourceFirmware.pagination,
  count: state.entities.resourceFirmware.count,
  lastUpdate: state.entities.resourceFirmware.lastUpdate,
  revision: state.entities.resourceFirmware.revision,
  filters: state.entities.resourceFirmware.filters,
  sortBy: state.entities.resourceFirmware.sortBy,
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
