import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import ListBase from "../../../Components/BasePage/ListBase"
import { FileSize } from "../../../Components/DataDisplay/Miscellaneous"
import { RestoreDialog } from "../../../Components/Dialog/Dialog"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { LastSeen } from "../../../Components/Time/Time"
import { BackupProviderType } from "../../../Constants/ResourcePicker"
import { api, SystemApiKey } from "../../../Service/Api"
import {
  deleteAllFilter,
  deleteFilterCategory,
  deleteFilterValue,
  loading,
  loadingFailed,
  onSortBy,
  updateFilter,
  updateRecords,
} from "../../../store/entities/system/backup"
import CustomModal from "./Helper"

class List extends ListBase {
  state = {
    loading: true,
    pagination: {
      limit: 10,
      page: 0,
    },
    rows: [],
    backupLocations: [],
    showModel: false,
    modalType: "",
    showRestoreWarning: false,
    backupFileName: "",
  }

  componentDidMount() {
    super.componentDidMount()
    api.settings
      .getBackupLocations()
      .then((res) => {
        const locations = {}
        res.data.spec.locations.forEach((loc) => {
          if (loc.type === BackupProviderType.Disk) {
            locations[loc.name] = loc.config.targetDirectory
          }
        })
        this.setState({ backupLocations: locations })
      })
      .catch((_e) => {})
  }

  hideImportWarning = () => {
    this.setState({ showRestoreWarning: false })
  }

  showRestoreWarning = (rows) => {
    if (rows.length !== 1) {
      return
    }
    this.setState({ showRestoreWarning: true, backupFileName: rows[0] })
  }

  runRestore = (backupFileName) => {
    if (backupFileName === undefined || backupFileName === "") {
      return
    }
    api.backup
      .runRestore({ id: backupFileName })
      .then((_res) => {
        // import success?
      })
      .catch((_e) => {
        // error on import?
      })
    this.setState({ showRestoreWarning: false })
  }

  actions = [
    { type: "restore", onClick: this.actionFunc(this.showRestoreWarning) },
    { type: "delete", onClick: this.onDeleteActionClick },
  ]

  toolbar = [
    { type: "refresh", group: "right1" },
    { type: "actions", group: "right1", actions: this.actions, disabled: false },
    {
      type: "customButton",
      group: "right1",
      text: "Backup Locations",
      isSmall: false,
      onClick: () => {
        this.setState({ showModel: true, modalType: "backup_locations" })
      },
    },
    {
      type: "customButton",
      group: "right1",
      text: "Backup",
      isSmall: false,
      variant: "secondary",
      onClick: () => {
        this.setState({ showModel: true, modalType: "run_backup" })
      },
    },
  ]

  saveBackupLocations = (rawLocations = {}) => {
    const names = Object.keys(rawLocations.locations)
    const locations = names.map((name) => {
      return {
        name: name,
        type: BackupProviderType.Disk,
        config: {
          targetDirectory: rawLocations.locations[name],
        },
      }
    })

    this.setState({ backupLocations: rawLocations.locations }, () => {
      const data = {
        id: SystemApiKey.BackupLocations,
        spec: {
          locations: locations,
        },
      }
      api.settings
        .updateSystem(data)
        .then((_res) => {
          this.setState({ showModel: false })
        })
        .catch((_e) => {
          // failure
        })
    })
  }

  closeModal = () => {
    this.setState({ showModel: false })
  }

  render() {
    const { showModel, modalType, backupLocations, showRestoreWarning, backupFileName } = this.state
    return (
      <>
        <PageTitle title="backup_and_restore" />
        <PageContent>{super.render()}</PageContent>
        <CustomModal
          isOpen={showModel}
          modalType={modalType}
          onClose={this.closeModal}
          backupLocations={backupLocations}
          saveBackupLocationsFunc={this.saveBackupLocations}
        />
        <RestoreDialog
          show={showRestoreWarning}
          fileName={backupFileName}
          onCloseFn={this.hideImportWarning}
          onOkFn={() => this.runRestore(backupFileName)}
        />
      </>
    )
  }
}

// Properties definition

const tableColumns = [
  { title: "location_name", fieldKey: "locationName", sortable: true },
  { title: "filename", fieldKey: "fileName", sortable: true },
  { title: "size", fieldKey: "fileSize", sortable: true },
  { title: "modified_on", fieldKey: "modifiedOn", sortable: true },
]

const toRowFuncImpl = (rawData, _history) => {
  return {
    cells: [
      rawData.locationName,
      rawData.fileName,
      { title: <FileSize bytes={rawData.fileSize} /> },
      { title: <LastSeen date={rawData.modifiedOn} /> },
    ],
    rid: rawData.id,
  }
}

const filtersDefinition = [
  { category: "fileName", categoryName: "filename", fieldType: "input", dataType: "string" },
  { category: "locationName", categoryName: "location_name", fieldType: "input", dataType: "string" },
]

// supply required properties
List.defaultProps = {
  apiGetRecords: api.backup.list,
  apiDeleteRecords: api.backup.delete,
  tableColumns: tableColumns,
  toRowFunc: toRowFuncImpl,
  resourceName: "Backup and Restore",
  filtersDefinition: filtersDefinition,
  deleteDialogTitle: "dialog.delete_title_backup",
}

const mapStateToProps = (state) => ({
  loading: state.entities.systemBackup.loading,
  records: state.entities.systemBackup.records,
  pagination: state.entities.systemBackup.pagination,
  count: state.entities.systemBackup.count,
  lastUpdate: state.entities.systemBackup.lastUpdate,
  revision: state.entities.systemBackup.revision,
  filters: state.entities.systemBackup.filters,
  sortBy: state.entities.systemBackup.sortBy,
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
