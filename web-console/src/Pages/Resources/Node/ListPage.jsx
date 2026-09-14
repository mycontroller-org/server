import { Button } from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { NodeRebootDialog, NodeResetDialog } from "../../../Components/Dialog/Dialog"
import { getStatus } from "../../../Components/Icons/Icons"
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
} from "../../../store/entities/resources/node"
import { getValue } from "../../../Util/Util"

class List extends ListBase {
  state = {
    loading: true,
    pagination: {
      limit: 10,
      page: 0,
    },
    rows: [],
    showDialogReboot: false,
    showDialogReset: false,
  }

  componentDidMount() {
    super.componentDidMount()
  }

  showRebootDialog = () => {
    this.setState({ showDialogReboot: true, showDialogReset: false })
  }

  showResetDialog = () => {
    this.setState({ showDialogReboot: false, showDialogReset: true })
  }

  hideDialog = () => {
    this.setState({ showDialogReboot: false, showDialogReset: false })
  }

  actions = [
    { type: "edit", disabled: true },
    { type: "delete", onClick: this.onDeleteActionClick },
  ]

  actions = [
    {
      type: "refresh_node_info",
      onClick: this.actionFunc((ids) => {
        api.action.nodeAction({ action: "refresh_node_info", id: ids })
      }),
    },
    {
      type: "heartbeat_request",
      onClick: this.actionFunc((ids) => {
        api.action.nodeAction({ action: "heartbeat_request", id: ids })
      }),
    },
    {
      type: "reboot",
      onClick: this.showRebootDialog,
    },
    { type: "separator" },
    {
      type: "firmware_update",
      onClick: this.actionFunc((ids) => {
        api.action.nodeAction({ action: "firmware_update", id: ids })
      }),
    },
    {
      type: "reset",
      onClick: this.showResetDialog,
    },
    { type: "separator" },
    {
      type: "delete",
      onClick: this.onDeleteActionClick,
    },
  ]

  toolbar = [
    { type: "refresh", group: "right1" },
    { type: "actions", group: "right1", actions: this.actions, disabled: false },
    {
      type: "addButton",
      group: "right1",
      onClick: () => {
        r(this.props.history, rMap.resources.node.add)
      },
    },
  ]

  render() {
    const { showDialogReboot, showDialogReset } = this.state
    return (
      <>
        <PageTitle title="nodes" />
        <PageContent>
          {super.render()}
          <NodeRebootDialog
            show={showDialogReboot}
            onCloseFn={this.hideDialog}
            onOkFn={this.actionFunc((ids) => {
              api.action.nodeAction({ action: "reboot", id: ids })
              this.hideDialog()
            })}
          />
          <NodeResetDialog
            show={showDialogReset}
            onCloseFn={this.hideDialog}
            onOkFn={this.actionFunc((ids) => {
              api.action.nodeAction({ action: "reset", id: ids })
              this.hideDialog()
            })}
          />
        </PageContent>
      </>
    )
  }
}

// Properties definition
const tableColumns = [
  { title: "gateway_id", fieldKey: "gatewayId", sortable: true },
  { title: "node_id", fieldKey: "nodeId", sortable: true },
  { title: "name", fieldKey: "name", sortable: true },
  { title: "version", fieldKey: "labels.version", sortable: true },
  { title: "library_version", fieldKey: "labels.library_version", sortable: true },
  { title: "battery", fieldKey: "others.battery_level", sortable: true },
  { title: "status", fieldKey: "state.status", sortable: true },
  { title: "last_seen", fieldKey: "lastSeen", sortable: true },
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
              r(history, rMap.resources.gateway.detail, { id: rawData.gatewayId })
            }}
          >
            {rawData.gatewayId}
          </Button>
        ),
      },
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.resources.node.detail, { id: rawData.id })
            }}
          >
            {rawData.nodeId}
          </Button>
        ),
      },
      {
        title: (
          <Button
            variant="link"
            isInline
            onClick={(_e) => {
              r(history, rMap.resources.node.detail, { id: rawData.id })
            }}
          >
            {rawData.name}
          </Button>
        ),
      },
      rawData.labels.version,
      rawData.labels.library_version,
      { title: getValue(rawData, "others.battery_level", "") },
      { title: getStatus(rawData.state.status) },
      { title: <LastSeen date={rawData.lastSeen} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "name", categoryName: "name", fieldType: "input", dataType: "string" },
  { category: "gatewayId", categoryName: "gateway_id", fieldType: "input", dataType: "string" },
  { category: "nodeId", categoryName: "node_id", fieldType: "input", dataType: "string" },
  { category: "labels.version", categoryName: "version", fieldType: "input", dataType: "string" },
  {
    category: "labels.library_version",
    categoryName: "library_version",
    fieldType: "input",
    dataType: "string",
  },
  { category: "labels", categoryName: "labels", fieldType: "label", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.node.list,
  apiDeleteRecords: api.node.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  deleteDialogTitle: "dialog.delete_title_node",
  filtersDefinition: filtersDefinition,
}

const mapStateToProps = (state) => ({
  loading: state.entities.resourceNode.loading,
  records: state.entities.resourceNode.records,
  pagination: state.entities.resourceNode.pagination,
  count: state.entities.resourceNode.count,
  lastUpdate: state.entities.resourceNode.lastUpdate,
  revision: state.entities.resourceNode.revision,
  filters: state.entities.resourceNode.filters,
  sortBy: state.entities.resourceNode.sortBy,
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
