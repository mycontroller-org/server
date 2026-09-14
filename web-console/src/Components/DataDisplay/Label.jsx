import { Flex, FlexItem, Label as PfLabel } from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import IFrame from "../IFrame/IFrame"
import { LastSeen } from "../Time/Time"
import "./Label.scss"
import { DisplayImage, FileSize } from "./Miscellaneous"

export const Label = ({ name, value, index = 0 }) => {
  return (
    <FlexItem key={name + index} className="mc-label">
      <span className="label-pair">
        <PfLabel className="label-key">{name}</PfLabel>
        {value && value.length > 0 && <PfLabel className="label-value">{value || ""}</PfLabel>}
      </span>
    </FlexItem>
  )
}

export const Labels = ({ data = {} }) => {
  const { t } = useTranslation()
  if (data === null) {
    return <span>{t("no_labels")}</span>
  }
  const names = Object.keys(data)

  if (names.length === 0) {
    return <span>{t("no_labels")}</span>
  }

  names.sort()

  const elements = names.map((name, index) => {
    return <Label key={index} name={name} value={data[name]} index={index} />
  })
  return <Flex>{elements}</Flex>
}

const hideItems = ["password", "token"]

export const KeyValue = ({ name, value, index = 0 }) => {
  let finalValue = ""
  switch (name) {
    case "node_web_url":
      finalValue = <IFrame url={value} />
      break

    case "modifiedOn":
    case "timestamp":
    case "lastEvaluation":
    case "lastSuccess":
    case "lastRun":
    case "since":
    case "ota_start_time":
    case "ota_end_time":
    case "ota_status_on":
      finalValue = <LastSeen date={value} tooltipPosition="top" />
      break

    case "size":
      finalValue = <FileSize bytes={value} />
      break

    default:
      if (String(value).startsWith("data:image")) {
        finalValue = <DisplayImage key={index} data={value} />
      } else {
        finalValue = hideItems.includes(name) ? "******" : value !== undefined ? String(value) : ""
      }
  }
  return getKeyValue(name, finalValue, index)
}

export const getKeyValue = (key = "", value = "", index = "") => {
  return (
    <div className="key-value-map" key={key + index}>
      <span className="key">{key}</span> <span className="value">{value}</span>
    </div>
  )
}

export const Statements = ({ statements = [] }) => {
  const items = Array.isArray(statements) ? statements : []
  const data = {}
  items.forEach((st, index) => {
    const n = index + 1
    data[`${n}.effect`] = st.effect || "Allow"
    data[`${n}.actions`] = Array.isArray(st.actions) ? st.actions.join(", ") : ""
    data[`${n}.resources`] = Array.isArray(st.resources) ? st.resources.join(", ") : ""
  })
  return <KeyValueMap data={data} />
}

export const KeyValueMap = ({ data = {} }) => {
  const { t } = useTranslation()

  if (data === undefined || data === null) {
    return <span>{t("no_data")}</span>
  }
  const names = Object.keys(data)
  if (names.length === 0) {
    return <span>{t("no_data")}</span>
  }

  names.sort()

  const elements = names.map((name, index) => {
    const value = typeof data[name] !== "object" ? data[name] : JSON.stringify(data[name], null, " ")
    return <KeyValue key={index} name={name} value={value} index={index} />
  })
  return elements
}
