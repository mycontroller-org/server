import { Button, List, ListItem, Modal, ModalVariant } from "@patternfly/react-core"
import React from "react"
import { Trans, useTranslation } from "react-i18next"
import "./Dialog.scss"

const DeleteDialog = ({ dialogTitle, show, onCloseFn, onOkFn }) => {
  const { t } = useTranslation()
  return (
    <Modal
      className="mc-model"
      variant={ModalVariant.small}
      position="top"
      title={<Trans i18nKey="dialog.delete_resource_heading">{t(dialogTitle)}</Trans>}
      isOpen={show}
      onClose={onCloseFn}
      showClose={false}
      actions={[
        <Button key="cancel" variant="secondary" onClick={onCloseFn}>
          {t("cancel")}
        </Button>,
        <Button key="confirm" variant="danger" onClick={onOkFn}>
          {t("delete")}
        </Button>,
      ]}
    >
      {t("dialog.delete_resource_msg")}
    </Modal>
  )
}

export const NodeRebootDialog = ({ show, onCloseFn, onOkFn }) => {
  const { t } = useTranslation()
  return dialog(
    t("dialog.reboot_nodes_heading"),
    show,
    onCloseFn,
    onOkFn,
    t("reboot"),
    t("dialog.reboot_nodes_msg"),
    t
  )
}

export const ClearSleepingQueueDialog = ({ show, onCloseFn, onOkFn }) => {
  const { t } = useTranslation()
  return dialog(
    t("dialog.clear_sleeping_queue_heading"),
    show,
    onCloseFn,
    onOkFn,
    t("clear"),
    t("dialog.clear_sleeping_queue_msg"),
    t
  )
}

export const NodeResetDialog = ({ show, onCloseFn, onOkFn }) => {
  const { t } = useTranslation()
  const message = (
    <Trans i18nKey="dialog.node_reset_msg">
      Are you sure want to reset the selected node(s)?
      <br />
      After this operation:
      <List>
        <ListItem>Configuration will be restored to factory settings.</ListItem>
        <ListItem>you may lose access to the selected node(s) from MyController!</ListItem>
      </List>
    </Trans>
  )
  return dialog(t("dialog.node_reset_heading"), show, onCloseFn, onOkFn, t("reset"), message, t)
}

export const RestoreDialog = ({ show, fileName, onCloseFn, onOkFn }) => {
  const { t } = useTranslation()
  const message = (
    <Trans i18nKey="dialog.restore_msg">
      Are you sure want to restore to selected backup file?
      <br />
      After this operation:
      <List>
        <ListItem>
          Server data will be restored to <b>{{ fileName }}</b>
        </ListItem>
        <ListItem>You cannot rollback this operation</ListItem>
        <ListItem>Always take a backup before performing a restore operation</ListItem>
        <ListItem>You may need to start the server manually</ListItem>
      </List>
    </Trans>
  )
  return dialog(t("dialog.restore_heading"), show, onCloseFn, onOkFn, t("restore"), message, t)
}

export const CommonDialog = ({
  dialogTitle,
  dialogMsg,
  show,
  onCloseFn,
  onOkFn,
  confirmBtnText = "continue",
  variant = "warning",
}) => {
  const { t } = useTranslation()
  return (
    <Modal
      className="mc-model"
      variant={ModalVariant.small}
      position="top"
      title={t(dialogTitle)}
      isOpen={show}
      onClose={onCloseFn}
      showClose={false}
      actions={[
        <Button key="cancel" variant="secondary" onClick={onCloseFn}>
          {t("cancel")}
        </Button>,
        <Button key="confirm" variant={variant} onClick={onOkFn}>
          {t(confirmBtnText)}
        </Button>,
      ]}
    >
      {t(dialogMsg)}
    </Modal>
  )
}

const dialog = (
  title = "",
  isOpen = false,
  onCloseFn = () => {},
  onOkFn = () => {},
  okBtnName = "NoName",
  message = "",
  t = () => {}
) => {
  return (
    <Modal
      key={title}
      className="mc-model"
      variant={ModalVariant.small}
      position="top"
      title={title}
      isOpen={isOpen}
      onClose={onCloseFn}
      showClose={false}
      actions={[
        <Button key="cancel" variant="secondary" onClick={onCloseFn}>
          {t("cancel")}
        </Button>,
        <Button key="confirm" variant="danger" onClick={onOkFn}>
          {okBtnName}
        </Button>,
      ]}
    >
      {message}
    </Modal>
  )
}

export default DeleteDialog
