import objectPath from "object-path"
import React from "react"
import { withTranslation } from "react-i18next"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DropDownPositionType } from "../../../Constants/Common"
import { DataType, FieldType } from "../../../Constants/Form"
import { Operator } from "../../../Constants/Filter"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"

class UpdatePage extends React.Component {
  render() {
    const { id } = this.props.match.params
    const { cancelFn = () => {} } = this.props

    const isNewEntry = id === undefined || id === ""

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.user.get}
        apiSaveRecord={api.user.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.settings.user.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.settings.user.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_user" />
          <PageContent hasNoPaddingTop>{editor}</PageContent>
        </>
      )
    }
    return editor
  }
}

export default withTranslation()(UpdatePage)

const getFormItems = (rootObject, id) => {
  objectPath.set(rootObject, "id", "", true)
  objectPath.set(rootObject, "username", "", true)
  objectPath.set(rootObject, "email", "", true)
  objectPath.set(rootObject, "fullName", "", true)
  objectPath.set(rootObject, "disabled", false, true)
  objectPath.set(rootObject, "policies", [], true)
  objectPath.set(rootObject, "password", "", true)
  objectPath.set(rootObject, "labels", {}, true)

  const isNew = !id
  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      isDisabled: true,
    },
    {
      label: "username",
      fieldId: "username",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_name",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "full_name",
      fieldId: "fullName",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_name",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "email",
      fieldId: "email",
      fieldType: FieldType.Email,
      dataType: DataType.String,
      value: "",
      helperTextInvalid: "helper_text.invalid_email",
      validated: "default",
      validator: { isEmail: {} },
    },
    {
      label: "password",
      fieldId: "password",
      fieldType: FieldType.Password,
      dataType: DataType.String,
      value: "",
      isRequired: isNew,
      helperText: isNew ? "" : "helper_text.password_leave_empty",
      helperTextInvalid: "helper_text.invalid_password",
      validated: "default",
      validator: isNew ? { isLength: { min: 4, max: 100 }, isNotEmpty: {} } : {},
    },
    {
      label: "disabled",
      fieldId: "disabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "policies",
      fieldId: "policies",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.ArrayString,
      value: [],
      isMulti: true,
      limit: 20,
      apiOptions: api.policy.list,
      optionValueKey: "id",
      getFiltersFunc: (value) => {
        return [{ k: "id", o: Operator.Regex, v: value }]
      },
      optionValueFunc: (item) => {
        return item.id
      },
      getOptionsDescriptionFunc: (item) => {
        return item.description || item.id
      },
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
      validated: "default",
      validator: { isLabel: {} },
    },
  ]

  return items
}
