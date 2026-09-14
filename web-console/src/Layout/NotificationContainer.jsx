import {
  Dropdown,
  DropdownItem,
  DropdownPosition,
  EmptyState,
  EmptyStateBody,
  EmptyStateIcon,
  EmptyStateVariant,
  KebabToggle,
  NotificationDrawer,
  NotificationDrawerBody,
  NotificationDrawerHeader,
  NotificationDrawerListItem,
  NotificationDrawerListItemBody,
  NotificationDrawerListItemHeader,
  Title,
} from "@patternfly/react-core"
import { SearchIcon } from "@patternfly/react-icons"
import moment from "moment"
import React from "react"
import { connect } from "react-redux"
import {
  notificationClearAll,
  notificationDrawerToggle,
  notificationMarkAllRead,
  notificationMarkAsRead,
} from "../store/entities/notification"
import { withTranslation } from "react-i18next"

class NotificationContainer extends React.Component {
  state = {
    isOptionsOpen: false,
  }

  toggleOptions = () => {
    this.setState((prevState) => {
      return { isOptionsOpen: !prevState.isOptionsOpen }
    })
  }

  notificationDrawerActions = [
    <DropdownItem key="markAllRead" onClick={this.props.markAllRead} component="button">
      {this.props.t("mark_all_read")}
    </DropdownItem>,
    <DropdownItem key="clearAll" onClick={this.props.clearAll} component="button">
      {this.props.t("clear_all")}
    </DropdownItem>,
  ]

  render() {
    const elements = []
    const { t } = this.props
    if (this.props.items.length > 0) {
      for (let index = this.props.items.length - 1; index >= 0; index--) {
        const a = this.props.items[index]
        const time = moment(a.timestamp)
        elements.push(
          <NotificationDrawerListItem
            key={a.id}
            variant={a.type}
            isRead={!a.unread}
            onClick={a.unread ? () => this.props.markAsRead(a.id) : () => {}}
          >
            <NotificationDrawerListItemHeader variant={a.type} title={a.title} />
            <NotificationDrawerListItemBody timestamp={time.fromNow()}>
              {a.description.map((txt, index) => {
                return <p key={index}>{txt}</p>
              })}
            </NotificationDrawerListItemBody>
          </NotificationDrawerListItem>
        )
      }
    } else {
      elements.push(
        <EmptyState key="empty" variant={EmptyStateVariant.full}>
          <EmptyStateIcon icon={SearchIcon} />
          <Title headingLevel="h2" size="lg">
            {t("alerts_empty_title")}
          </Title>
          <EmptyStateBody>{t("alerts_empty_message")}</EmptyStateBody>
        </EmptyState>
      )
    }
    return (
      <NotificationDrawer>
        <NotificationDrawerHeader
          title={t("notifications")}
          count={this.props.unreadCount}
          onClose={this.props.onNotificationClose}
        >
          <Dropdown
            onSelect={this.toggleOptions}
            toggle={<KebabToggle onToggle={this.toggleOptions} id="toggle-id-0" />}
            isOpen={this.state.isOptionsOpen}
            isPlain
            dropdownItems={this.notificationDrawerActions}
            id="notification-dropdown-0"
            position={DropdownPosition.right}
          />
        </NotificationDrawerHeader>
        <NotificationDrawerBody>{elements}</NotificationDrawerBody>
      </NotificationDrawer>
    )
  }
}

const mapStateToProps = (state) => ({
  items: state.entities.notification.items,
  unreadCount: state.entities.notification.unreadCount,
})

const mapDispatchToProps = (dispatch) => ({
  clearAll: () => dispatch(notificationClearAll()),
  markAllRead: () => dispatch(notificationMarkAllRead()),
  markAsRead: (id) => dispatch(notificationMarkAsRead({ id })),
  onNotificationClose: () => dispatch(notificationDrawerToggle()),
})

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(NotificationContainer))
