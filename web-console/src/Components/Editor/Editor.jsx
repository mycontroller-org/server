import { Alert, Divider, Flex, FlexItem, Grid, Radio, Stack, StackItem } from "@patternfly/react-core"
import { DownloadIcon } from "@patternfly/react-icons"
import objectPath from "object-path"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { v4 as uuidv4 } from "uuid"
import { toObject, toString } from "../../Util/Language"
import { validate, validateItem } from "../../Util/Validator"
import ActionBar from "../ActionBar/ActionBar"
import CodeEditorBasic from "../CodeEditor/CodeEditorBasic"
import { Form } from "../Form/Form"
import { updateItems, updateRootObject } from "../Form/Functions"
import Loading from "../Loading/Loading"
import "./Editor.scss"

class Editor extends React.Component {
  state = {
    loading: true,
    isReloadable: false,
    rootObject: {},
    formView: true,
    inValidItems: [],
  }
  editorRef = React.createRef()

  handleEditorOnMount = (editor, _monaco) => {
    this.editorRef.current = editor
  }

  componentDidMount() {
    this.onReloadClick()
  }

  onReloadClick = () => {
    if (this.props.resourceId) {
      this.props
        .apiGetRecord(this.props.resourceId)
        .then((res) => {
          const rootObject = res.data
          this.updateStateWithRootObject(rootObject, true)
        })
        .catch((_e) => {
          this.setState({ loading: false, isReloadable: true })
        })
    } else if (this.props.rootObject) {
      this.updateStateWithRootObject(this.props.rootObject)
    } else {
      this.setState({ loading: false, rootObject: {}, isReloadable: false })
    }
  }

  updateStateWithRootObject = (rootObject, isReloadable = false) => {
    this.setState(
      (prevState) => {
        // validate fields
        const { inValidItems } = prevState
        const formItems = this.getUpdatedFormItems(rootObject, inValidItems)
        formItems.forEach((item) => {
          updateInValidList(inValidItems, item, item.value)
        })
        return {
          rootObject: rootObject,
          inValidItems: inValidItems,
          loading: false,
          isReloadable: isReloadable,
        }
      },
      () => {
        if (this.editorRef.current) {
          const codeString = toString(this.props.language, rootObject)
          this.editorRef.current.setValue(codeString)
        }
      }
    )
  }

  onViewChange = () => {
    this.setState((prevState) => {
      if (!prevState.formView) {
        const data = toObject(this.props.language, this.editorRef.current.getValue())
        return { formView: !prevState.formView, rootObject: data }
      }
      return { formView: !prevState.formView }
    })
  }

  onFormValueChange = (item, newValue) => {
    this.setState((prevState) => {
      const { rootObject, inValidItems } = prevState
      updateRootObject(rootObject, item, newValue)
      updateInValidList(inValidItems, item, newValue)
      // update reset fields
      if (item.resetFields) {
        const fields = Object.keys(item.resetFields)
        fields.forEach((field) => {
          const input = item.resetFields[field]
          let value = ""
          if (typeof input === "function") {
            value = input(rootObject, newValue)
          } else {
            value = input
          }
          objectPath.set(rootObject, field, value)
        })
      }
      if (this.props.onChangeFunc) {
        this.props.onChangeFunc(rootObject)
      }
      return { rootObject: rootObject, inValidItems: inValidItems }
    })
  }

  isLocked = (rootObject = this.state.rootObject) => {
    if (this.props.readOnly) {
      return true
    }
    if (typeof this.props.readOnlyIf === "function") {
      return !!this.props.readOnlyIf(rootObject)
    }
    return false
  }

  onSaveClick = () => {
    if (this.isLocked()) {
      return
    }
    if (this.props.apiSaveRecord || this.props.onSaveFunc) {
      this.setState((prevState) => {
        const { formView, rootObject, inValidItems } = prevState
        let data = {}
        if (formView) {
          // validate fields
          const formItems = this.getUpdatedFormItems(rootObject, inValidItems)
          formItems.forEach((item) => {
            updateInValidList(inValidItems, item, item.value)
          })
          if (inValidItems.length > 0) {
            return { inValidItems: inValidItems }
          }
          data = rootObject
        } else {
          const codeString = this.editorRef.current.getValue()
          data = toObject(this.props.language, codeString)
        }
        if (this.props.apiSaveRecord) {
          this.props
            .apiSaveRecord(data)
            .then((res) => {
              this.callOnSaveRedirect(data, res.data)
            })
            .catch((_e) => {})
        } else if (this.props.onSaveFunc) {
          this.props.onSaveFunc(data)
          this.callOnSaveRedirect(data)
        }
      })
      return {}
    }
  }

  onDownloadClick = () => {}

  callOnSaveRedirect = (data = {}, saveResponse = {}) => {
    if (this.props.onSaveRedirectFunc) {
      this.props.onSaveRedirectFunc(data, saveResponse)
    }
  }

  getUpdatedFormItems = (rootObject, inValidItems) => {
    if (this.props.getFormItems) {
      const formItems = this.props.getFormItems(rootObject, this.props.t)
      updateItems(rootObject, formItems)
      updateValidations(formItems, inValidItems) // update validations
      return formItems
    }
    return []
  }

