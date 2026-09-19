import React from "react"
import "./LastUpdate.scss"
import { useTranslation } from "react-i18next"
import { formatAbsolute, formatFromNow } from "../../i18n/relativeTime"

const LastUpdate = ({ time }) => {
  const { t, i18n } = useTranslation()
  return (
    <span className="mc-last-update">
      {t("data_retrieved_from_server")}:{" "}
      <i>
        {formatFromNow(time, i18n.language, t)} ({formatAbsolute(time, i18n.language)})
      </i>
    </span>
  )
}

export default LastUpdate
