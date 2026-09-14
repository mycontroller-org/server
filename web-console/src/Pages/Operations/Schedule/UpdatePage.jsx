import objectPath from "object-path"
import React from "react"
import Editor from "../../../Components/Editor/Editor"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { DataType, FieldType } from "../../../Constants/Form"
import { CallerType, WebhookMethodTypeOptions } from "../../../Constants/ResourcePicker"
import {
  CustomVariableType,
  CustomVariableTypeOptions,
  ScheduleFrequency,
  ScheduleFrequencyOptions,
  ScheduleType,
  ScheduleTypeOptions,
  WeekDayOptions,
} from "../../../Constants/Schedule"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import { FieldHandlersList } from "../../../Util/Common"
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
        apiGetRecord={api.schedule.get}
        apiSaveRecord={api.schedule.update}
        minimapEnabled
        onSaveRedirectFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.scheduler.list)
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            r(this.props.history, rMap.operations.scheduler.list)
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, id)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_schedule" />
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
  const isOnDateJob = objectPath.get(rootObject, "spec.frequency", "") === ScheduleFrequency.OnDate

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
      label: "validity",
      fieldId: "!validity",
      fieldType: FieldType.Divider,
    },
    {
      label: "enabled",
      fieldId: "validity.enabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
      isDisabled: isOnDateJob,
    },
  ]

  const validityEnabled = objectPath.get(rootObject, "validity.enabled", false)

  if (validityEnabled) {
    items.push(
      {
        label: "date_yyyy_mm_dd",
        fieldId: "validity.date",
        fieldType: FieldType.DateRangePicker,
        dataType: DataType.Object,
        value: {},
        isRequired: false,
        //helperTextInvalid: "Invalid date",
        validated: "default",
        //validator: { isNotEmpty: {} },
      },
      {
        label: "time_hh_mm_ss",
        fieldId: "validity.time",
        fieldType: FieldType.TimeRangePicker,
        dataType: DataType.Object,
        value: {},
        isRequired: false,
        //helperTextInvalid: "Invalid date",
        validated: "default",
        // validator: { isNotEmpty: {} },
      },
      {
        label: "validate_time_everyday",
        fieldId: "validity.validateTimeEveryday",
        fieldType: FieldType.Switch,
        dataType: DataType.Boolean,
        value: false,
      }
    )
  }

  items.push(
    {
      label: "schedule",
      fieldId: "!schedule",
      fieldType: FieldType.Divider,
    },
    {
      label: "schedule_type",
      fieldId: "type",
      isRequired: true,
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: ScheduleTypeOptions,
      value: "",
      resetFields: { spec: {} },
      validator: { isNotEmpty: {} },
    }
  )

  const scheduleType = objectPath.get(rootObject, "type", "")

  switch (scheduleType) {
    case ScheduleType.Repeat:
      items.push(
        {
          label: "interval_0h0m0s",
          fieldId: "spec.interval",
          fieldType: FieldType.Text,
          dataType: DataType.String,
          value: "",
          isRequired: true,
          validator: { isNotEmpty: {} },
        },
        {
          label: "repeat_count",
          fieldId: "spec.repeatCount",
          fieldType: FieldType.Text,
          dataType: DataType.Integer,
          value: "",
          isRequired: true,
          helperTextInvalid: "helper_text.invalid_repeat_count",
          validated: "default",
          validator: { isInt: {}, isNotEmpty: {} },
        }
      )
      break

    case ScheduleType.Cron:
      items.push({
        label: "cron_expression",
        fieldId: "spec.cronExpression",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        validator: { isNotEmpty: {} },
      })
      break

    case ScheduleType.Simple:
    case ScheduleType.Sunrise:
    case ScheduleType.Sunset:
      updateSimpleJob(rootObject, items)
      break

    default:
  }

  // add notify handlers
  items.push(
    {
      label: "variables",
      fieldId: "!variables",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "variables",
      showUpdateButton: true,
      callerType: CallerType.Variable,
      fieldType: FieldType.VariablesMap,
      keyLabel: "variable_name",
      dataType: DataType.Object,
      value: {},
    }
  )

  // load custom variables
  const customVariableItems = loadCustomVariablesItems(rootObject)
  items.push(...customVariableItems)

  items.push(
    {
      label: "parameters_to_handler",
      fieldId: "!parameters",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "handlerParameters",
      showUpdateButton: true,
      callerType: CallerType.Parameter,
      fieldType: FieldType.VariablesMap,
      keyLabel: "name",
      dataType: DataType.Object,
      value: {},
    },
    {
      label: "handlers",
      fieldId: "!handlers",
      fieldType: FieldType.Divider,
    },
    { ...FieldHandlersList }
  )

  return items
}

const updateOneTimeJobDependencies = (rootObject, _newValue) => {
  const validity = objectPath.get(rootObject, "validity", {})
  const frequency = objectPath.get(rootObject, "spec.frequency", "")
  if (frequency === ScheduleFrequency.OnDate) {
    return {
      enabled: false,
      date: {
        from: "",
        to: "",
      },
      time: {
        from: "",
        to: "",
      },
      validateTimeEveryday: false,
    }
  }
  return validity
}

