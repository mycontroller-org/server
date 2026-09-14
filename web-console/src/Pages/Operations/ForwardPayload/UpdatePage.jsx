import React from "react"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { ResourceType } from "../../../Constants/ResourcePicker"
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
        apiGetRecord={api.forwardPayload.get}
        apiSaveRecord={api.forwardPayload.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.forwardPayload.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.forwardPayload.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_forward_payload" />
          <PageContent hasNoPaddingTop>{editor} </PageContent>
        </>
      )
    }
    return editor
  }
}

export default UpdatePage

// support functions

const getFormItems = (_rootObject, id) => {
  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: id ? true : false,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_id",
      validated: "default",
      validator: { isLength: { min: 4, max: 100 }, isNotEmpty: {}, isID: {} },
    },
    {
      label: "description",
      fieldId: "description",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "enabled",
      fieldId: "enabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
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
      validator: { isLabel: {} },
    },
    {
      label: "mapping",
      fieldId: "!mapping",
      fieldType: FieldType.Divider,
    },
    {
      label: "source_field",
      fieldId: "srcFieldId",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: (value) => {
        return [{ k: "name", o: "regex", v: value }]
      },
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    },
    {
      label: "destination_field",
      fieldId: "dstFieldId",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.field.list,
      getFiltersFunc: (value) => {
        return [{ k: "name", o: "regex", v: value }]
      },
      optionValueFunc: getResourceOptionValueFunc,
      getOptionsDescriptionFunc: getOptionsDescriptionFuncImpl,
    },
  ]

  return items
}

// get display field name
const getOptionsDescriptionFuncImpl = (item) => {
  return item.name
}

const getResourceOptionValueFunc = (item) => {
  return `${ResourceType.Field}:${item.gatewayId}.${item.nodeId}.${item.sourceId}.${item.fieldId}`
}
