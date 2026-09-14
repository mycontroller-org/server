import { Stack, StackItem } from "@patternfly/react-core"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { toObject, toString } from "../../Util/Language"
import ActionBar from "../ActionBar/ActionBar"
import Loading from "../Loading/Loading"
import CodeEditorBasic from "./CodeEditorBasic"

class CodeEditor extends React.Component {
  state = {
    loading: true,
    dataOriginal: {},
  }
  editorRef = React.createRef()

  componentDidMount() {
    this.onReloadClick()
  }

  onReloadClick = () => {
    if (this.props.resourceId) {
      this.props
        .apiGetRecord(this.props.resourceId)
        .then((res) => {
          this.setState({ dataOriginal: res.data, loading: false }, () => {
            if (this.editorRef.current) {
              const codeString = toString(this.props.language, res.data)
              this.editorRef.current.setValue(codeString)
            }
          })
        })
        .catch((_e) => {
          this.setState({ loading: false })
        })
    } else {
      this.setState({ loading: false, dataOriginal: "" })
    }
  }

  onSaveClick = () => {
    if (this.props.apiSaveRecord) {
      const codeString = this.editorRef.current.getValue()
      const data = toObject(this.props.language, codeString)
      this.props
        .apiSaveRecord(data)
        .then((_res) => {
          if (this.props.onSave) {
            this.props.onSave()
          }
        })
        .catch((_e) => {})
    }
  }

  handleOnMount = (editor, _monaco) => {
    this.editorRef.current = editor
  }

  render() {
    const { loading, dataOriginal } = this.state
    const { t } = this.props

    if (loading) {
      return <Loading key="loading" />
    }

    const codeString = toString(this.props.language, dataOriginal)

    const otherOptions = this.props.otherOptions ? this.props.otherOptions : {}

    const basicOptions = {
      readOnly: this.props.readOnly,
      minimap: { enabled: this.props.minimapEnabled },
    }

    const options = {
      ...basicOptions,
      ...otherOptions,
    }

    const actionButtons = [
      { text: t("save"), variant: "primary", onClickFunc: this.onSaveClick, isDisabled: false },
      { text: t("reload"), variant: "secondary", onClickFunc: this.onReloadClick, isDisabled: false },
    ]

    if (this.props.onCancelFunc) {
      actionButtons.push({
        text: t("cancel"),
        variant: "secondary",
        onClickFunc: this.props.onCancelFunc,
        isDisabled: false,
      })
    }

    const rightBar = [{ text: t("download"), variant: "secondary", onClickFunc: () => {}, isDisabled: true }]

    return (
      <Stack hasGutter>
        <StackItem>
          <CodeEditorBasic
            height={this.props.height}
            language={this.props.language}
            data={codeString}
            options={options}
            handleEditorOnMount={this.handleOnMount}
          />
        </StackItem>
        <StackItem>
          <ActionBar leftBar={actionButtons} rightBar={rightBar} />
        </StackItem>
      </Stack>
    )
  }
}

CodeEditor.propTypes = {
  resourceId: PropTypes.string,
  apiGetRecord: PropTypes.func,
  apiSaveRecord: PropTypes.func,
  onCancelFunc: PropTypes.func,
  language: PropTypes.string,
  minimapEnabled: PropTypes.bool,
  readOnly: PropTypes.bool,
  otherOptions: PropTypes.object,
}

export default withTranslation()(CodeEditor)
