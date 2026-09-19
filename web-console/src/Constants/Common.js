export const NOTIFICATION_LIMIT = 50

export const TOAST_ALERT_TIMEOUT = 5000

export const COPYRIGHT_YEAR = "2015 - 2026"

export const ALERT_TYPE_DEFAULT = "default"
export const ALERT_TYPE_INFO = "info"
export const ALERT_TYPE_SUCCESS = "success"
export const ALERT_TYPE_WARN = "warning"
export const ALERT_TYPE_ERROR = "danger"

export const DATA_CACHE_TIMEOUT = 0 // in seconds, data cache in list tables. 0 - disabled

export const SECRET_ENCRYPTION_PREFIX = "ENCRYPTED;MC;AES256;BASE64;"

export const MINIMUM_LOGIN_EXPIRES_IN_DAYS = "24h" // 1 day
export const MAXIMUM_LOGIN_EXPIRES_IN_DAYS = "720h" // 24 * 30
export const URL_HOMEPAGE = "https://mycontroller.org"
export const URL_DOCUMENTATION = "https://v2.mycontroller.org/docs/"
export const URL_FORUM = "https://forum.mycontroller.org"
export const URL_SOURCE_CODE = "https://github.com/mycontroller-org/server"

export const DEFAULT_LANGUAGE = "en_GB"

export const THEME_LIGHT = "light"
export const THEME_DARK = "dark"
export const THEME_SYSTEM = "system"
export const DEFAULT_THEME = THEME_SYSTEM

export const THEME_OPTIONS = [
  { value: THEME_SYSTEM, label: "theme_system" },
  { value: THEME_LIGHT, label: "theme_light" },
  { value: THEME_DARK, label: "theme_dark" },
]

export const normalizeTheme = (value) => {
  if (value === THEME_LIGHT || value === THEME_DARK || value === THEME_SYSTEM) {
    return value
  }
  return THEME_SYSTEM
}

export const resolveTheme = (selection) => {
  const theme = normalizeTheme(selection)
  if (theme !== THEME_SYSTEM) {
    return theme
  }
  if (typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return THEME_DARK
  }
  return THEME_LIGHT
}

export const SECURE_SHARE_PREFIX = "/secure_share/"
export const INSECURE_SHARE_PREFIX = "/insecure_share/"

export const NAV_WIDTH = 290

export const getWindowWidth = (windowSizeWidth, isNavOpen) => {
  if (windowSizeWidth < 1200) {
    return windowSizeWidth
  } else if (isNavOpen) {
    return windowSizeWidth - NAV_WIDTH
  }
  return windowSizeWidth
}

export const DropDownPositionType = {
  UP: "up",
  DOWN: "down",
}

export const DropDownPositionTypeOptions = [
  { value: DropDownPositionType.DOWN, label: "down" },
  { value: DropDownPositionType.UP, label: "up" },
]

export const FIELD_NAME = "fieldname"
export const NodeField = {
  BATTERY_LEVEL: "battery_level",
}
