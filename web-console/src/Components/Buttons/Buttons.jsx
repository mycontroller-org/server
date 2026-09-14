import { Button, ButtonVariant } from "@patternfly/react-core"
import { EraserIcon, PlusIcon, SyncAltIcon } from "@patternfly/react-icons"
import React from "react"
import { useTranslation } from "react-i18next"
import { redirect as r } from "../../Service/Routes"

export const AddButton = ({ onClick }) => {
  const { t } = useTranslation()
  return button(t("add"), ButtonVariant.primary, PlusIcon, onClick)
}

export const RefreshButton = ({ onClick }) => {
  return button(null, ButtonVariant.plain, SyncAltIcon, onClick)
}

export const LinkButton = ({ text, icon, onClick, isSmall = false }) => {
  return button(text, ButtonVariant.link, icon, onClick, isSmall)
}

export const ResetButton = ({
  text = "reset",
  icon = EraserIcon,
  variant = ButtonVariant.warning,
  onClick,
  isSmall = false,
  isDisabled = false,
}) => {
  const { t } = useTranslation()
  return button(t(text), variant, icon, onClick, isSmall, isDisabled)
}

export const CustomButton = ({ text = "", icon = null, variant = "secondary", onClick, isSmall = false }) => {
  return button(text, variant, icon, onClick, isSmall)
}

const button = (text, variant, icon, onClickFn = () => {}, isSmall = false, isDisabled = false) => {
  const Icon = icon
  if (text === undefined || text === null || text === "") {
    return (
      <Button isSmall={isSmall} onClick={onClickFn} variant={variant} isDisabled={isDisabled}>
        {icon ? <Icon /> : null}
      </Button>
    )
  } else {
    return (
      <Button
        isSmall={isSmall}
        icon={icon ? <Icon /> : null}
        onClick={onClickFn}
        variant={variant}
        isDisabled={isDisabled}
      >
        {text}
      </Button>
    )
  }
}

export const RouteLink = ({ history, path, id, text }) => {
  return (
    <Button
      variant="link"
      isInline
      onClick={(_e) => {
        r(history, path, { id: id })
      }}
    >
      {text}
    </Button>
  )
}

export const IconButton = ({ icon, color = "#434343", onClick = () => {}, className = "" }) => {
  const IconBtn = icon
  return <IconBtn className={className} color={color} onClick={onClick} onTouchEnd={onClick} />
}
