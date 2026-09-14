import { Alert, ClipboardCopy, Modal, ModalVariant } from "@patternfly/react-core"
import objectPath from "object-path"
import React from "react"
import TokenBarcode from "../../../Components/DataDisplay/TokenBarcode"
import Editor from "../../../Components/Editor/Editor"
import Loading from "../../../Components/Loading/Loading"
import PageContent from "../../../Components/PageContent/PageContent"
import PageTitle from "../../../Components/PageTitle/PageTitle"
import { Operator } from "../../../Constants/Filter"
import { DataType, FieldType } from "../../../Constants/Form"
import { api } from "../../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../../Service/Routes"
import { withTranslation } from "react-i18next"

class UpdatePage extends React.Component {
  state = {
    loading: false,
    showToken: false,
    token: "",
  }

  displayToken = (token) => {
    this.setState({ showToken: true, token: token })
  }

  hideToken = () => {
    this.setState({ showToken: false }, () => {
      this.redirectCallBack()
    })
  }

  redirectCallBack = () => {
    r(this.props.history, rMap.settings.serviceAccount.list)
  }

  componentDidMount() {}

  render() {
    const { loading, showToken, token } = this.state

    if (loading) {
      return <Loading key="loading" />
    }

    const { id } = this.props.match.params
    const { cancelFn = () => {}, t } = this.props

    const isNewEntry = id === undefined || id === ""

    const tokenModal = (
      <Modal
        position="top"
        variant={ModalVariant.large}
        isOpen={showToken}
        aria-label="no_header_footer"
        aria-describedby="modal-no-header-description"
        onClose={this.hideToken}
        title={t("service_account_copy_title")}
      >
        <ClipboardCopy isReadOnly hoverTip={t("copy")} clickTip={t("copied")}>
          {token}
        </ClipboardCopy>
        <TokenBarcode value={token} />
        <Alert
          style={{ marginTop: "10px" }}
          variant="warning"
          isInline
          title={t("service_account_copy_warning_msg")}
        />
        <Alert
          style={{ marginTop: "10px" }}
          variant="info"
          isInline
          title={t("service_account_barcode_help")}
        />
      </Modal>
    )

    const editor = (
      <Editor
        key="editor"
        resourceId={id}
        language="yaml"
        apiGetRecord={api.serviceAccount.get}
        apiSaveRecord={isNewEntry ? api.serviceAccount.create : api.serviceAccount.update}
        minimapEnabled
        onSaveRedirectFunc={(_data, saveResponse = {}) => {
          if (id) {
            cancelFn()
          } else {
            if (isNewEntry) {
              this.displayToken(saveResponse.token ? saveResponse.token : t("service_account_not_generated"))
            } else {
              this.redirectCallBack()
            }
          }
        }}
        onCancelFunc={() => {
          if (id) {
            cancelFn()
          } else {
            this.redirectCallBack()
          }
        }}
        getFormItems={(rootObject) => getFormItems(rootObject, isNewEntry)}
      />
    )

    if (isNewEntry) {
      return (
        <>
          <PageTitle key="page-title" title="add_a_service_account" />
          <PageContent hasNoPaddingTop>
            {editor}
            {isNewEntry ? tokenModal : null}
          </PageContent>
        </>
      )
    }
    return editor
  }
}

export default withTranslation()(UpdatePage)

// support functions

const getFormItems = (rootObject, isNew = false) => {
  // default values
  objectPath.set(rootObject, "neverExpire", true, true)

  const neverExpire = objectPath.get(rootObject, "neverExpire", false)
  if (neverExpire) {
    objectPath.set(rootObject, "expiresOn", "")
  }

  // do not set id, new id will be updated on the server side
  const items = [
    {
      label: "id",
      fieldId: "id",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      isDisabled: true,
    },
    {
      label: "username",
      fieldId: "username",
      fieldType: FieldType.SelectTypeAheadAsync,
      dataType: DataType.String,
      value: "",
      isRequired: false,
      isDisabled: !isNew,
      helperText: isNew ? "helper_text.service_account_user" : "",
      apiOptions: api.user.list,
      optionValueKey: "username",
      getFiltersFunc: (value) => {
        return [{ k: "username", o: Operator.Regex, v: value }]
      },
      isCreatable: false,
      optionValueFunc: (item) => item.username,
      getOptionsDescriptionFunc: (item) => item.id,
    },
    {
      label: "name",
      fieldId: "name",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_name",
      validated: "default",
      validator: { isLength: { min: 2, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "description",
      fieldId: "description",
      fieldType: FieldType.Text,
      dataType: DataType.String,
    },
    {
      label: "never_expire",
      fieldId: "neverExpire",
      fieldType: FieldType.Switch,
      dataType: DataType.Boolean,
      value: false,
      resetFields: { expiresOn: "" },
    },
  ]

  if (!neverExpire) {
    items.push({
      label: "expires_on",
      fieldId: "expiresOn",
      fieldType: FieldType.DatePicker,
      dataType: DataType.String,
      value: {},
      isRequired: true,
      validated: "default",
    })
  }

  if (!Array.isArray(rootObject.statements)) {
    rootObject.statements = []
  }

  items.push(
    {
      label: "statements",
      fieldId: "!statements_divider",
      fieldType: FieldType.Divider,
    },
    {
      label: "",
      fieldId: "statements",
      fieldType: FieldType.PolicyStatements,
      dataType: DataType.ArrayObject,
      value: [],
      isRequired: false,
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
      validated: "default",
      validator: { isLabel: {} },
    }
  )

  return items
}
