export const GridBreakPoints = { lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }
export const GridColumns = { lg: 100, md: 80, sm: 50, xs: 25, xxs: 10 }

export const WidgetTitleHeight = 34 // 34px

export const getWidgetWidth = (windowWidth, gridColumns) => {
  if (windowWidth > 1200) {
    return (windowWidth * (gridColumns / 100)).toFixed(0)
  } else if (windowWidth > 996) {
    return (windowWidth * (gridColumns / 80)).toFixed(0)
  } else if (windowWidth > 768) {
    return (windowWidth * (gridColumns / 50)).toFixed(0)
  } else if (windowWidth > 480) {
    return (windowWidth * (gridColumns / 25)).toFixed(0)
  } else {
    return (windowWidth * (gridColumns / 10)).toFixed(0)
  }
  // return windowWidth
}

export const getWidgetHeight = (windowHeight, gridHeight, showTitle = true) => {
  if (showTitle) {
    return (windowHeight * (gridHeight / 100)).toFixed(0) - WidgetTitleHeight
  }
  return (windowHeight * (gridHeight / 100)).toFixed(0)
}
