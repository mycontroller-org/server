import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { DataType, FieldType } from "../../../../../Constants/Form"
import { InterpolationType, InterpolationTypeLineOptions } from "../../../../../Constants/Metric"
import { ResourceType, ResourceTypeOptions } from "../../../../../Constants/ResourcePicker"
import { ChartType, ChartTypeOptions } from "../../../../../Constants/Widgets/ChartsPanel"
import { ColorsSetBig } from "../../../../../Constants/Widgets/Color"
import { getValue } from "../../../../../Util/Util"
import Editor from "../../../../Editor/Editor"
import ErrorBoundary from "../../../../ErrorBoundary/ErrorBoundary"
import {
  getOptionsDescriptionFunc,
  getResourceFilterFunc,
  getResourceOptionsAPI,
  getResourceOptionValueFunc,
} from "../../../ResourcePicker/ResourceUtils"

class MixedResourceConfigPicker extends React.Component {
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
    const { value, id, onChange, yAxisConfig = {} } = this.props

    const yAxisOptions = Object.keys(yAxisConfig)
      .sort()
      .map((key) => {
        return { value: key, label: `Y Axis ${key}` }
      })

    return (
      <>
        <Button key={"edit-btn-" + id} variant="control" onClick={this.onOpen}>
          <EditIcon />
        </Button>
        <Modal
          key={"edit-field-data" + id}
          title={"Update Resource"}
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
              getFormItems={(rootObject) => getItems(rootObject, yAxisOptions)}
              saveButtonText="update"
            />
          </ErrorBoundary>
        </Modal>
      </>
    )
  }
}

MixedResourceConfigPicker.propTypes = {
  value: PropTypes.string,
  id: PropTypes.string,
  name: PropTypes.string,
  onChange: PropTypes.func,
  yAxisConfig: PropTypes.object,
}

export default MixedResourceConfigPicker

const getItems = (rootObject, yAxisOptions = []) => {
  // update default values
  objectPath.set(rootObject, "disabled", false, true)
  objectPath.set(rootObject, "resourceType", ResourceType.Field, true)
  objectPath.set(rootObject, "quickId", "", true)
  objectPath.set(rootObject, "nameKey", "name", true)
  objectPath.set(rootObject, "chartType", ChartType.AreaChart, true)
  objectPath.set(rootObject, "yAxis", 0, true)
  objectPath.set(rootObject, "color", "#ff8533", true)
  objectPath.set(rootObject, "unit", "", true)
  objectPath.set(rootObject, "useGlobalStyle", true, true)
  objectPath.set(rootObject, "style.fillOpacity", 0.2, true)
  objectPath.set(rootObject, "style.interpolation", InterpolationType.Natural, true)
  objectPath.set(rootObject, "style.roundDecimal", 0, true)
  objectPath.set(rootObject, "style.strokeWidth", 1, true)

  const items = []
  items.push(
    {
      label: "disabled",
      fieldId: "disabled",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    },
    {
      label: "resource_type",
      fieldId: "resourceType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      options: ResourceTypeOptions.filter(
        (opt) => opt.value === ResourceType.Field || opt.value === ResourceType.Node
      ),
      validator: { isNotEmpty: {} },
    }
  )

  const resourceType = getValue(rootObject, "resourceType", "")

  if (resourceType !== "") {
    const resourceAPI = getResourceOptionsAPI(resourceType)
    const resourceOptionValueFunc = getResourceOptionValueFunc(resourceType)
    const resourceFilterFunc = getResourceFilterFunc(resourceType)
    const resourceDescriptionFunc = getOptionsDescriptionFunc(resourceType)
    items.push(
      {
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
      },
      {
        label: "name_key",
        fieldId: "nameKey",
        fieldType: FieldType.Text,
        dataType: DataType.String,
        value: "",
        isRequired: true,
        helperText: "",
        helperTextInvalid: "helper_text.invalid_key",
        validated: "default",
        validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
      }
    )
  }

  items.push(
    {
      label: "chart_type",
      fieldId: "chartType",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      options: ChartTypeOptions.filter((opt) => opt.value !== ChartType.PieChart),
      validator: { isNotEmpty: {} },
    },
    {
      label: "y_axis",
      fieldId: "yAxis",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      options: yAxisOptions,
      validator: { isNotEmpty: {} },
    },
    {
      label: "color",
      fieldId: "color",
      fieldType: FieldType.ColorBox,
      dataType: DataType.String,
      colors: ColorsSetBig,
      isRequired: true,
      value: "",
      validator: { isNotEmpty: {} },
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
      label: "use_global_style",
      fieldId: "useGlobalStyle",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
    }
  )

  const useGlobalStyle = getValue(rootObject, "useGlobalStyle", false)

  if (!useGlobalStyle) {
    items.push(
      {
        label: "style",
        fieldId: "!style_config",
        fieldType: FieldType.Divider,
      },
      {
        label: "fill_opacity_%",
        fieldId: "style.fillOpacity",
        fieldType: FieldType.SliderSimple,
        dataType: DataType.Integer,
        value: "",
        min: 1,
        max: 100,
        step: 1,
      },
      {
        label: "stroke_width_px",
        fieldId: "style.strokeWidth",
        fieldType: FieldType.SliderSimple,
        dataType: DataType.Float,
        value: "",
        min: 0,
        max: 20,
        step: 0.5,
      },
      {
        label: "round_decimal",
        fieldId: "style.roundDecimal",
        fieldType: FieldType.SliderSimple,
        dataType: DataType.Integer,
        value: "",
        min: 0,
        max: 10,
        step: 1,
      },
      {
        label: "interpolation",
        fieldId: "style.interpolation",
        fieldType: FieldType.SelectTypeAhead,
        dataType: DataType.String,
        value: "",
        options: InterpolationTypeLineOptions,
        isRequired: true,
        validator: { isNotEmpty: {} },
      }
    )
  }

  return items
}
