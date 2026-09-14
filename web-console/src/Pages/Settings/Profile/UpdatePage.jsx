import React from "react"
import Editor from "../../../Components/Editor/Editor"
import { DataType, FieldType } from "../../../Constants/Form"
import { api } from "../../../Service/Api"
import { getValue } from "../../../Util/Util"

class UpdatePage extends React.Component {
  render() {
    const { id } = this.props.match.params
    const { cancelFn = () => {} } = this.props

    const editor = (
      <Editor
        key="editor"
        resourceId={id ? id : "none"}
        language="yaml"
        apiGetRecord={api.auth.profile}
        apiSaveRecord={api.auth.updateProfile}
        minimapEnabled
        onSaveRedirectFunc={() => {
          cancelFn()
        }}
        onCancelFunc={() => {
          cancelFn()
        }}
        getFormItems={getFormItems}
      />
    )

    return editor
  }
}

export default UpdatePage

// support functions
const getFormItems = (rootObject) => {
  const newPassword = getValue(rootObject, "newPassword", "")
  const confirmPassword = getValue(rootObject, "confirmPassword", "")

  let passwordValidator = null
  let isNewPasswordRequired = false

  if (newPassword !== "" || confirmPassword !== "") {
    isNewPasswordRequired = true
    passwordValidator = { isLength: { min: 3, max: 100 }, isNotEmpty: {} }
  }

  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: true,
    },
    {
      label: "username",
      fieldId: "username",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      isRequired: true,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_profile_username",
      validated: "default",
      validator: { isLength: { min: 3, max: 100 }, isNotEmpty: {}, isID: {} },
    },
    {
      label: "email",
      fieldId: "email",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      isRequired: true,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_email",
      validated: "default",
      validator: { isEmail: {} },
    },
    {
      label: "full_name",
      fieldId: "fullName",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      isRequired: true,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_name",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "current_password",
      fieldId: "currentPassword",
      fieldType: FieldType.Password,
      dataType: DataType.String,
      isRequired: true,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_current_password",
      validated: "default",
    },
    {
      label: "new_password",
      fieldId: "newPassword",
      fieldType: FieldType.Password,
      dataType: DataType.String,
      isRequired: isNewPasswordRequired,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_profile_password",
      validated: "default",
      validator: passwordValidator,
    },
    {
      label: "confirm_password",
      fieldId: "confirmPassword",
      fieldType: FieldType.Password,
      dataType: DataType.String,
      isRequired: isNewPasswordRequired,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_profile_confirm_password",
      validated: "default",
      validator: passwordValidator,
    },
    {
      label: "labels",
      fieldId: "!labels",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "labels",
      fieldType: FieldType.Labels,
      dataType: DataType.Object,
      value: {},
    },
  ]

  return items
}
