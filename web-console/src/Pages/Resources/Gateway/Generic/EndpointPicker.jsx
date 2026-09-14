import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import Editor from "../../../../Components/Editor/Editor"
import ErrorBoundary from "../../../../Components/ErrorBoundary/ErrorBoundary"
import { DataType, FieldType } from "../../../../Constants/Form"
import { Protocol } from "../../../../Constants/Gateway"
import { WebhookMethodType, WebhookMethodTypeOptions } from "../../../../Constants/ResourcePicker"
import { validate } from "../../../../Util/Validator"
import { callBackEndpointConfigUpdateButtonCallback, getEndpointConfigDisplayValue } from "./Endpoints"
import { withTranslation } from "react-i18next"
import { Language } from "../../../../Constants/CodeEditor"
import { getValue } from "../../../../Util/Util"

class EndpointConfigPicker extends React.Component {
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
    const {
      value,
      id,
      name,
      onChange,
      isNodeEndpoint,
      protocolType = "",
      isPreRun = false,
      title = "update_endpoint",
      t,
    } = this.props
    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={`${t(title)}: ${name}`}
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
                onChange(rootObject)
                this.onClose()
              }}
              onChangeFunc={() => {}}
              minimapEnabled={false}
              onCancelFunc={this.onClose}
              isWidthLimited={false}
              getFormItems={(rootObject) => {
                switch (protocolType) {
                  case Protocol.HTTP:
                    return getHttpItems(rootObject, isNodeEndpoint, isPreRun)

                  case Protocol.MQTT:
                    return getMqttItems(rootObject)

                  default:
                    // NOOP
                    return
                }
              }}
              saveButtonText="Update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

EndpointConfigPicker.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
  isNodeEndpoint: PropTypes.bool,
  protocolType: PropTypes.string,
  isPreRun: PropTypes.bool,
  title: PropTypes.string,
}

export default withTranslation()(EndpointConfigPicker)

const getHttpItems = (rootObject, isNodeEndpoint = false, isPreRun = false) => {
  if (!isNodeEndpoint) {
    objectPath.set(rootObject, "disabled", false, true)
    objectPath.set(rootObject, "includeGlobal", true, true)
  }

  objectPath.set(rootObject, "url", "", true)
  objectPath.set(rootObject, "method", "", true)
  objectPath.set(rootObject, "responseCode", 0, true)
  objectPath.set(rootObject, "headers", {}, true)
  objectPath.set(rootObject, "queryParameters", {}, true)
  objectPath.set(rootObject, "script", "", true)

  const items = []

  if (!isNodeEndpoint && !isPreRun) {
    items.push(
      {
        label: "disabled",
        fieldId: "disabled",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: "",
        isRequired: false,
      },
      {
        label: "include_global_configuration",
        fieldId: "includeGlobalConfiguration",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      }
    )
  }

  if (!isPreRun) {
    objectPath.set(rootObject, "insecure", false, true)
    items.push({
      label: "insecure",
      fieldId: "insecure",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    })
  }

  if (!isNodeEndpoint && !isPreRun) {
    items.push({
      label: "execution_interval",
      fieldId: "executionInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_interval",
      validated: "default",
      validator: { isNotEmpty: {} },
    })
  }

  items.push(
    {
      label: "url",
      fieldId: "url",
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
      label: "method",
      fieldId: "method",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: WebhookMethodTypeOptions,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
    }
  )

  items.push(
    {
      label: "response_code",
      fieldId: "responseCode",
      fieldType: FieldType.Text,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_data",
      validated: "default",
      validator: { isNotEmpty: {}, isInteger: {} },
    },
    {
      label: "headers",
      fieldId: "headers",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.ArrayObject,
      value: {},
      keyLabel: "header_name",
      valueLabel: "value",
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

  const methodType = objectPath.get(rootObject, "method", "get").toUpperCase()
  if (methodType !== "" && methodType !== WebhookMethodType.GET) {
    objectPath.set(rootObject, "body", "", true)
    objectPath.set(rootObject, "bodyLanguage", Language.PLANTEXT, true)

    const bodyLanguage = getValue(rootObject, "bodyLanguage", Language.PLANTEXT)
    items.push({
      label: "body",
      fieldId: "body",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.Object,
      language: bodyLanguage,
      enableLanguageOptions: true,
      onLanguageUpdate: (newLanguage) => objectPath.set(rootObject, "bodyLanguage", newLanguage, false),
      updateButtonText: "update_body",
      value: {},
      isRequired: false,
    })
  }

  items.push({
    label: "script",
    fieldId: "script",
    fieldType: FieldType.ScriptEditor,
    dataType: DataType.String,
    value: "",
    saveButtonText: "update",
    updateButtonText: "update_script",
    language: Language.JAVASCRIPT,
    minimapEnabled: true,
    isRequired: false,
  })

  if (!isPreRun) {
    objectPath.set(rootObject, "preRun", {}, true)
    objectPath.set(rootObject, "postRun", {}, true)

    items.push(
      {
        label: "pre_run",
        fieldId: "!pre_run",
        fieldType: FieldType.Divider,
      },
      {
        label: "",
        fieldId: "preRun",
        fieldType: FieldType.KeyValueMap,
        dataType: DataType.Object,
        value: "",
        keyLabel: "name",
        valueLabel: "endpoint",
        showUpdateButton: true,
        saveButtonText: "update",
        updateButtonText: "update",
        minimapEnabled: true,
        isRequired: false,
        validateKeyFunc: (key) => validate("isVariableKey", key),
        validateValueFunc: (value) => value.url !== undefined || value.url !== "",
        valueField: getEndpointConfigDisplayValue,
        updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
          callBackEndpointConfigUpdateButtonCallback(
            cbIndex,
            cbItem,
            cbOnChange,
            true,
            Protocol.HTTP,
            true,
            "update_pre_run"
          ),
      },
      {
        label: "post_run",
        fieldId: "!post_run",
        fieldType: FieldType.Divider,
      },
      {
        label: "",
        fieldId: "postRun",
        fieldType: FieldType.KeyValueMap,
        dataType: DataType.Object,
        value: "",
        keyLabel: "name",
        valueLabel: "endpoint",
        showUpdateButton: true,
        saveButtonText: "update",
        updateButtonText: "update",
        minimapEnabled: true,
        isRequired: false,
        validateKeyFunc: (key) => validate("isVariableKey", key),
        validateValueFunc: (value) => value.url !== undefined || value.url !== "",
        valueField: getEndpointConfigDisplayValue,
        updateButtonCallback: (cbIndex, cbItem, cbOnChange) =>
          callBackEndpointConfigUpdateButtonCallback(
            cbIndex,
            cbItem,
            cbOnChange,
            true,
            Protocol.HTTP,
            true,
            "update_post_run"
          ),
      }
    )
  }

  return items
}

const getMqttItems = (rootObject) => {
  objectPath.set(rootObject, "topic", "", true)
  objectPath.set(rootObject, "qos", 0, true)
  objectPath.set(rootObject, "script", "", true)

  const items = [
    {
      label: "topic",
      fieldId: "topic",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_topic",
      validator: { isNotEmpty: {} },
    },
    {
      label: "qos",
      fieldId: "qos",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.Integer,
      value: "",
      isRequired: true,
      options: [
        { value: "0", label: "At most once (0)" },
        { value: "1", label: "At least once (1)" },
        { value: "2", label: "Exactly once (2)" },
      ],
    },
    {
      label: "script",
      fieldId: "script",
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
