import objectPath from "object-path"
import React from "react"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { Provider, ProviderOptions } from "../../../Constants/VirtualAssistant"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import { v4 as uuidv4 } from "uuid"

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
        apiGetRecord={api.virtualAssistant.get}
        apiSaveRecord={api.virtualAssistant.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.virtualAssistant.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.virtualAssistant.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_virtual_assistant" />
          <PageContent hasNoPaddingTop>{editor} </PageContent>
        </>
      )
    }
    return editor
  }
}

export default UpdatePage

// support functions

const getFormItems = (rootObject, id) => {
  objectPath.set(rootObject, "enabled", true, true)
  objectPath.set(rootObject, "labels", {}, true)
  objectPath.set(rootObject, "config", {}, true)
  objectPath.set(rootObject, "deviceFilter", {}, true)

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
      validator: { isLength: { min: 2, max: 100 }, isID: {}, isNotEmpty: {} },
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
      label: "device_filter",
      fieldId: "!device_filter",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "deviceFilter",
      fieldType: FieldType.Labels,
      dataType: DataType.Object,
      keyLabel: "label",
      value: {},
      validator: { isLabel: {} },
    },
    {
      label: "provider",
      fieldId: "!provider_type",
      fieldType: FieldType.Divider,
    },
    {
      label: "type",
      fieldId: "providerType",
      isRequired: true,
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: ProviderOptions,
      validated: "error",
      value: "",
      resetFields: { config: {} },
    },
  ]

  const providerType = objectPath.get(rootObject, "providerType", "").toLowerCase()

  switch (providerType) {
    case Provider.GoogleAssistant:
      const mySensorsItems = getGoogleAssistantItems(rootObject)
      items.push(...mySensorsItems)
      break

    case Provider.Alexa:
      break

    default:
      break
  }

  return items
}

// get provider items
// get Google Assistant Provider items
const getGoogleAssistantItems = (rootObject) => {
  // set ID, if not set
  const newId = uuidv4().toString()
  objectPath.set(rootObject, "config.agentUserId", newId, true)
  const items = [
    {
      label: "agent_user_id",
      fieldId: "config.agentUserId",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_id",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
  ]
  return items
}
