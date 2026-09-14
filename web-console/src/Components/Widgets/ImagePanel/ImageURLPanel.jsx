import React from "react"
import PropTypes from "prop-types"
import Loading from "../../Loading/Loading"
import { getValue, isEqual } from "../../../Util/Util"
import { Stack, StackItem } from "@patternfly/react-core"
import { RefreshIntervalType } from "../../../Constants/Metric"
import { ImageSourceType } from "../../../Constants/Widgets/ImagePanel"
import { connect } from "react-redux"
import "./ImagePanel.scss"

class ImageURLPanel extends React.Component {
  state = {
    isLoading: true,
    error: "",
    refreshHash: Date.now(),
  }

  updateRefreshInterval = () => {
    // update metrics query on interval
    if (this.interval) {
      clearInterval(this.interval)
    }
    const { config } = this.props
    const refreshInterval = getValue(config, "refreshInterval", RefreshIntervalType.None)
    if (refreshInterval >= 1000) {
      this.interval = setInterval(() => {
        const diffTime = new Date().getTime() - this.lastMetricUpdate + 200 // add 200 milliseconds as offset
        if (diffTime >= refreshInterval) {
          this.updateComponents()
        }
      }, refreshInterval)
    }
  }

  componentDidMount() {
    this.updateComponents()
    this.updateRefreshInterval()
  }

  componentWillUnmount() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  }

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.config, this.props.config)) {
      this.updateComponents()
      this.updateRefreshInterval()
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return !isEqual(this.state, nextState) || !isEqual(this.props, nextProps)
  }

  updateComponents = () => {
    this.lastMetricUpdate = new Date().getTime() // reference point for auto refresh job
    if (this.props.config === undefined) {
      this.setState({ isLoading: false, error: "config not found" })
      return
    }

    this.setState({ isLoading: false, refreshHash: Date.now() })
  }

  render() {
    const { isLoading, error, refreshHash } = this.state

    if (error !== "") {
      return <span>Error: {error}</span>
    }

    if (isLoading) {
      return <Loading />
    }

    const { sourceType, imageURL, imageLocation, rotation } = this.props.config
    const { authToken } = this.props

    let finalUrl = `${imageURL}?mcRefreshHash=${refreshHash}`
    if (sourceType === ImageSourceType.Disk) {
      finalUrl = `${imageLocation}?access_token=${authToken}&mcRefreshHash=${refreshHash}`
    }

    return (
      <Stack>
        <StackItem isFilled>
          <img
            className="auto-resize"
            style={{ transform: `rotate(${rotation}deg)` }}
            src={finalUrl}
            alt="dynamic data"
          />
        </StackItem>
        <StackItem></StackItem>
      </Stack>
    )
  }
}

ImageURLPanel.propTypes = {
  config: PropTypes.object,
}

const mapStateToProps = (state) => ({
  authToken: state.entities.auth.token,
})

export default connect(mapStateToProps)(ImageURLPanel)
