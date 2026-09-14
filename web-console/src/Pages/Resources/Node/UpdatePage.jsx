import objectPath from "object-path"
import React from "react"
import { v4 as uuidv4 } from "uuid"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"

class UpdatePage extends React.Component {
  state = {
    loading: true,
  }

  componentDidMount() {
    this.setState({ loading: false })
  }

  render() {
    const { loading } = this.state

    if (loading) {
      return <span>Loading...</span>
    }

    const { id } = this.props.match.params
    const { cancelFn = () => {} } = this.props

    const isNewEntry = id === undefined || id === ""

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.node.get}
        apiSaveRecord={api.node.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.node.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.node.list)
          }
        }}
        getFormItems={getFormItems}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_node" />
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
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_node_id",
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
      label: "assigned_firmware",
      fieldId: "labels.assigned_firmware",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      apiOptions: api.firmware.list,
      optionValueKey: "id",
      getFiltersFunc: (value) => {
        return [{ k: "id", o: "regex", v: value }]
      },
      getOptionsDescriptionFunc: (item) => {
        return item.description
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
      value: "",
    },
  ]

  return items
}
