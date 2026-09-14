import React from "react"
import { ImageSourceType } from "../../../Constants/Widgets/ImagePanel"
import { getValue } from "../../../Util/Util"
import ImageFieldPanel from "./ImageFieldPanel"
import ImageURLPanel from "./ImageURLPanel"

const ImagePanel = ({ widgetId = "", config = {}, history = {}, dimensions = {} }) => {
  const sourceType = getValue(config, "sourceType", "")

  switch (sourceType) {
    case ImageSourceType.Field:
      return <ImageFieldPanel widgetId={widgetId} config={config} history={history} dimensions={dimensions} />

    case ImageSourceType.Disk:
    case ImageSourceType.URL:
      return <ImageURLPanel widgetId={widgetId} config={config} history={history} dimensions={dimensions} />

    default:
      return <span>{`Selected image source type[${sourceType}] not available`}</span>
  }
}

export default ImagePanel
