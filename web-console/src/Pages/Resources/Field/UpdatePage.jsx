import React from "react"
import Editor from "../../../Components/Editor/Editor"
import Loading from "../../../Components/Loading/Loading"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { MetricTypeOptions } from "../../../Constants/Metric"
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
    const { cancelFn = () => {} } = this.props

    const isNewEntry = id === undefined || id === ""

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.field.get}
        apiSaveRecord={api.field.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.field.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.resources.field.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_field" />
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
  // do not set id, new id will be updated on the server side
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
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
      isDisabled: !rootObject.nodeId || !rootObject.gatewayId,
      apiOptions: api.source.list,
      optionValueKey: "sourceId",
      getFiltersFunc: (value) => {
        return [
          { k: "gatewayId", o: "eq", v: rootObject.gatewayId },
          { k: "nodeId", o: "eq", v: rootObject.nodeId },
          { k: "sourceId", o: "regex", v: value },
        ]
      },
      getOptionsDescriptionFunc: (item) => {
        return item.name
      },
    },
    {
      label: "field_id",
      fieldId: "fieldId",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_field_id",
      validated: "default",
      validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {}, isID: {} },
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
      label: "unit",
      fieldId: "unit",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
    },
    {
      label: "metric_type",
      fieldId: "metricType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: MetricTypeOptions,
      isRequired: true,
      validator: { isNotEmpty: {} },
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
    {
      label: "value_formatter",
      fieldId: "!value_formatter",
      fieldType: FieldType.Divider,
    },
    {
      label: "on_receive",
      fieldId: "formatter.onReceive",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.String,
      value: "",
      saveButtonText: "update",
      updateButtonText: "update_script",
      language: "javascript",
      minimapEnabled: true,
      isRequired: false,
    },
  ]

  return items
}
