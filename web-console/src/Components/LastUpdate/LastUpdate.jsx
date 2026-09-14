import moment from "moment"
import React from "react"
import "./LastUpdate.scss"
import { useTranslation } from "react-i18next"

const LastUpdate = ({ time }) => {
  const { t } = useTranslation()
  const lastUpdate = moment(time)
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
