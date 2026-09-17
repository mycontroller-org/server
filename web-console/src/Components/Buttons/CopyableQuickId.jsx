import { ClipboardCopy } from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import "./CopyableQuickId.scss"

const copyWithExecCommand = (text) => {
  const el = document.createElement("textarea")
  el.value = text
  el.setAttribute("readonly", "")
  el.style.position = "fixed"
  el.style.top = "0"
  el.style.left = "0"
  el.style.opacity = "0"
  document.body.appendChild(el)
  el.focus()
  el.select()
  el.setSelectionRange(0, text.length)
  const ok = document.execCommand("copy")
  document.body.removeChild(el)
  return ok
}

const copyText = (_event, text) => {
  const value = text == null ? "" : String(text)
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(value).catch(() => {
      copyWithExecCommand(value)
    })
    return
  }
  copyWithExecCommand(value)
}

const CopyableQuickId = ({ value }) => {
  const { t } = useTranslation()
  if (!value) {
    return <span>-</span>
  }
  return (
    <span
      className="copyable-quick-id"
      onClick={(event) => {
        event.preventDefault()
        event.stopPropagation()
      }}
    >
      <ClipboardCopy
        variant="inline-compact"
        isReadOnly
        hoverTip={t("copy")}
        clickTip={t("copied")}
        onCopy={copyText}
      >
        {value}
      </ClipboardCopy>
    </span>
  )
}

export default CopyableQuickId
