import React from "react"
import { api } from "../../../../Service/Api"
import { loadData, unloadData } from "../../../../store/entities/websocket"
import { connect } from "react-redux"
import Loading from "../../../Loading/Loading"
import { getValue, isEqual } from "../../../../Util/Util"
import { getResource } from "../Common/Utils"
import ControlObjects from "../Common/Common"
import { ControlType } from "../../../../Constants/Widgets/ControlPanel"
import { Button, Modal, ModalVariant } from "@patternfly/react-core"
import { isShouldComponentUpdateWithWsData } from "../../Helper/Common"

const wsKey = "dashboard_control_panel_mixed"

class MixedControl extends React.Component {
  state = {
    isLoading: true,
    isModalOpen: false,
    modalCallBack: () => {},
    message: "",
  }

  getWsKey = () => {
    return `${wsKey}_${this.props.widgetId}`
  }

  componentDidMount() {
    this.updateComponents()
  }

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.config, this.props.config)) {
      this.updateComponents()
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return isShouldComponentUpdateWithWsData(this.getWsKey(), this.props, this.state, nextProps, nextState)
  }

  componentWillUnmount() {
    this.props.unloadData({ key: this.getWsKey() })
  }
  updateComponents = () => {
    if (this.props.config === undefined) {
      return
    }

    const { resources: resourcesRaw = [] } = this.props.config
    if (resourcesRaw.length === 0) {
      this.setState({ isLoading: false })
      return
    }

    const resourceQuickIds = resourcesRaw.map((res) => {
      const { type, quickId: id } = res.resource
      return `${type}:${id}`
    })

    api.quickId
      .getResources({ id: resourceQuickIds })
      .then((res) => {
        const { data = {} } = res
        this.props.loadData({ key: this.getWsKey(), resources: data })
        this.setState({ isLoading: false })
      })
      .catch((_e) => {
        this.setState({ isLoading: false })
      })
  }

  hideModal = () => {
    this.setState({ isModalOpen: false })
  }

  sendPayloadWrapper = (askConfirmation = false, message = "", callbackFunc = () => {}) => {
    if (askConfirmation) {
      this.setState({ isModalOpen: true, message: message, modalCallBack: callbackFunc })
    } else {
      callbackFunc()
    }
  }

  render() {
    const { isLoading, isModalOpen, message, modalCallBack } = this.state

    if (isLoading) {
      return <Loading />
    }
    const { wsData, widgetId, config = {}, history } = this.props
    const { resources: resourcesConfig = [] } = config

    if (resourcesConfig.length === 0) {
      return <span>No resources configured</span>
    }

    // console.log("wsData:", this.props.wsData)
    const resourcesRaw = getValue(wsData, this.getWsKey(), {})

    const resources = []

    resourcesConfig.forEach((resCfg) => {
      const {
        type: resourceType,
        quickId: qId,
        nameKey: resourceNameKey,
        valueTimestampKey: resourceTimestampKey,
        keyPath,
      } = resCfg.resource
      const quickId = `${resourceType}:${qId}`
      const resourceRaw = resourcesRaw[quickId]

      if (resourceRaw !== undefined) {
        const resource = getResource(
          resourceType,
          resourceRaw,
          resourceNameKey,
          resourceTimestampKey,
          ControlType.MixedControl,
          keyPath
        )
        // update config
        resource["config"] = resCfg
        resources.push(resource)
      } else {
        // TODO: ...
        console.log("data not available for quickId", quickId, resourceRaw)
      }
    })

    return (
      <>
        <ControlObjects
          key="control_objects"
          widgetId={widgetId}
          config={config}
          history={history}
          payloadOn="false"
          resources={resources}
          sendPayloadWrapper={this.sendPayloadWrapper}
        />
        <Modal
          variant={ModalVariant.small}
          position="top"
          title="Need your confirm"
          isOpen={isModalOpen}
          onClose={this.hideModal}
          actions={[
            <Button
              key="confirm"
              variant="warning"
              onClick={() => {
                this.hideModal()
                modalCallBack()
              }}
            >
              Confirm
            </Button>,
            <Button key="cancel" variant="tertiary" onClick={this.hideModal}>
              Cancel
            </Button>,
          ]}
        >
          {message}
        </Modal>
      </>
    )
  }
}

const mapStateToProps = (state) => ({
  wsData: state.entities.websocket.data,
})

const mapDispatchToProps = (dispatch) => ({
  loadData: (data) => dispatch(loadData(data)),
  unloadData: (data) => dispatch(unloadData(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(MixedControl)
