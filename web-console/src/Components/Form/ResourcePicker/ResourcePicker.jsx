import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { DataType, FieldType } from "../../../Constants/Form"
import {
  BackupProviderType,
  BackupProviderTypeOptions,
  CallerType,
  FieldDataType,
  FieldDataTypeOptions,
  ResourceType,
  ResourceTypeOptions,
  StorageExportTypeOptions,
  TelegramParseModeOptions,
  WebhookMethodType,
  WebhookMethodTypeOptions,
} from "../../../Constants/ResourcePicker"
import { getValue } from "../../../Util/Util"
import { validate } from "../../../Util/Validator"
import Editor from "../../Editor/Editor"
import ErrorBoundary from "../../ErrorBoundary/ErrorBoundary"
import {
  getOptionsDescriptionFunc,
  getResourceFilterFunc,
  getResourceOptionsAPI,
  getResourceOptionValueFunc,
} from "./ResourceUtils"

class ResourcePicker extends React.Component {
  state = {
    isOpen: false,
  }

  onClose = () => {
    this.setState({ isOpen: false })
  }

  onOpen = () => {
    this.setState({ isOpen: true })
  }

  render() {
    const { isOpen } = this.state
    const { value, id, name, onChange, callerType, t } = this.props
    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={`${t("update_value")}: ${name}`}
          variant={ModalVariant.medium}
          position="top"
          isOpen={isOpen}
          onClose={this.onClose}
          onEscapePress={this.onClose}
        >
          <ErrorBoundary>
            <Editor
              key={"editor" + id}
              disableEditor={false}
              language="yaml"
              rootObject={value}
              onSaveFunc={(rootObject) => {
                if (onChange) {
                  onChange(rootObject)
                }
                this.onClose()
              }}
              onChangeFunc={() => {}}
              minimapEnabled={false}
              onCancelFunc={this.onClose}
              isWidthLimited={false}
              getFormItems={(rootObject) => getItems(rootObject, callerType)}
              saveButtonText="update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

ResourcePicker.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
  callerType: PropTypes.string,
}

export default withTranslation()(ResourcePicker)

const getItems = (rootObject, callerType) => {
  const FieldTypes = FieldDataTypeOptions.filter((t) => {
    if (callerType === CallerType.Variable) {
      switch (t.value) {
        // allow only the following types as variable
        case FieldDataType.TypeString:
        case FieldDataType.TypeResourceByLabels:
        case FieldDataType.TypeResourceByQuickId:
        case FieldDataType.TypeWebhook:
          return true
        default:
          return false
      }
    } else {
      // block only string for parameters
      return t.value !== FieldDataType.TypeString
    }
  })

  const items = []

  items.push({
    label: "type",
    fieldId: "type",
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    value: "",
    isRequired: true,
    isDisabled: false,
    helperText: "",
    helperTextInvalid: "helper_text.invalid_type",
    validated: "default",
    options: FieldTypes,
    validator: { isNotEmpty: {} },
    resetFields: {
      resetTypeFunc: (rootObject, newValue) => {
        const disabledValue = getValue(rootObject, "disabled", "")
        // delete all the keys
        Object.keys(rootObject).forEach((key) => delete rootObject[key])
        // keep only type
        rootObject["type"] = newValue
        if (callerType === CallerType.Parameter) {
          rootObject["disabled"] = disabledValue
        }
      },
    },
  })

  if (callerType === CallerType.Parameter) {
    items.push({
      label: "disabled",
      fieldId: "disabled",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      isDisabled: false,
    })
  }

  const dataType = getValue(rootObject, "type", FieldDataType.TypeString)

  switch (dataType) {
    case FieldDataType.TypeResourceByQuickId:
    case FieldDataType.TypeResourceByLabels:
      const resItems = getResourceDataItems(rootObject, dataType, callerType)
      items.push(...resItems)
      break

    case FieldDataType.TypeEmail:
      const emailItems = getEmailDataItems(rootObject)
      items.push(...emailItems)
      break

    case FieldDataType.TypeTelegram:
      const telegramItems = getTelegramDataItems(rootObject)
      items.push(...telegramItems)
      break

    case FieldDataType.TypeBackup:
      const exporterItems = getBackupItems(rootObject)
      items.push(...exporterItems)
      break

    case FieldDataType.TypeWebhook:
      const webhookItems = getWebhookItems(rootObject, callerType)
      items.push(...webhookItems)
      break

    default:
      items.push({
        label: "value",
        fieldId: "value",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
      })
  }

  return items
}

const getResourceDataItems = (rootObject = {}, dataType, callerType) => {
  const items = []
  items.push({
    label: "resource_type",
    fieldId: "resourceType",
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    value: "",
    isRequired: true,
    options: ResourceTypeOptions,
    validator: { isNotEmpty: {} },
  })

  const resourceType = getValue(rootObject, "resourceType", "")

  switch (dataType) {
    case FieldDataType.TypeResourceByQuickId:
      if (resourceType !== "") {
        const resourceAPI = getResourceOptionsAPI(resourceType)
        const resourceOptionValueFunc = getResourceOptionValueFunc(resourceType)
        const resourceFilterFunc = getResourceFilterFunc(resourceType)
        const resourceDescriptionFunc = getOptionsDescriptionFunc(resourceType)
        items.push({
          label: "resource",
          fieldId: "quickId",
          apiOptions: resourceAPI,
          optionValueFunc: resourceOptionValueFunc,
          fieldType: FieldType.SelectTypeAheadAsync,
          getFiltersFunc: resourceFilterFunc,
          getOptionsDescriptionFunc: resourceDescriptionFunc,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          options: [],
          validator: { isNotEmpty: {} },
        })
      }

      break

    case FieldDataType.TypeResourceByLabels:
      items.push({
        label: "labels",
        fieldId: "labels",
        fieldType: FieldType.Labels,
        dataType: DataType.Object,
        value: {},
      })
      break

    default:
    // no change
  }

  if (callerType === CallerType.Variable || resourceType === ResourceType.DataRepository) {
    items.push({
      label: "key_path",
      fieldId: "keyPath",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_key_path",
      validated: "default",
      validator: { isNotEmpty: {} },
    })
  }

  if (callerType === CallerType.Parameter) {
    items.push(
      {
        label: "payload",
        fieldId: "payload",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
      },
      {
        label: "pre_delay",
        fieldId: "preDelay",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
      }
    )
  }

  return items
}

const getEmailDataItems = (_rootObject) => {
  const items = [
    {
      label: "from",
      fieldId: "from",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      // validator: { isEmail: {} },
    },
    {
      label: "to",
      fieldId: "to",
      fieldType: FieldType.DynamicArray,
      dataType: DataType.ArrayString,
      validateValueFunc: (key) => {
        return validate("isEmail", key)
      },
      value: [],
    },
    {
      label: "subject",
      fieldId: "subject",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
    {
      label: "body",
      fieldId: "body",
      fieldType: FieldType.TextArea,
      dataType: DataType.String,
      value: "",
    },
  ]
  return items
}

const getTelegramDataItems = (_rootObject) => {
  const items = [
    {
      label: "chat_ids",
      fieldId: "chatIds",
      fieldType: FieldType.DynamicArray,
      dataType: DataType.ArrayString,
      value: [],
    },
    {
      label: "parse_mode",
      fieldId: "parseMode",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      options: TelegramParseModeOptions,
    },
    {
      label: "text",
      fieldId: "text",
      fieldType: FieldType.TextArea,
      dataType: DataType.String,
      isRequired: true,
      value: "",
      helperText: "",
      helperTextInvalid: "helper_text.invalid_text",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
  ]
  return items
}

const getBackupItems = (rootObject) => {
  const items = []
  items.push({
    label: "provider_type",
    fieldId: "providerType",
    isRequired: true,
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    options: BackupProviderTypeOptions,
    value: "",
    // resetFields: { "spec": {} },
    validator: { isNotEmpty: {} },
  })

  const providerType = getValue(rootObject, "providerType", "")
  switch (providerType) {
    case BackupProviderType.Disk:
      items.push(
        {
          label: "storage_export_type",
          fieldId: "storageExportType",
          fieldType: FieldType.SelectTypeAhead,
          dataType: DataType.String,
          options: StorageExportTypeOptions,
          value: "",
          isRequired: false,
        },
        {
          label: "target_directory",
          fieldId: "targetDirectory",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        },
        {
          label: "prefix",
          fieldId: "prefix",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: false,
        },
        {
          label: "retention_count",
          fieldId: "retentionCount",
          fieldType: FieldType.Text,
          dataType: DataType.Integer,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_backup_retention_count",
          validated: "default",
          validator: { isInteger: {} },
        }
      )
      break

    default:
  }

  return items
}

const getWebhookItems = (rootObject, callerType) => {
  const items = []

  objectPath.set(rootObject, "responseCode", 0, true)

  if (callerType === CallerType.Variable) {
    objectPath.set(rootObject, "method", WebhookMethodType.GET, true)
    items.push(
      {
        label: "url",
        fieldId: "server",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_url",
        validated: "default",
        validator: { isNotEmpty: {}, isURL: {} },
      },
      {
        label: "insecure",
        fieldId: "insecure",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      }
    )
  } else {
    objectPath.set(rootObject, "method", WebhookMethodType.POST, true)
    items.push(
      {
        label: "server",
        fieldId: "server",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: false,
      },
      {
        label: "api",
        fieldId: "api",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: false,
      }
    )
  }

  items.push(
    {
      label: "method",
      fieldId: "method",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: WebhookMethodTypeOptions,
      value: "",
      isRequired: false,
    },
    {
      label: "response_code",
      fieldId: "responseCode",
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
      fieldId: "headers",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.ArrayObject,
      value: {},
      validateKeyFunc: (key) => validate("isNotEmpty", key, {}),
      validateValueFunc: (value) => validate("isNotEmpty", value, {}),
    },
    {
      label: "query_parameters",
      fieldId: "queryParameters",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.Object,
      language: "yaml",
      updateButtonText: "update_query_parameters",
      value: {},
      isRequired: false,
    }
  )

  const methodType = getValue(rootObject, "method", "")

  if (methodType !== WebhookMethodType.GET) {
    items.push({
      label: "custom_data",
      fieldId: "customData",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: "",
      isRequired: false,
    })

    const customData = getValue(rootObject, "customData", false)
    if (customData) {
      items.push({
        label: "data",
        fieldId: "data",
        fieldType: FieldType.ScriptEditor,
        dataType: DataType.Object,
        language: "yaml",
        updateButtonText: "update_data",
        value: {},
        isRequired: false,
      })
    }
  }

  return items
}
