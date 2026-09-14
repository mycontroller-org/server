import objectPath from "object-path"
import React from "react"
import { v4 as uuidv4 } from "uuid"
import Editor from "../../../Components/Editor/Editor"
import Loading from "../../../Components/Loading/Loading"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"

class UpdatePage extends React.Component {
  state = {
    loading: false,
  }

  componentDidMount() {}

  render() {
    const { loading } = this.state

    if (loading) {
      return <Loading key="loading" />
    }

    const { id } = this.props.match.params
    const { history, cancelFn = () => {} } = this.props

    const isNewEntry = id === undefined || id === ""

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.source.get}
        apiSaveRecord={api.source.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(history, rMap.resources.source.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(history, rMap.resources.source.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_source" />
          <PageContent hasNoPaddingTop>{editor} </PageContent>
        </>
      )
    }
    return editor
  }
}

export default UpdatePage

// support functions

const getFormItems = (rootObject) => {
  // set ID, if not set
  const newID = uuidv4().toString()
  objectPath.set(rootObject, "id", newID, true)

  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      isDisabled: true,
      helperText: "",
      helperTextInvalid: "",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
    {
      label: "gateway",
      fieldId: "gatewayId",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      apiOptions: api.gateway.list,
      getFiltersFunc: (value) => {
        return [{ k: "id", o: "regex", v: value }]
      },
      getOptionsDescriptionFunc: (item) => {
        return item.description
      },
    },
    {
      label: "node_id",
      fieldId: "nodeId",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      isDisabled: !rootObject.gatewayId,
      apiOptions: api.node.list,
      optionValueKey: "nodeId",
      getFiltersFunc: (value) => {
        return [
          { k: "gatewayId", o: "eq", v: rootObject.gatewayId },
          { k: "nodeId", o: "regex", v: value },
        ]
      },
      getOptionsDescriptionFunc: (item) => {
        return item.name
      },
    },
    {
      label: "source_id",
      fieldId: "sourceId",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_source_id",
      validated: "default",
      validator: { isLength: { min: 1, max: 100 }, isID: {}, isNotEmpty: {} },
    },
    {
      label: "name",
      fieldId: "name",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_name",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
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
      value: "",
    },
  ]

  return items
}
