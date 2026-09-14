import objectPath from "object-path"
import React from "react"
import { connect } from "react-redux"
import Editor from "../../../Components/Editor/Editor"
import { DEFAULT_LANGUAGE } from "../../../Constants/Common"
import { DataType, FieldType } from "../../../Constants/Form"
import { LanguageOptions } from "../../../i18n/languages"
import { api } from "../../../Service/Api"
import { updateLocale } from "../../../store/entities/locale"
import { getValue } from "../../../Util/Util"

class UpdatePage extends React.Component {
  render() {
    const { id } = this.props.match.params
    const { cancelFn = () => {} } = this.props

    const editor = (
      <Editor
        key="editor"
        resourceId={id ? id : "none"}
        language="yaml"
        apiGetRecord={api.settings.getSystemSettings}
        apiSaveRecord={api.settings.updateSystem}
        minimapEnabled
        onSaveRedirectFunc={(data) => {
          this.props.updateLocale({ language: data.spec.language })
          cancelFn()
        }}
        onCancelFunc={() => {
          cancelFn()
        }}
        getFormItems={getFormItems}
      />
    )

    return editor
  }
}

const mapStateToProps = (state) => ({
  languageSelected: state.entities.locale.language,
})

const mapDispatchToProps = (dispatch) => ({
  updateLocale: (data) => dispatch(updateLocale(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(UpdatePage)

// support functions

const getFormItems = (rootObject) => {
  const autoUpdate = getValue(rootObject, "spec.geoLocation.autoUpdate", false)

  // update values
  if (getValue(rootObject, "spec.language", "") === "" ){
    objectPath.set(rootObject, "spec.language", DEFAULT_LANGUAGE, false)
  }
  if (getValue(rootObject, "spec.nodeStateJob.executionInterval", "") === "" ){
    objectPath.set(rootObject, "spec.nodeStateJob.executionInterval", "15m", false)
  }
  if (getValue(rootObject, "spec.nodeStateJob.inactiveDuration", "") === "" ){
    objectPath.set(rootObject, "spec.nodeStateJob.inactiveDuration", "15m", false)
  }
  
  const items = [
    {
      label: "geo_location",
      fieldId: "!geo_location",
      fieldType: FieldType.Divider,
    },
    {
      label: "auto_update",
      fieldId: "spec.geoLocation.autoUpdate",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "location_name",
      fieldId: "spec.geoLocation.locationName",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      isDisabled: autoUpdate,
    },
    {
      label: "latitude",
      fieldId: "spec.geoLocation.latitude",
      fieldType: FieldType.Text,
      dataType: DataType.Float,
      isRequired: !autoUpdate,
      value: "",
      isDisabled: autoUpdate,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_latitude",
      validated: "default",
      validator: autoUpdate ? {} : { isFloat: {}, isNotEmpty: {} },
    },
    {
      label: "longitude",
      fieldId: "spec.geoLocation.longitude",
      fieldType: FieldType.Text,
      dataType: DataType.Float,
      isRequired: !autoUpdate,
      value: "",
      isDisabled: autoUpdate,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_longitude",
      validated: "default",
      validator: autoUpdate ? {} : { isFloat: {}, isNotEmpty: {} },
    },
    {
      label: "login_page",
      fieldId: "!login_page",
      fieldType: FieldType.Divider,
    },
    {
      label: "message",
      fieldId: "spec.login.message",
      fieldType: FieldType.TextArea,
      dataType: DataType.String,
      isRequired: false,
      value: "",
      // rows: 3,
    },
    {
      label: "others",
      fieldId: "!others_page",
      fieldType: FieldType.Divider,
    },
    {
      label: "language",
      fieldId: "spec.language",
      fieldType: FieldType.Select,
      dataType: DataType.String,
      options: LanguageOptions(),
      OptionsTranslationDisabled: true,
      isDisabled: false,
      value: "",
    },
    {
      label: "node_state_updater",
      fieldId: "!node_state_updater",
      fieldType: FieldType.Divider,
    },
    {
      label: "execution_interval",
      fieldId: "spec.nodeStateJob.executionInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_duration",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
    {
      label: "inactive_duration",
      fieldId: "spec.nodeStateJob.inactiveDuration",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_duration",
      validated: "default",
      validator: { isNotEmpty: {} },
    },
  ]

  return items
}