const updateSimpleJob = (rootObject, items = []) => {
  items.push({
    label: "frequency",
    fieldId: "spec.frequency",
    isRequired: true,
    fieldType: FieldType.SelectTypeAhead,
    dataType: DataType.String,
    options: ScheduleFrequencyOptions,
    value: "",
    resetFields: {
      "spec.dayOfWeek": "",
      "spec.date": "",
      "spec.dayOfMonth": "",
      validity: updateOneTimeJobDependencies,
    },
    validator: { isNotEmpty: {} },
  })

  const frequency = objectPath.get(rootObject, "spec.frequency", "")
  switch (frequency) {
    case ScheduleFrequency.Daily:
      items.push({
        label: "day_of_week",
        fieldId: "spec.dayOfWeek",
        isRequired: true,
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        options: WeekDayOptions,
        isMulti: true,
        value: "",
        validator: { isNotEmpty: {} },
      })
      break

    case ScheduleFrequency.Weekly:
      items.push({
        label: "day_of_week",
        fieldId: "spec.dayOfWeek",
        isRequired: true,
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        options: WeekDayOptions,
        value: "",
        validator: { isNotEmpty: {} },
      })
      break

    case ScheduleFrequency.Monthly:
      items.push({
        label: "day_of_month",
        fieldId: "spec.dateOfMonth",
        fieldType: FieldType.Text,
        dataType: DataType.Integer,
        value: "",
        isRequired: true,
        helperTextInvalid: "helper_text.invalid_day_of_month",
        validated: "default",
        validator: { isNotEmpty: {}, isInt: { gt: 0, lt: 32 } },
      })
      break

    case ScheduleFrequency.OnDate:
      items.push({
        label: "date_yyyy_mm_dd",
        fieldId: "spec.date",
        fieldType: FieldType.DatePicker,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperTextInvalid: "helper_text.invalid_date",
        validated: "default",
        validator: { isNotEmpty: {} },
      })
      break

    default:
  }

  const scheduleType = objectPath.get(rootObject, "type", "")

  if (scheduleType === ScheduleType.Simple) {
    items.push({
      label: "time_hh_mm_ss",
      fieldId: "spec.time",
      fieldType: FieldType.TimePicker,
      dataType: DataType.String,
      value: "",
      placeholder: "hh:mm:ss",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_time",
      validated: "default",
      validator: { isNotEmpty: {} },
    })
  } else if (scheduleType !== "") {
    items.push({
      label: "offset_0h0m0s",
      fieldId: "spec.offset",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperTextInvalid: "helper_text.invalid_offset",
      validated: "default",
      validator: { isNotEmpty: {} },
    })
  }
}

const loadCustomVariablesItems = (rootObject) => {
  let customVariableType = objectPath.get(rootObject, "customVariableType", "")
  if (customVariableType === "" || customVariableType === undefined) {
    objectPath.set(rootObject, "customVariableType", CustomVariableType.None, false)
  }
  customVariableType = objectPath.get(rootObject, "customVariableType", CustomVariableType.None)

  const items = [
    {
      label: "custom_variables",
      fieldId: "!custom_variables",
      fieldType: FieldType.Divider,
    },
    {
      label: "type",
      fieldId: "customVariableType",
      fieldType: FieldType.ToggleButtonGroup,
      dataType: DataType.String,
      options: CustomVariableTypeOptions,
      value: "",
      resetFields: { customVariableConfig: { javascript: "", webhookApi: "" } },
    },
  ]

  switch (customVariableType) {
    case CustomVariableType.None:
      // no action needed
      break

    case CustomVariableType.Javascript:
      items.push({
        label: "script",
        fieldId: "customVariableConfig.javascript",
        fieldType: FieldType.ScriptEditor,
        dataType: DataType.String,
        value: "",
        saveButtonText: "update",
        updateButtonText: "update_script",
        language: "javascript",
        minimapEnabled: true,
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_script_empty",
        validated: "default",
        validator: { isNotEmpty: {} },
      })
      break

    case CustomVariableType.Webhook:
      const webhookItems = getWebhookItems(rootObject)
      items.push(...webhookItems)
      break

    default:
  }

  return items
}

const getWebhookItems = (_rootObject) => {
  const items = [
    {
      label: "url",
      fieldId: "customVariableConfig.webhook.url",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "Invalid URL",
      validated: "default",
      validator: { isNotEmpty: {}, isURL: {} },
    },
    {
      label: "method",
      fieldId: "customVariableConfig.webhook.method",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      options: WebhookMethodTypeOptions,
      value: "",
      isRequired: true,
      validator: { isNotEmpty: {} },
    },
    {
      label: "insecure",
      fieldId: "customVariableConfig.webhook.insecure",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "headers",
      fieldId: "customVariableConfig.webhook.headers",
      fieldType: FieldType.KeyValueMap,
      dataType: DataType.ArrayObject,
      value: {},
      validateKeyFunc: (key) => validate("isNotEmpty", key, {}),
      validateValueFunc: (value) => validate("isNotEmpty", value, {}),
    },
    {
      label: "query_parameters",
      fieldId: "customVariableConfig.webhook.queryParameters",
      fieldType: FieldType.ScriptEditor,
      dataType: DataType.Object,
      language: "yaml",
      updateButtonText: "update_query_parameters",
      value: {},
      isRequired: false,
    },
    {
      label: "include_configuration",
      fieldId: "customVariableConfig.webhook.includeConfig",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
  ]
  return items
}
