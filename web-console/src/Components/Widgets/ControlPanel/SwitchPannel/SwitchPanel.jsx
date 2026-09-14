import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import Loading from "../../../Loading/Loading"
import { getResource, getListAPI } from "../Common/Utils"
import { loadData, unloadData } from "../../../../store/entities/websocket"
import { getValue, isEqual } from "../../../../Util/Util"
import { getQuickId } from "../../../../Constants/ResourcePicker"
import ControlObjects from "../Common/Common"
import { isShouldComponentUpdateWithWsData } from "../../Helper/Common"

const wsKey = "dashboard_control_panel_switch"

class SwitchPanel extends React.Component {
  state = {
    isLoading: true,
  }

  getWsKey = () => {
    return `${wsKey}_${this.props.widgetId}`
  }

  componentWillUnmount() {
    this.props.unloadData({ key: this.getWsKey() })
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

  updateComponents = () => {
    if (this.props.config === undefined) {
      return
    }
    const {
      filters: resourceFilters = {},
      limit: itemsLimit,
      type: resourceType,
    } = this.props.config.resource
    const filterKeys = Object.keys(resourceFilters)
    const filters = filterKeys.map((key) => {
      return { k: key, v: resourceFilters[key] }
    })

    const listAPI = getListAPI(resourceType)

    listAPI({ filter: filters, limit: itemsLimit })
      .then((res) => {
        const resources = {}
        res.data.data.forEach((item) => {
          // return getField(resourceType, field, resourceNameKey)
          const quickId = getQuickId(resourceType, item)
          resources[quickId] = item
        })
        this.props.loadData({ key: this.getWsKey(), resources: resources })
        this.setState({ isLoading: false, resources })
      })
      .catch((_e) => {
        this.setState({ isLoading: false })
      })
  }

  render() {
    const { isLoading } = this.state

    if (isLoading) {
      return <Loading />
    }
    const { type: resourceType, nameKey: resourceNameKey } = this.props.config.resource
    const { wsData, widgetId, config, history } = this.props

    // console.log("wsData:", this.props.wsData)
    const resourcesRaw = getValue(wsData, this.getWsKey(), {})
    const resources = []
    Object.keys(resourcesRaw).forEach((id) => {
      const field = resourcesRaw[id]
      resources.push(getResource(resourceType, field, resourceNameKey))
    })

    return (
      <ControlObjects
        key={`control_objects_${widgetId}`}
        widgetId={widgetId}
        config={config}
        history={history}
        payloadOn="true"
        resources={resources}
        sendPayloadWrapper={(_askConfirmation, _message, callbackFunc) => {
          callbackFunc()
        }}
      />
    )
  }
}

SwitchPanel.propTypes = {
  config: PropTypes.object,
}

const mapStateToProps = (state) => ({
  wsData: state.entities.websocket.data,
})

const mapDispatchToProps = (dispatch) => ({
  loadData: (data) => dispatch(loadData(data)),
  unloadData: (data) => dispatch(unloadData(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(SwitchPanel)
