import { Modal, ModalVariant } from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import Editor from "../../Components/Editor/Editor"
import { DataType, FieldType } from "../../Constants/Form"

const EditSettings = ({ showEditSettings, dashboard, onCancel, /*onChange,*/ onSave }) => {
  const { t } = useTranslation()

  return (
    <Modal
      key="edit-settings"
      title={t("edit_dashboard_settings")}
      variant={ModalVariant.medium}
      position="top"
      isOpen={showEditSettings}
      onClose={onCancel}
      onEscapePress={onCancel}
    >
      <Editor
        key="editor"
        language="yaml"
        rootObject={dashboard}
        onChangeFunc={() => {}}
        onSaveFunc={onSave}
        minimapEnabled={false}
        onCancelFunc={onCancel}
        isWidthLimited={false}
        getFormItems={getFormItems}
      />
    </Modal>
  )
}

export default EditSettings

// support functions

const getFormItems = (_rootObject) => {
  const items = [
    {
      label: "title",
      fieldId: "title",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_title",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "description",
      fieldId: "description",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "favorite",
      fieldId: "favorite",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
  ]

  return items
}
