import {
  Checkbox,
  DatePicker,
  Divider,
  Form as PfForm,
  FormGroup,
  Grid,
  SelectVariant,
  Switch,
  TextArea,
  TextInput,
  TimePicker,
} from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import { DataType, FieldType } from "../../Constants/Form"
import { getRandomId } from "../../Util/Util"
import { validate } from "../../Util/Validator"
import ColorBox from "../Color/ColorBox/ColorBox"
import AsyncSelect from "../Select/AsyncSelect"
import ConditionArrayMapForm from "./ConditionArrayForm"
import DynamicArrayForm from "./DynamicArrayForm"
import DynamicListGeneric from "./DynamicListGeneric"
import "./Form.scss"
import KeyValueMapForm from "./KeyValueMapForm"
import DateRangePicker from "./RangePicker/DateRangePicker"
import TimeRangePicker from "./RangePicker/TimeRangePicker"
import { getResourceDisplayValue, resourceUpdateButtonCallback } from "./ResourcePicker/ResourceUtils"
import ScriptEditor from "./ScriptEditor/ScriptEditor"
import Select from "./Select"
import SimpleSlider from "./Slider/Simple"
import ThresholdsColorForm from "./ThresholdsColorForm"
import ToggleButtonGroup from "./ToggleButtonGroup/ToggleButtonGroup"
import {
  chartYAxisConfigUpdateButtonCallback,
  displayYAxisConfig,
} from "./Widget/ChartsPanel/ChartYAxisConfigMapUtils"
import MixedResourceConfigList from "./Widget/ChartsPanel/MixedResourceConfigList/MixedResourceConfigList"
import MixedControlListForm from "./Widget/MixedControlList"
import PolicyStatementsForm from "./PolicyStatementsForm"

// item sample
// const item = {
//   label: "Title",
//   fieldId: "abc.xyz",
//   fieldType: FieldType.Text,
//   dataType: DataType.String,
//   value: null,
//   fields: [],
//   isRequired: false,
//   helperText: "",
//   helperTextInvalid: "",
//   validated: "", //success, warning, error, default
//   resetFields: {},
// }

export const Form = ({ isHorizontal = false, isWidthLimited = true, items = [], onChange = () => {} }) => {
  const { t } = useTranslation()
  const formGroups = []
  items.forEach((itemRaw, index) => {
    const item = { ...itemRaw }
    // translate label
    if (itemRaw.label !== "" && itemRaw.label !== undefined) {
      item.label = t(itemRaw.label)
    }

    // translate helper text
    if (itemRaw.helperText !== "" && itemRaw.helperText !== undefined) {
      item.helperText = t(itemRaw.helperText)
    }

    // translate invalid helper text
    if (itemRaw.helperTextInvalid !== "" && itemRaw.helperTextInvalid !== undefined) {
      item.helperTextInvalid = t(itemRaw.helperTextInvalid)
    }

    // translate options list if it is not disabled
    if (!itemRaw.OptionsTranslationDisabled) {
      if (Array.isArray(itemRaw.options)) {
        item.options = itemRaw.options.map((opt) => {
          let desc = ""
          if (opt.description !== "" && opt.description !== undefined) {
            desc = t(opt.description)
          }
          return { ...opt, label: t(opt.label), description: desc }
        })
      }
    }

    const onChangeWithItem = (data) => {
      onChange(item, data)
    }
    const field = getField(item, onChangeWithItem)
    const withGroup =
      item.fieldType !== FieldType.Divider
        ? wrapWithFormGroup(item, field, index)
        : wrapWithoutFormGroup(item, field, index)
    formGroups.push(withGroup)
  })

  const form = (
    <PfForm className="mc-form" isHorizontal={isHorizontal} isWidthLimited={false}>
      <Grid>{formGroups}</Grid>
    </PfForm>
  )
  if (isWidthLimited) {
    return (
      <Grid lg={7} md={9} sm={12}>
        {form}
      </Grid>
    )
  }
  return form
}

