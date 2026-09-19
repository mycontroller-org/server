import moment from "moment"
import React from "react"
import "./LastUpdate.scss"
import { useTranslation } from "react-i18next"
import { toMomentLocale } from "../../i18n/momentLocale"

const LastUpdate = ({ time }) => {
  const { t, i18n } = useTranslation()
  const lastUpdate = moment(time).locale(toMomentLocale(i18n.language))
  return (
    <span className="mc-last-update">
      {t("data_retrieved_from_server")}:{" "}
      <i>
        {lastUpdate.fromNow()} ({lastUpdate.format("MMM Do YYYY, HH:mm:ss")})
      </i>
    </span>
  )
}

export default LastUpdate
