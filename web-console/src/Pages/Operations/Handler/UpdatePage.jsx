import objectPath from "object-path"
import React from "react"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { HandlerType, HandlerTypeOptions } from "../../../Constants/Handler"
import {
  BackupProviderType,
  BackupProviderTypeOptions,
  StorageExportTypeOptions,
  WebhookMethodTypeOptions,
} from "../../../Constants/ResourcePicker"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import { validate } from "../../../Util/Validator"

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
        apiGetRecord={api.handler.get}
        apiSaveRecord={api.handler.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.handler.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.handler.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_handler" />
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
  objectPath.set(rootObject, "labels", { location: "server" }, true)

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
    },
    {
      label: "",
      fieldId: "!labels_end",
      fieldType: FieldType.Divider,
    },
    {
      label: "handler_type",
      fieldId: "type",
      isRequired: true,
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: HandlerTypeOptions,
      value: "",
      resetFields: { spec: {} },
      validator: { isNotEmpty: {} },
    },
  ]

  const handlerType = objectPath.get(rootObject, "type", "")

  switch (handlerType) {
    case HandlerType.Email:
      items.push(
        {
          label: "host",
          fieldId: "spec.host",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_host",
          validated: "default",
          validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
        },
        {
          label: "port",
          fieldId: "spec.port",
          fieldType: FieldType.Text,
          dataType: DataType.Integer,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_port",
          validated: "default",
          validator: { isPort: {}, isNotEmpty: {} },
        },
        {
          label: "insecure",
          fieldId: "spec.insecure",
          fieldType: FieldType.Switch,
          dataType: DataType.Boolean,
          value: false,
        },
        {
          label: "username",
          fieldId: "spec.username",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        },
        {
          label: "password",
          fieldId: "spec.password",
          fieldType: FieldType.Password,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        },
        {
          label: "from_email",
          fieldId: "spec.fromEmail",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        },
        {
          label: "to_emails",
          fieldId: "spec.toEmails",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        }
      )
      break

    case HandlerType.Telegram:
      items.push(
        {
          label: "token",
          fieldId: "spec.token",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_token",
          validated: "default",
          validator: { isNotEmpty: {} },
        },
        {
          label: "chat_ids",
          fieldId: "spec.chatIds",
          fieldType: FieldType.DynamicArray,
          dataType: DataType.ArrayString,
          value: [],
        }
      )
      break

    case HandlerType.Webhook:
      const webhookItems = getWebhookItems(rootObject)
      items.push(...webhookItems)
      break

    case HandlerType.Backup:
      const exporterItems = getBackupItems(rootObject)
      items.push(...exporterItems)
      break

    default:
  }

  return items
}

const getWebhookItems = (rootObject) => {
  objectPath.set(rootObject, "spec.allowOverride", true, true)
  objectPath.set(rootObject, "spec.responseCode", 0, true)

  const items = [
    {
      label: "server",
      fieldId: "spec.server",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_url",
      validated: "default",
      validator: { isNotEmpty: {}, isURL: {} },
    },
    {
      label: "api",
      fieldId: "spec.api",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_api",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
    {
      label: "method",
      fieldId: "spec.method",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: WebhookMethodTypeOptions,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "insecure",
      fieldId: "spec.insecure",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
    {
      label: "response_code",
      fieldId: "spec.responseCode",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_response_code",
      validated: "default",
      validator: { isInteger: {} },
    },
    {
      label: "headers",
      fieldId: "spec.headers",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.ArrayObject,
      value: {},
      validateKeyFunc: (key) => validate("isNotEmpty", key, {}),
      validateValueFunc: (value) => validate("isNotEmpty", value, {}),
    },
    {
      label: "query_parameters",
      fieldId: "spec.queryParameters",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.Object,
      language: "yaml",
      updateButtonText: "update_query_parameters",
      value: {},
      isRequired: false,
    },
    {
      label: "allow_overwrite",
      fieldId: "spec.allowOverwrite",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
    },
  ]

  return items
}

const getBackupItems = (rootObject) => {
  const items = []
  items.push({
    label: "provider_type",
    fieldId: "spec.providerType",
    isRequired: true,
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    options: BackupProviderTypeOptions,
    value: "",
    // resetFields: { "spec": {} },
    validator: { isNotEmpty: {} },
  })

  const providerType = objectPath.get(rootObject, "spec.providerType", "")

  switch (providerType) {
    case BackupProviderType.Disk:
      items.push(
        {
          label: "export_type",
          fieldId: "spec.storageExportType",
          fieldType: FieldType.SelectTypeAhead,
          dataType: DataType.String,
          options: StorageExportTypeOptions,
          value: "",
          isRequired: true,
          validator: { isNotEmpty: {} },
        },
        {
          label: "target_directory",
          fieldId: "spec.targetDirectory",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_directory",
          validated: "default",
          validator: { isNotEmpty: {} },
        },
        {
          label: "prefix",
          fieldId: "spec.prefix",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_prefix",
          validated: "default",
          validator: { isNotEmpty: {}, isKey: {} },
        },
        {
          label: "retention_count",
          fieldId: "spec.retentionCount",
          fieldType: FieldType.Text,
          dataType: DataType.Integer,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_backup_retention_count",
          validated: "default",
          validator: { isInteger: {} },
        },
        {
          label: "include_secure_share",
          fieldId: "spec.includeSecureShare",
          fieldType: FieldType.Switch,
          dataType: DataType.Boolean,
          value: false,
        },
        {
          label: "include_insecure_share",
          fieldId: "spec.includeInsecureShare",
          fieldType: FieldType.Switch,
          dataType: DataType.Boolean,
          value: false,
        }
      )
      break

    default:
  }

  return items
}