  render() {
    const { loading, rootObject, formView, isReloadable, inValidItems } = this.state
    const { saveButtonText, isWidthLimited = true, disableEditor = false, t } = this.props
    const locked = this.isLocked(rootObject)
    if (loading) {
      return <Loading />
    }

    const content = []

    if (formView) {
      // update items with root object
      const formItems = this.getUpdatedFormItems(rootObject, inValidItems)

      content.push(
        <Form
          key={"form-view"}
          isHorizontal
          isWidthLimited={isWidthLimited}
          items={formItems}
          onChange={this.onFormValueChange}
        />
      )
    } else if (!disableEditor) {
      const codeString = toString(this.props.language, rootObject)

      const otherOptions = this.props.otherOptions ? this.props.otherOptions : {}

      const basicOptions = {
        readOnly: locked || this.props.readOnly,
        minimap: { enabled: this.props.minimapEnabled },
      }

      const options = {
        ...basicOptions,
        ...otherOptions,
      }

      content.push(
        <CodeEditorBasic
          key={"code-editor"}
          height={this.props.height}
          language={this.props.language}
          data={codeString}
          options={options}
          handleEditorOnMount={this.handleEditorOnMount}
        />
      )
    }

    let saveDisabled = false

    if (formView && inValidItems.length > 0) {
      saveDisabled = true
    }

    const saveText = t(saveButtonText ? saveButtonText : "save")

    const actionButtons = []
    if (!locked) {
      actionButtons.push({
        text: saveText,
        variant: "primary",
        onClickFunc: this.onSaveClick,
        isDisabled: saveDisabled,
      })
    }

    if (isReloadable) {
      actionButtons.push({
        text: t("reload"),
        variant: "secondary",
        onClickFunc: this.onReloadClick,
        isDisabled: false,
      })
    }

    if (this.props.onCancelFunc) {
      actionButtons.push({
        text: t("cancel"),
        variant: "secondary",
        onClickFunc: this.props.onCancelFunc,
        isDisabled: false,
      })
    }

    const rightBar = formView
      ? []
      : [
          {
            text: t("download"),
            icon: DownloadIcon,
            variant: "secondary",
            onClickFunc: this.onDownloadClick,
            isDisabled: true,
          },
        ]

    let errorMessage = null
    if (formView && inValidItems.length > 0) {
      errorMessage = <Alert variant="danger" isInline title={t("form_generic_error_message")} />
      if (isWidthLimited) {
        errorMessage = (
          <Grid lg={7} md={9} sm={12}>
            {errorMessage}
          </Grid>
        )
      }
    }

    return (
      <div className={formView ? "mc-editor" : ""}>
        <Stack hasGutter>
          <StackItem>
            <Flex>
              <FlexItem>
                <span>
                  <strong>{t("configure_via")}:</strong>
                </span>
              </FlexItem>
              <FlexItem>
                <Radio
                  isChecked={formView}
                  onChange={this.onViewChange}
                  label={t("form_view")}
                  id={"form_view_" + uuidv4()}
                />
              </FlexItem>
              <FlexItem>
                <Radio
                  isChecked={!formView}
                  isDisabled={disableEditor}
                  onChange={disableEditor ? () => {} : this.onViewChange}
                  label={t("yaml_view")}
                  id={"yaml_view_" + uuidv4()}
                />
              </FlexItem>
            </Flex>
            <Divider component="hr" />
          </StackItem>
          <StackItem isFilled>{content}</StackItem>
          <StackItem>{errorMessage}</StackItem>
          <StackItem>
            <ActionBar leftBar={actionButtons} rightBar={rightBar} />
          </StackItem>
        </Stack>
      </div>
    )
  }
}

Editor.propTypes = {
  resourceId: PropTypes.string,
  apiGetRecord: PropTypes.func,
  apiSaveRecord: PropTypes.func,
  rootObject: PropTypes.object,
  onChangeFunc: PropTypes.func,
  onSaveFunc: PropTypes.func,
  onSaveRedirectFunc: PropTypes.func, // on successful save called this redirect will be called
  onCancelFunc: PropTypes.func,
  language: PropTypes.string,
  minimapEnabled: PropTypes.bool,
  readOnly: PropTypes.bool,
  otherOptions: PropTypes.object,
  getFormItems: PropTypes.func,
  isWidthLimited: PropTypes.bool,
  saveButtonText: PropTypes.string,
  readOnlyIf: PropTypes.func,
}

export default withTranslation()(Editor)

// helper functions

const updateValidations = (items, inValidItems) => {
  items.forEach((item) => {
    if (inValidItems.indexOf(item.fieldId) !== -1) {
      item.validated = "error"
    } else {
      item.validated = "default"
    }
  })
}

const updateInValidList = (inValidItems, item, value) => {
  const isValid = validateItem(item, value)
  let makeInvalid = false

  if (isValid) {
    if (item.isRequired) {
      const isEmpty = validate("isEmpty", value, {})
      if (isEmpty) {
        makeInvalid = true
      }
    }
    if (inValidItems.indexOf(item.fieldId) !== -1 && !makeInvalid) {
      inValidItems.splice(inValidItems.indexOf(item.fieldId), 1)
    }
  } else {
    makeInvalid = true
  }

  if (makeInvalid) {
    if (inValidItems.indexOf(item.fieldId) === -1) {
      inValidItems.push(item.fieldId)
    }
  }
}
