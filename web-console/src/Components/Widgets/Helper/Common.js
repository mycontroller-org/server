import { getValue, isEqual } from "../../../Util/Util"

// this function verifies, really there is an update required
// this should be used only for wsData component
export const isShouldComponentUpdateWithWsData = (
  wsKey = "",
  thisProps = {},
  thisState = {},
  nextProps = {},
  nextState = {}
) => {
  const thisWsData = getValue(thisProps, `wsData.${wsKey}`, "this")
  const nextWsData = getValue(nextProps, `wsData.${wsKey}`, "next")
  if (!isEqual(thisWsData, nextWsData)) {
    return true
  } else if (
    !isEqual(thisProps.widgetId, nextProps.widgetId) ||
    !isEqual(thisProps.config, nextProps.config) ||
    !isEqual(thisProps.dimensions, nextProps.dimensions) ||
    !isEqual(thisProps.history, nextProps.history)
  ) {
    return true
  }
  return !isEqual(thisState, nextState)
}
