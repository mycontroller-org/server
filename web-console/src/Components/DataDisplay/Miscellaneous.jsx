import { Label, List, ListItem, Split, SplitItem } from "@patternfly/react-core"
import { filesize as fileSize } from "filesize"
import React from "react"
import { useTranslation } from "react-i18next"
import { getValue } from "../../Util/Util"
import { LastSeen } from "../Time/Time"

export const DisplayEnabled = ({ data, field, defaultValue = false }) => {
  const { t } = useTranslation()
  const value = getValue(data, field, defaultValue)
  return <span>{value ? t("enabled") : t("disabled")}</span>
}

export const DisplayTrue = ({ data, field, defaultValue = false }) => {
  const value = getValue(data, field, defaultValue)
  return <span>{value ? "true" : "false"}</span>
}

export const DisplaySuccess = ({ data, field, defaultValue = false }) => {
  const value = getValue(data, field, defaultValue)
  return <span>{value ? "success" : "failed"}</span>
}

export const DisplayList = ({ data, field, defaultValue = [] }) => {
  const values = getValue(data, field, defaultValue)
  const finalData = []
  if (Array.isArray(values)) {
    values.forEach((v) => {
      finalData.push(<ListItem key={v}>{v}</ListItem>)
    })
  }
  return <List>{finalData}</List>
}

export const FileSize = ({ bytes = 0 }) => {
  return <span>{fileSize(bytes, { standard: "iec" })}</span>
}

export const DisplayImage = ({ data = "" }) => {
  return <img width="200" height="200" src={data} />
}

export const DisplayFieldValue = ({ value = "", timestamp = "" }) => {
  return (
    <Split hasGutter className="mc-resource-variables">
      <SplitItem>
        <Label variant="filled" color="grey">
          <strong>{String(value)}</strong>
        </Label>
      </SplitItem>
      <SplitItem>
        <LastSeen date={timestamp} tooltipPosition="top" />
      </SplitItem>
    </Split>
  )
}
