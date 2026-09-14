import { Dropdown, DropdownItem, DropdownSeparator, DropdownToggle } from "@patternfly/react-core"
import {
  AddCircleOIcon,
  CaretDownIcon,
  CircleIcon,
  EditIcon,
  EraserIcon,
  FileImportIcon,
  HeartbeatIcon,
  InfoAltIcon,
  OutlinedCircleIcon,
  OutlinedTrashAltIcon,
  RebootingIcon,
  RetweetIcon,
  SearchIcon,
  UploadIcon,
} from "@patternfly/react-icons"
import React from "react"
import { withTranslation } from "react-i18next"
import DeleteDialog from "../Dialog/Dialog"

class Actions extends React.Component {
  state = {
    isOpen: false,
  }

  onToggle = (isOpen) => {
    this.setState({ isOpen })
  }
  onSelect = (_event) => {
    this.setState({ isOpen: !this.state.isOpen })
  }

  render() {
    const { isOpen } = this.state
    const { items, isDisabled, rowsSelectionCount = 1, t } = this.props

    const dropdownItems = items.map((item, index) => {
      switch (item.type) {
        case "new":
          return drawItem(item.type, t("new"), AddCircleOIcon, item.disabled, item.onClick)
        case "enable":
          return drawItem(item.type, t("enable"), CircleIcon, item.disabled, item.onClick)
        case "disable":
          return drawItem(item.type, t("disable"), OutlinedCircleIcon, item.disabled, item.onClick)
        case "reload":
          return drawItem(item.type, t("reload"), RetweetIcon, item.disabled, item.onClick)
        case "discover_nodes":
          return drawItem(item.type, t("discover_nodes"), SearchIcon, item.disabled, item.onClick)
        case "edit":
          return drawItem(
            item.type,
            t("edit"),
            EditIcon,
            rowsSelectionCount !== 1 || item.disabled,
            item.onClick
          )
        case "delete":
          return drawItem(item.type, t("delete"), OutlinedTrashAltIcon, item.disabled, item.onClick)
        case "reboot":
          return drawItem(item.type, t("reboot"), RebootingIcon, item.disabled, item.onClick)
        case "heartbeat_request":
          return drawItem(item.type, t("request_heartbeat"), HeartbeatIcon, item.disabled, item.onClick)
        case "refresh_node_info":
          return drawItem(item.type, t("fetch_info"), InfoAltIcon, item.disabled, item.onClick)
        case "firmware_update":
          return drawItem(item.type, t("update_firmware"), UploadIcon, item.disabled, item.onClick)
        case "reset":
          return drawItem(item.type, t("reset"), EraserIcon, item.disabled, item.onClick)
        case "restore":
          return drawItem(
            item.type,
            t("restore"),
            FileImportIcon,
            rowsSelectionCount !== 1 || item.disabled,
            item.onClick
          )
        case "separator":
          return <DropdownSeparator key={"separator-" + index} />
        default:
          return <DropdownItem key={"el-" + index}>{item}</DropdownItem>
      }
    })
    return (
      <>
        <Dropdown
          className="mc-actions"
          onSelect={this.onSelect}
          toggle={
            <DropdownToggle
              isPlain={true}
              id="toggle-id"
              isDisabled={isDisabled || rowsSelectionCount === 0}
              onToggle={this.onToggle}
              toggleIndicator={CaretDownIcon}
            >
              {t("actions")}
            </DropdownToggle>
          }
          isOpen={isOpen}
          dropdownItems={dropdownItems}
        />
        <DeleteDialog
          key="deleteDialog"
          dialogTitle={this.props.deleteDialogTitle}
          show={this.state.showDeleteDialog}
          onCloseFn={this.onCloseDeleteDialog}
          onOkFn={this.onConfirmDeleteDialog}
        />
      </>
    )
  }
}

export default withTranslation()(Actions)

const drawItem = (key, text, icon, disabled, onClickFn) => {
  const Icon = icon
  return (
    <DropdownItem key={key} isDisabled={disabled} icon={<Icon />} onClick={onClickFn}>
      {text}
    </DropdownItem>
  )
}
