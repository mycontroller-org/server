import YAML from "js-yaml"
import { Language } from "../Constants/CodeEditor"

export const toString = (language = "", data = {}) => {
  if (data === undefined || data === "") {
    return ""
  }
  switch (language) {
    case Language.YAML:
    case Language.JSON:
      if (data === undefined || data === "" || Object.keys(data).length === 0) {
        return ""
      }
      if (language === Language.YAML) {
        return YAML.dump(data)
      } else if (language === Language.JSON) {
        return JSON.stringify(data)
      }
      break

    case Language.JAVASCRIPT:
    case Language.HTML:
    case Language.PLANTEXT:
    case Language.XML:
      return data

    default:
      return Object.toString(data)
  }
}

export const toObject = (language = "", data = "") => {
  switch (language) {
    case Language.YAML:
    case Language.JSON:
      if (data === undefined || data === "") {
        return {}
      }
      if (language === "yaml") {
        return YAML.load(data)
      } else if (language === "json") {
        return JSON.parse(data)
      }
      break

    case Language.JAVASCRIPT:
    case Language.HTML:
    case Language.PLANTEXT:
    case Language.XML:
      return data

    default:
      return { error: "language '" + language + "' not supported" }
  }
}
