import { Button, Divider, Split, SplitItem, Stack, StackItem } from "@patternfly/react-core"
import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { ResourceType } from "../../../Constants/Resource"
import { ICON_PREFIX, ImageType } from "../../../Constants/Widgets/ImagePanel"
import { api } from "../../../Service/Api"
import { loadData, unloadData } from "../../../store/entities/websocket"
import { getValue, isEqual } from "../../../Util/Util"
import Loading from "../../Loading/Loading"
import { LastSeen } from "../../Time/Time"
import { navigateToResource } from "../Helper/Resource"
import { getIcon } from "./IconHelper"
import "./ImagePanel.scss"

const wsKey = "dashboard_image_panel_image_from_field"

class ImageFromFieldPanel extends React.Component {
  state = {
    isLoading: true,
    error: "",
    refreshHash: Date.now(),
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
    return !isEqual(this.state, nextState) || !isEqual(this.props, nextProps)
  }

  updateComponents = () => {
    if (this.props.config === undefined) {
      this.setState({ isLoading: false, error: "config not found" })
      return
    }

    const { field } = this.props.config
    const resourceIDs = [field.id]

    api.quickId
      .getResources({ id: resourceIDs })
      .then((res) => {
        this.props.loadData({ key: this.getWsKey(), resources: res.data })
        this.setState({ isLoading: false, refreshHash: Date.now() })
      })
      .catch((_e) => {
        this.setState({ isLoading: false, refreshHash: Date.now() })
      })
  }

  render() {
    const { isLoading, error, refreshHash } = this.state

    if (error !== "") {
      return <span>Error: {error}</span>
    }

    if (isLoading) {
      return <Loading />
    }

    const { field, rotation } = this.props.config
    const { wsData, history, authToken, dimensions } = this.props

    let resourceData = {}

    const resourcesRaw = getValue(wsData, this.getWsKey(), {})
    Object.keys(resourcesRaw).forEach((id) => {
      if (field.id === id) {
        resourceData = resourcesRaw[id]
      }
    })

    const resourceValue = getValue(resourceData, "current.value", "")
    let imageElement = null
    switch (field.type) {
      case ImageType.Image:
        imageElement = getImage(resourceValue, rotation)
        break

      case ImageType.CustomMapping:
        imageElement = getMappedImageOrIcon(
          resourceValue,
          field,
          rotation,
          authToken,
          refreshHash,
          dimensions
        )
        break

      default:
        imageElement = <span>Unknown image type: {field.type}</span>
    }

    let resourceDetails = null

    if (field.displayValue) {
      const resourceNameKey = field.nameKey === "" || field.nameKey === undefined ? "name" : field.nameKey
      const resourceName = getValue(resourceData, resourceNameKey, "undefined")
      resourceDetails = (
        <div className="image-field">
          <StackItem>
            <Divider className="divider" component="li" />
          </StackItem>
          <StackItem>
            <Split>
              <SplitItem isFilled>
                <Button
                  className="name"
                  variant="link"
                  onClick={() =>
                    navigateToResource(ResourceType.Field, getValue(resourceData, "id", ""), history)
                  }
                >
                  {resourceName}
                </Button>
              </SplitItem>
              <SplitItem style={{ textAlign: "right" }}>
                <span className="value">{resourceValue}</span>
              </SplitItem>
            </Split>
          </StackItem>
          <StackItem>
            <span className="value-timestamp">
              <LastSeen date={getValue(resourceData, "current.timestamp", "")} tooltipPosition="top" />
            </span>
          </StackItem>
        </div>
      )
    }

    return (
      <Stack>
        <StackItem isFilled style={{ overflow: "hidden" }}>
          {imageElement}
        </StackItem>
        {resourceDetails}
      </Stack>
    )
  }
}

ImageFromFieldPanel.propTypes = {
  config: PropTypes.object,
}

const mapStateToProps = (state) => ({
  wsData: state.entities.websocket.data,
  authToken: state.entities.auth.token,
})

const mapDispatchToProps = (dispatch) => ({
  loadData: (data) => dispatch(loadData(data)),
  unloadData: (data) => dispatch(unloadData(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(ImageFromFieldPanel)

// helper functions

const getMappedImageOrIcon = (
  resourceValue = "",
  field = {},
  rotation = 0,
  authToken = "",
  refreshHash = "",
  dimensions = {}
) => {
  const thresholdMode = getValue(field, "thresholdMode", false)

  // threshold mode
  // sort available keys and get the nearest key
  if (thresholdMode) {
    const keys = Object.keys(getValue(field, "custom_mapping", [])).sort((a, b) => a - b)
    if (keys.length > 0) {
      let key = keys[0]
      if (keys.length > 1) {
        keys.forEach((v) => {
          if (resourceValue >= v) {
            key = v
          }
        })
      }
      resourceValue = key
    }
  }

  // get image or icon detail
  const value = getValue(field, `custom_mapping.${resourceValue}`, "")

  // if it is a icon
  if (value.startsWith(ICON_PREFIX)) {
    let iconValue = value.replace(ICON_PREFIX, "")
    if (iconValue.indexOf(":") !== -1) {
      iconValue = iconValue.substring(iconValue.indexOf(":") + 1)
    }
    return getIcon(iconValue, rotation, dimensions)
  } else {
    // load custom mapping image
    const imageData = `${value}?access_token=${authToken}&mcRefreshHash=${refreshHash}`
    return getImage(imageData, rotation)
  }
}

const getImage = (imageData = "", rotation = 0) => {
  return (
    <img
      className="auto-resize"
      style={{
        transform: `rotate(${rotation}deg)`,
        objectFit: "contain",
      }}
      src={imageData}
      align="left"
      alt="dynamic data"
    />
  )
}