// helper functions

const getField = (item, onChange) => {
  const itemId = `${item.fieldId}_${getRandomId(5)}`
  switch (item.fieldType) {
    case FieldType.Text:
    case FieldType.Password:
    case FieldType.Email:
      return (
        <TextInput
          id={itemId}
          type={item.fieldType}
          value={item.value}
          onChange={onChange}
          validated={item.validated}
          isDisabled={item.isDisabled}
        />
      )

    case FieldType.TextArea:
      return (
        <TextArea
          id={itemId}
          type={item.fieldType}
          value={item.value}
          onChange={onChange}
          validated={item.validated}
          isDisabled={item.isDisabled}
          resizeOrientation={"vertical"}
          rows={item.rows ? item.rows : 5}
        />
      )

    case FieldType.Checkbox:
      return (
        <Checkbox
          isDisabled={item.isDisabled}
          label={item.fieldLabel}
          isChecked={item.value}
          onChange={onChange}
        ></Checkbox>
      )

    case FieldType.Labels:
      // update validation functions
      if (!item.validateKeyFunc) {
        item.validateKeyFunc = (key) => {
          return validate("isLabelKey", key)
        }
      }
      return (
        <KeyValueMapForm
          keyValueMap={item.value}
          keyLabel={item.keyLabel}
          valueLabel={item.valueLabel}
          actionSpan={item.actionSpan}
          onChange={onChange}
          validateKeyFunc={item.validateKeyFunc}
          isActionDisabled={item.isDisabled}
          isKeyDisabled={item.isDisabled}
          isValueDisabled={item.isDisabled}
        />
      )

    case FieldType.KeyValueMap:
      // update validation functions
      if (!item.validateKeyFunc) {
        item.validateKeyFunc = (key) => {
          return validate("isKey", key)
        }
      }
      return (
        <KeyValueMapForm
          keyValueMap={item.value}
          keyLabel={item.keyLabel}
          valueLabel={item.valueLabel}
          onChange={onChange}
          actionSpan={item.actionSpan}
          showUpdateButton={item.showUpdateButton}
          validateKeyFunc={item.validateKeyFunc}
          validateValueFunc={item.validateValueFunc}
          keyField={item.keyField}
          valueField={item.valueField}
          updateButtonCallback={item.updateButtonCallback}
        />
      )

    case FieldType.VariablesMap:
      // update validation functions
      if (!item.validateKeyFunc) {
        item.validateKeyFunc = (key) => {
          return validate("isVariableKey", key)
        }
      }
      return (
        <KeyValueMapForm
          key={item.fieldId}
          keyValueMap={item.value}
          keyLabel={item.keyLabel}
          valueLabel={item.valueLabel}
          isKeyDisabled={item.isKeyDisabled}
          isValueDisabled={item.isValueDisabled}
          actionSpan={item.actionSpan}
          showUpdateButton={item.showUpdateButton}
          onChange={onChange}
          validateKeyFunc={item.validateKeyFunc}
          valueField={getResourceDisplayValue}
          updateButtonCallback={(cbIndex, cbItem, cbOnChange) =>
            resourceUpdateButtonCallback(item.callerType, cbIndex, cbItem, cbOnChange)
          }
        />
      )

    case FieldType.ChartYAxisConfigMap:
      return (
        <KeyValueMapForm
          key={item.fieldId}
          keyValueMap={item.value}
          keyLabel={"axis"}
          valueLabel={"config"}
          isKeyDisabled={true}
          isValueDisabled={true}
          actionSpan={0}
          isActionDisabled={true}
          showUpdateButton={true}
          onChange={onChange}
          validateKeyFunc={null}
          forceSync={true}
          valueField={displayYAxisConfig}
          updateButtonCallback={chartYAxisConfigUpdateButtonCallback}
        />
      )

    case FieldType.MixedControlList:
      // update validation functions
      if (!item.validateValueFunc) {
        item.validateValueFunc = (key) => {
          return validate("isObject", key)
        }
      }
      return (
        <MixedControlListForm
          key={item.fieldId}
          valuesList={item.value}
          valueLabel={item.valueLabel}
          onChange={onChange}
          validateValueFunc={item.validateValueFunc}
        />
      )

    case FieldType.DynamicListGeneric:
      // update validation functions
      if (!item.validateValueFunc) {
        item.validateValueFunc = (key) => {
          return validate("isObject", key)
        }
      }
      return (
        <DynamicListGeneric
          key={item.fieldId}
          valuesList={item.value}
          valueLabel={item.valueLabel}
          onChange={onChange}
          showUpdateButton={item.showUpdateButton}
          validateKeyFunc={item.validateKeyFunc}
          validateValueFunc={item.validateValueFunc}
          valueField={item.valueField}
          updateButtonCallback={item.updateButtonCallback}
        />
      )

    case FieldType.PolicyStatements:
      return (
        <PolicyStatementsForm
          key={item.fieldId}
          valuesList={item.value || []}
          onChange={onChange}
          isDisabled={!!item.isDisabled}
        />
      )

    case FieldType.ChartMixedResourceConfig:
      // update validation functions
      if (!item.validateValueFunc) {
        item.validateValueFunc = (key) => {
          return validate("isObject", key)
        }
      }
      return (
        <MixedResourceConfigList
          key={item.fieldId}
          valuesList={item.value}
          yAxisConfig={item.yAxisConfig}
          valueLabel={item.valueLabel}
          onChange={onChange}
          validateValueFunc={item.validateValueFunc}
        />
      )

    case FieldType.ConditionsArrayMap:
      // update validation functions
      if (!item.validateKeyFunc) {
        item.validateKeyFunc = (key) => {
          return validate("isKey", key)
        }
      }
      return (
        <ConditionArrayMapForm
          key={item.fieldId}
          conditionsArrayMap={item.value}
          keyLabel={item.keyLabel}
          conditionLabel={item.conditionLabel}
          direction={item.direction}
          valueLabel={item.valueLabel}
          onChange={onChange}
          validateKeyFunc={item.validateKeyFunc}
        />
      )

    case FieldType.DynamicArray:
      // update validation functions
      if (!item.validateValueFunc) {
        item.validateValueFunc = (key) => {
          return validate("isID", key)
        }
      }
      return (
        <DynamicArrayForm
          key={item.fieldId}
          valuesList={item.value}
          valueLabel={item.valueLabel}
          onChange={onChange}
          validateValueFunc={item.validateValueFunc}
          isSelectTypeAheadAsync={item.isSelectTypeAheadAsync}
          selectOptions={item.selectOptions}
        />
      )

    case FieldType.DatePicker:
      return (
        <DatePicker
          key={item.fieldId}
          value={item.value}
          onChange={(_event, value) => onChange(value)}
          isDisabled={item.isDisabled}
          placeholder="yyyy-mm-dd"
        />
      )

    case FieldType.DateRangePicker:
      return (
        <DateRangePicker
          key={item.fieldId}
          value={item.value !== "" ? item.value : {}}
          onChange={onChange}
          validated={item.validated}
          isDisabled={item.isDisabled}
        />
      )

    case FieldType.TimePicker:
      return (
        <TimePicker
          key={item.fieldId}
          time={item.value}
          onChange={(_event, value) => onChange(value)}
          validated={item.validated}
          isDisabled={item.isDisabled}
          placeholder={item.placeholder}
          className="mc-time-range-picker"
          is24Hour
        />
      )

    case FieldType.TimeRangePicker:
      return (
        <TimeRangePicker
          key={item.fieldId}
          value={item.value !== "" ? item.value : {}}
          onChange={onChange}
          validated={item.validated}
          isDisabled={item.isDisabled}
          is24Hour
        />
      )

    case FieldType.ThresholdsColor:
      return <ThresholdsColorForm keyValueMap={item.value} onChange={onChange} />

    case FieldType.ColorBox:
      return (
        <ColorBox
          key={item.fieldId}
          colors={item.colors}
          color={item.value}
          onChange={onChange}
          isDisabled={item.isDisabled}
        />
      )

    case FieldType.Divider:
      return (
        <ul>
          <li>
            <span className="mc-form-header-label">{item.label}</span>
          </li>
          <Divider component="li" />
        </ul>
      )

    case FieldType.SelectTypeAhead:
      return (
        <Select
          variant={item.isMulti ? SelectVariant.typeaheadMulti : SelectVariant.typeahead}
          direction={item.direction}
          options={item.options}
          onChange={onChange}
          selected={item.value}
          isDisabled={item.isDisabled}
          isMulti={item.isMulti}
          isArrayData={item.dataType === DataType.ArrayString}
        />
      )

    case FieldType.SelectTypeAheadAsync:
      return (
        <AsyncSelect
          isDisabled={item.isDisabled}
          apiOptions={item.apiOptions}
          getFiltersFunc={item.getFiltersFunc}
          optionValueKey={item.optionValueKey}
          optionValueFunc={item.optionValueFunc}
          getOptionsDescriptionFunc={item.getOptionsDescriptionFunc}
          onSelectionFunc={onChange}
          selected={item.value}
          isMulti={item.isMulti}
          limit={item.limit}
          direction={item.direction}
          isCreatable={item.isCreatable}
          createText={item.createText}
        />
      )

    case FieldType.Select:
      return (
        <Select
          variant={SelectVariant.single}
          options={item.options}
          onChange={onChange}
          selected={item.value}
          isDisabled={item.isDisabled}
          disableClear={item.disableClear}
          direction={item.direction}
        />
      )

    case FieldType.Switch:
      return (
        <Switch
          id={itemId}
          aria-label={itemId}
          isDisabled={item.isDisabled}
          label={item.labelOn}
          labelOff={item.labelOff}
          isChecked={item.value}
          onChange={onChange}
        />
      )

    case FieldType.ToggleButtonGroup:
      return (
        <ToggleButtonGroup
          options={item.options}
          onSelectionFunc={onChange}
          selected={item.value}
          isDisabled={item.isDisabled}
        />
      )

    case FieldType.ScriptEditor:
      return (
        <ScriptEditor
          name={item.label}
          id={itemId}
          language={item.language}
          minimapEnabled={item.minimapEnabled}
          options={item.options}
          onSaveFunc={onChange}
          saveButtonText={item.saveButtonText}
          updateButtonText={item.updateButtonText}
          value={item.value}
          isDisabled={item.isDisabled}
          enableLanguageOptions={item.enableLanguageOptions}
          onLanguageUpdate={item.onLanguageUpdate}
        />
      )

    case FieldType.SliderSimple:
      return (
        <SimpleSlider
          id={itemId}
          min={item.min}
          max={item.max}
          step={item.step}
          value={item.value}
          onChange={onChange}
          isDisabled={item.isDisabled}
        />
      )

    default:
      return null
  }
}

const wrapWithFormGroup = (item, field, index) => {
  return (
    <FormGroup
      className="mc-form-group"
      hasNoPaddingTop={false}
      isRequired={item.isRequired}
      key={index}
      label={item.label}
      type={item.type}
      fieldId={item.fieldId}
      helperText={item.helperText}
      helperTextInvalid={item.helperTextInvalid}
      validated={item.validated}
      disabled={item.isDisabled}
    >
      {field}
    </FormGroup>
  )
}

const wrapWithoutFormGroup = (_item, field, index) => {
  return (
    <div key={index} className="mc-form-group-header">
      {field}
    </div>
  )
}
