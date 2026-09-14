import { Button, Modal, ModalVariant, Split, SplitItem, Stack, StackItem } from "@patternfly/react-core"
import { EditIcon } from "@patternfly/react-icons"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { Language, LanguageOptions } from "../../../Constants/CodeEditor"
import { toObject, toString } from "../../../Util/Language"
import ActionBar from "../../ActionBar/ActionBar"
import CodeEditorBasic from "../../CodeEditor/CodeEditorBasic"
import ErrorBoundary from "../../ErrorBoundary/ErrorBoundary"
import Loading from "../../Loading/Loading"
import Select from "../../Select/Select"

class ScriptEditor extends React.Component {
  state = {
    isOpen: false,
    language: "",
    loading: true,
  }
  editorRef = React.createRef()

  componentDidMount() {
    const { language } = this.props
    this.setState({ language: language, loading: false })
  }

  componentDidUpdate(prevProp) {
    if (prevProp.language !== this.props.language) {
      this.setState({ language: this.props.language, loading: false })
    }
  }

  onLanguageChange = (newSelection) => {
    const { onLanguageUpdate = () => {} } = this.props
    if (onLanguageUpdate) {
      onLanguageUpdate(newSelection)
    }
    this.setState({ language: newSelection })
  }

  handleEditorOnMount = (editor, _monaco) => {
    this.editorRef.current = editor
  }

  onSaveClick = () => {
    const { onSaveFunc, language } = this.props
    if (onSaveFunc) {
      const codeString = this.editorRef.current.getValue()
      const formattedData = toObject(language, codeString)
      onSaveFunc(formattedData)
    }
    this.onClose()
  }

  onClose = () => {
    this.setState({ isOpen: false })
  }

  onOpen = () => {
    this.setState({ isOpen: true })
  }

  render() {
    const { isOpen, language = Language.JAVASCRIPT, loading } = this.state

    if (loading) {
      return <Loading key="loading_script_editor" />
    }

    const {
      id,
      name,
      value,
      saveButtonText,
      updateButtonText,
      options,
      minimapEnabled,
      readOnly,
      t,
      enableLanguageOptions = false,
    } = this.props

    const otherOptions = options ? options : {}

    const basicOptions = {
      readOnly: readOnly,
      minimap: { enabled: minimapEnabled },
    }

    const finalOptions = {
      ...basicOptions,
      ...otherOptions,
    }

    const content = []

    const valueString = toString(language, value)

    content.push(
      <CodeEditorBasic
        key={"code-editor"}
        // height={this.props.height}
        language={language}
        data={valueString}
        options={finalOptions}
        handleEditorOnMount={this.handleEditorOnMount}
      />
    )

    const errorMessage = ""

    const saveText = t(saveButtonText ? saveButtonText : "save")
    const updateBtnText = t(updateButtonText ? updateButtonText : "update_script")

    const actionButtons = [
      { text: saveText, variant: "primary", onClickFunc: this.onSaveClick, isDisabled: false },
    ]
    actionButtons.push({
      text: t("cancel"),
      variant: "secondary",
      onClickFunc: this.onClose,
      isDisabled: false,
    })

    let languageOptions = (
      <Split hasGutter>
        <SplitItem>{t("language")}</SplitItem>
        <SplitItem>
          <Select
            title=""
            options={LanguageOptions}
            defaultValue={language}
            onSelectionFunc={this.onLanguageChange}
            disabled={!enableLanguageOptions}
          />
        </SplitItem>
      </Split>
    )

    return (
      <ErrorBoundary>
        <Button key={"edit-btn-" + id} variant="secondary" isBlock onClick={this.onOpen}>
          {updateBtnText} &nbsp;
          <EditIcon />
        </Button>
        <Modal
          key={"edit-script-data" + id}
          title={name}
          variant={ModalVariant.large}
          position="top"
          isOpen={isOpen}
          onClose={this.onClose}
          onEscapePress={this.onClose}
        >
          <Stack hasGutter>
            <StackItem>{languageOptions}</StackItem>
            <StackItem isFilled>{content}</StackItem>
            <StackItem>{errorMessage}</StackItem>
            <StackItem>
              <ActionBar leftBar={actionButtons} />
            </StackItem>
          </Stack>
        </Modal>
      </ErrorBoundary>
    )
  }
}

ScriptEditor.propTypes = {
  id: PropTypes.string,
  name: PropTypes.string,
  value: PropTypes.string,
  onSaveFunc: PropTypes.func,
  saveButtonText: PropTypes.string,
  updateButtonText: PropTypes.string,
  language: PropTypes.string,
  minimapEnabled: PropTypes.bool,
  options: PropTypes.object,
  enableLanguageOptions: PropTypes.bool,
  onLanguageUpdate: PropTypes.func,
}

export default withTranslation()(ScriptEditor)
