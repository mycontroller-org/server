import {
  Bullseye,
  Button,
  Grid,
  GridItem,
  Modal,
  ModalVariant,
  Split,
  TextInput,
} from "@patternfly/react-core"
import { AddCircleOIcon, EditIcon, MinusCircleIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { DataType, FieldType } from "../../Constants/Form"
import Editor from "../Editor/Editor"
import ErrorBoundary from "../ErrorBoundary/ErrorBoundary"
import "./Form.scss"

const DEFAULT_STATEMENT = {
  effect: "Allow",
  actions: ["get", "list"],
  resources: [],
}

const EffectOptions = [
  { value: "Allow", label: "Allow" },
  { value: "Deny", label: "Deny" },
]

const ActionOptions = [
  { value: "*", label: "*" },
  { value: "get", label: "get" },
  { value: "list", label: "list" },
  { value: "create", label: "create" },
  { value: "update", label: "update" },
  { value: "delete", label: "delete" },
  { value: "enable", label: "enable" },
  { value: "disable", label: "disable" },
  { value: "reload", label: "reload" },
  { value: "action", label: "action" },
]

const normalizeEffect = (effect) => {
  const v = String(effect || "").trim().toLowerCase()
  if (v === "deny") {
    return "Deny"
  }
  if (v === "allow") {
    return "Allow"
  }
  return ""
}

const normalizeStatement = (value) => {
  const st = value && typeof value === "object" ? { ...value } : { ...DEFAULT_STATEMENT }
  const effect = normalizeEffect(st.effect)
  if (effect) {
    st.effect = effect
  } else if (!st.effect) {
    st.effect = "Allow"
  }
  st.actions = Array.isArray(st.actions) ? [...st.actions] : []
  st.resources = Array.isArray(st.resources) ? [...st.resources] : []
  return st
}

const statementSummary = (st) => {
  const s = normalizeStatement(st)
  const effect = s.effect || "Allow"
  const actions = s.actions.length ? s.actions.join(", ") : "-"
  const resources = s.resources.length ? s.resources.join(", ") : "-"
  return `${effect} | actions: ${actions} | resources: ${resources}`
}

class PolicyStatementsForm extends React.Component {
  state = {
    editIndex: -1,
  }

  notify = (items) => {
    if (this.props.onChange) {
      this.props.onChange(items)
    }
  }

  onAdd = () => {
    const items = [...(this.props.valuesList || [])]
    items.push({ ...DEFAULT_STATEMENT, actions: ["get", "list"], resources: [] })
    this.notify(items)
    this.setState({ editIndex: items.length - 1 })
  }

  onDelete = (index) => {
    const items = [...(this.props.valuesList || [])]
    items.splice(index, 1)
    this.notify(items)
    if (this.state.editIndex === index) {
      this.setState({ editIndex: -1 })
    }
  }

  onSaveStatement = (data) => {
    const { editIndex } = this.state
    if (editIndex < 0) {
      return
    }
    const items = [...(this.props.valuesList || [])]
    items[editIndex] = normalizeStatement(data)
    this.notify(items)
    this.setState({ editIndex: -1 })
  }

  render() {
    const { valuesList = [], t, isDisabled = false } = this.props
    const { editIndex } = this.state
    const items = Array.isArray(valuesList) ? valuesList : []
    const editing =
      !isDisabled && editIndex >= 0 ? normalizeStatement(items[editIndex] || DEFAULT_STATEMENT) : null

    const rows = items.map((item, index) => {
      const summary = statementSummary(item)
      const valid =
        Array.isArray(item.actions) &&
        item.actions.length > 0 &&
        Array.isArray(item.resources) &&
        item.resources.length > 0
      return (
        <React.Fragment key={"stmt-row-" + index}>
          <GridItem span={11}>
            <Split>
              <TextInput
                id={"stmt_summary_" + index}
                value={summary}
                isDisabled
                validated={valid ? "default" : "error"}
                title={summary}
              />
              {isDisabled ? null : (
                <Button
                  variant="control"
                  onClick={() => this.setState({ editIndex: index })}
                  aria-label={t("edit_statement")}
                >
                  <EditIcon />
                </Button>
              )}
            </Split>
          </GridItem>
          <GridItem span={1}>
            {isDisabled ? null : (
              <Bullseye className="btn-layout">
                <Split hasGutter className="btn-split">
                  <MinusCircleIcon
                    onClick={() => this.onDelete(index)}
                    className="btn-remove icon-btn"
                  />
                  {index === items.length - 1 ? (
                    <AddCircleOIcon onClick={this.onAdd} className="btn-add icon-btn" />
                  ) : null}
                </Split>
              </Bullseye>
            )}
          </GridItem>
        </React.Fragment>
      )
    })

    return (
      <Grid className="mc-key-value-map-items">
        {rows}
        {items.length === 0 && !isDisabled ? (
          <GridItem span={12}>
            <Button variant="secondary" isBlock onClick={this.onAdd}>
              {t("add_an_item")}
            </Button>
          </GridItem>
        ) : null}
        {editing ? (
          <Modal
            title={t("edit_statement")}
            variant={ModalVariant.medium}
            position="top"
            isOpen={true}
            onClose={() => this.setState({ editIndex: -1 })}
            onEscapePress={() => this.setState({ editIndex: -1 })}
          >
            <ErrorBoundary>
              <Editor
                key={"stmt-editor-" + editIndex}
                language="yaml"
                rootObject={editing}
                disableEditor={false}
                minimapEnabled={false}
                isWidthLimited={false}
                saveButtonText="update"
                onCancelFunc={() => this.setState({ editIndex: -1 })}
                onChangeFunc={() => {}}
                onSaveFunc={this.onSaveStatement}
                getFormItems={getStatementFormItems}
              />
            </ErrorBoundary>
          </Modal>
        ) : null}
      </Grid>
    )
  }
}

PolicyStatementsForm.propTypes = {
  valuesList: PropTypes.array,
  onChange: PropTypes.func,
  isDisabled: PropTypes.bool,
}

export default withTranslation()(PolicyStatementsForm)

const getStatementFormItems = (rootObject = {}) => {
  const effect = normalizeEffect(rootObject.effect)
  if (effect) {
    objectPath.set(rootObject, "effect", effect, false)
  } else if (!rootObject.effect) {
    objectPath.set(rootObject, "effect", "Allow", false)
  }
  if (!Array.isArray(rootObject.actions)) {
    objectPath.set(rootObject, "actions", ["get", "list"], false)
  }
  if (!Array.isArray(rootObject.resources)) {
    objectPath.set(rootObject, "resources", [], false)
  }

  return [
    {
      label: "effect",
      fieldId: "effect",
      fieldType: FieldType.Select,
      dataType: DataType.String,
      options: EffectOptions,
      OptionsTranslationDisabled: true,
      value: "Allow",
      isRequired: true,
      disableClear: true,
    },
    {
      label: "actions",
      fieldId: "actions",
      fieldType: FieldType.SelectTypeAhead,
      dataType: DataType.ArrayString,
      isMulti: true,
      OptionsTranslationDisabled: true,
      options: ActionOptions,
      value: [],
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_value",
      validator: { isLengthArray: { min: 1 } },
    },
    {
      label: "resources",
      fieldId: "resources",
      fieldType: FieldType.DynamicArray,
      dataType: DataType.ArrayString,
      value: [],
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_value",
      validateValueFunc: (v) => typeof v === "string" && String(v).trim().length > 0,
    },
  ]
}
