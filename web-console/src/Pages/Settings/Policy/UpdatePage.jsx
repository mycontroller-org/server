import objectPath from "object-path"
import React from "react"
import { withTranslation } from "react-i18next"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
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
        apiGetRecord={api.policy.get}
        apiSaveRecord={api.policy.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.settings.policy.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.settings.policy.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
        readOnlyIf={(rootObject) => !!rootObject.system}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_policy" />
          <PageContent hasNoPaddingTop>{editor}</PageContent>
        </>
      )
    }
    return editor
  }
}

export default withTranslation()(UpdatePage)

const normalizeStatement = (st) => {
  const s = st && typeof st === "object" ? { ...st } : {}
  const effect = String(s.effect || "").trim().toLowerCase()
  if (effect === "deny") {
    s.effect = "Deny"
  } else if (effect === "allow" || !s.effect) {
    s.effect = "Allow"
  }
  s.actions = Array.isArray(s.actions) ? s.actions : []
  s.resources = Array.isArray(s.resources) ? s.resources : []
  return s
}

const getFormItems = (rootObject, id) => {
  objectPath.set(rootObject, "id", "", true)
  objectPath.set(rootObject, "description", "", true)
  objectPath.set(rootObject, "system", false, true)
  objectPath.set(rootObject, "labels", {}, true)

  if (!Array.isArray(rootObject.statements)) {
    rootObject.statements = []
  } else {
    rootObject.statements = rootObject.statements.map(normalizeStatement)
  }

  const isSystem = objectPath.get(rootObject, "system", false)
  const isNew = !id

  return [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: !isNew || isSystem,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_id",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {}, isID: {} },
    },
    {
      label: "description",
      fieldId: "description",
      fieldType: FieldType.TextArea,
      dataType: DataType.String,
      value: "",
      isDisabled: isSystem,
    },
    {
      label: "statements",
      fieldId: "!statements_divider",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "statements",
      fieldType: FieldType.PolicyStatements,
      dataType: DataType.ArrayObject,
      value: [],
      isRequired: true,
      isDisabled: isSystem,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_value",
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
      isDisabled: isSystem,
      validated: "default",
      validator: { isLabel: {} },
    },
  ]
}
