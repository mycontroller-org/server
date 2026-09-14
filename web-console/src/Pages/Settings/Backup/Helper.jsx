import { Modal, ModalVariant } from "@patternfly/react-core"
import React from "react"
import Editor from "../../../Components/Editor/Editor"
import ErrorBoundary from "../../../Components/ErrorBoundary/ErrorBoundary"
import { Operator } from "../../../Constants/Filter"
import { DataType, FieldType } from "../../../Constants/Form"
import { HandlerType } from "../../../Constants/Handler"
import { StorageExportTypeOptions } from "../../../Constants/ResourcePicker"
import { api } from "../../../Service/Api"
import { validate } from "../../../Util/Validator"
import { useTranslation } from "react-i18next"

const CustomModal = ({
  isOpen,
  modalType = "",
  onClose = () => {},
  saveBackupLocationsFunc = () => {},
  backupLocations = [],
}) => {
  const { t } = useTranslation()

  const isBackupLocations = modalType === "backup_locations"
  const modalTitle = isBackupLocations ? "backup_locations" : "run_a_backup"

  const editor = isBackupLocations
    ? getBackupLocationsEditor(backupLocations, onClose, saveBackupLocationsFunc)
    : getRunOnDemandBackupEditor(backupLocations, onClose)

  return (
    <Modal
      key="edit-widget-backup-settings"
      title={t(modalTitle)}
      variant={ModalVariant.medium}
      position="top"
      isOpen={isOpen}
      onClose={onClose}
      onEscapePress={onClose}
    >
      <ErrorBoundary>{editor}</ErrorBoundary>
    </Modal>
  )
}
export default CustomModal

// export locations helper
const getBackupLocationsEditor = (backupLocations = {}, onClose, saveBackupLocationsFunc) => {
  return (
    <Editor
      key="import-export-editor"
      language="yaml"
      rootObject={{ locations: backupLocations }}
      onChangeFunc={() => {}}
      onSaveFunc={saveBackupLocationsFunc}
      minimapEnabled={true}
      onCancelFunc={onClose}
      isWidthLimited={false}
      getFormItems={getBackupLocationsItems}
    />
  )
}

// restore helper
const getRunOnDemandBackupEditor = (backupLocations, onClose) => {
  const locationOptions = getBackupLocations(backupLocations)
  return (
    <Editor
      key="import-export-editor"
      language="yaml"
      saveButtonText="run_backup"
      rootObject={{}}
      onChangeFunc={() => {}}
      onSaveFunc={(rootObject) => {
        api.backup
          .runOnDemandBackup(rootObject)
          .then((_res) => {
            onClose()
          })
          .catch((_e) => {
            onClose()
          })
      }}
      minimapEnabled={true}
      onCancelFunc={onClose}
      isWidthLimited={false}
      getFormItems={(rootObject) => getRunExportItems(rootObject, locationOptions)}
    />
  )
}

const getBackupLocations = (locations = {}) => {
  const names = Object.keys(locations)
  const options = names.map((name) => {
    return {
      value: locations[name],
      label: name,
    }
  })
  return options
}

// run export items
const getRunExportItems = (_rootObject = {}, locationOptions = []) => {
  return [
    {
      label: "prefix",
      fieldId: "prefix",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_prefix",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isKey: {} },
    },
    {
      label: "storage_export_type",
      fieldId: "storageExportType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: StorageExportTypeOptions,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "target_location",
      fieldId: "targetLocation",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: locationOptions,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "handler",
      fieldId: "handler",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.handler.list,
      getFiltersFunc: (value) => {
        return [
          { k: "id", o: Operator.Regex, v: value },
          { k: "type", o: Operator.Equal, v: HandlerType.Backup },
        ]
      },
      getOptionsDescriptionFunc: (item) => {
        return item.description
      },
    },
  ]
}

// export locations items
const getBackupLocationsItems = (_rootObject) => {
  const items = [
    {
      label: "locations",
      fieldId: "!locations",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "locations",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.Object,
      value: {},
      keyLabel: "name",
      valueLabel: "location",
      validateKeyFunc: (value) => {
        return validate("isKey", value)
      },
    },
  ]

  return items
}
