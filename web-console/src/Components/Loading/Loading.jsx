import React from "react"
import { useTranslation } from "react-i18next"

const Loading = () => {
  const { t } = useTranslation()
  return <span key="loading">{t("loading")}...</span>
}

export default Loading
