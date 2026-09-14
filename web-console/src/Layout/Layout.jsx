import {
  Badge,
  Banner,
  Brand,
  Button,
  Divider,
  Dropdown,
  DropdownGroup,
  DropdownItem,
  DropdownSeparator,
  DropdownToggle,
  Nav,
  NavExpandable,
  NavItem,
  NavItemSeparator,
  NavList,
  NotificationBadge,
  Page,
  PageHeader,
  PageHeaderTools,
  PageHeaderToolsGroup,
  PageHeaderToolsItem,
  PageSidebar,
  Popover,
  SkipToContent,
  Text,
  TextContent,
  TextVariants,
  Tooltip,
} from "@patternfly/react-core"
// make sure you've installed @patternfly/patternfly
//import accessibleStyles from "@patternfly/react-styles/css/utilities/Accessibility/accessibility";
//import spacingStyles from "@patternfly/react-styles/css/utilities/Spacing/spacing";
//import { css } from "@patternfly/react-styles";
import {
  BellIcon,
  BookIcon,
  ChatIcon,
  CogIcon,
  ExternalLinkAltIcon,
  ExternalLinkSquareAltIcon,
  GithubIcon,
  HelpIcon,
  InfoAltIcon,
  LanguageIcon,
  PaletteIcon,
  PowerOffIcon,
  UnpluggedIcon,
  UserIcon,
} from "@patternfly/react-icons"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import { withRouter } from "react-router"
import { Redirect, Route, Switch } from "react-router-dom"
import ErrorBoundary from "../Components/ErrorBoundary/ErrorBoundary"
import { HeaderSpinner } from "../Components/Spinner/Spinner"
import Toaster from "../Components/Toaster/Toaster"
import { URL_FORUM, URL_SOURCE_CODE } from "../Constants/Common"
import { languages } from "../i18n/languages"
import logoMain from "../Logo/mc-white.svg"
import AboutPage from "../Pages/About/About"
import { api } from "../Service/Api"
import { hiddenRoutes, redirect as r, routeMap as rMap, routes } from "../Service/Routes"
import { wsConnect, wsDisconnect } from "../Service/Websocket"
import { aboutShow } from "../store/entities/about"
import { clearAuth } from "../store/entities/auth"
import { updateLocale } from "../store/entities/locale"
import { updateTheme } from "../store/entities/theme"
import { notificationDrawerToggle } from "../store/entities/notification"
import Theme from "../Theme/Theme"
//import imgAvatar from "./imgAvatar.svg";
import "./Layout.scss"
import NotificationContainer from "./NotificationContainer"

class PageLayoutExpandableNav extends React.Component {
  state = {
    isUserDropdownOpen: false,
    isHelpDropdownOpen: false,
    isLanguageDropdownOpen: false,
    isThemeDropdownOpen: false,
    activeGroup: "grp-1",
    activeItem: "grp-1_itm-1",
    themes: [],
  }

  componentDidMount() {
    wsConnect()
    // load themes
    this.fetchThemes()
  }

  fetchThemes = () => {
    const defaultThemes = [{ name: "default", description: "default theme" }]
    const _page = {
      limit: 15, // limits 15 entries
      offset: 0,
      filter: [{ k: "labels.gui_theme", o: "eq", v: "true" }],
      // sort: { f: "id", o: "asc" },
    }
    api.dataRepository
      .list(_page)
      .then((res) => {
        const themes = res.data.data.map((theme) => {
          return { name: theme.id, description: theme.description }
        })
        this.setState({ themes: [...defaultThemes, ...themes] })
      })
      .catch((_e) => {
        this.setState({ themes: defaultThemes })
      })
  }

  componentWillUnmount() {
    wsDisconnect()
  }

  onUserDropdownToggle = (isUserDropdownOpen) => {
    this.setState({
      isUserDropdownOpen,
    })
  }

  onUserDropdownSelect = (_event) => {
    this.setState({
      isUserDropdownOpen: !this.state.isUserDropdownOpen,
    })
  }

  onHelpDropdownToggle = (isHelpDropdownOpen) => {
    this.setState({
      isHelpDropdownOpen,
    })
  }

  onHelpDropdownSelect = (_event) => {
    this.setState({
      isHelpDropdownOpen: !this.state.isHelpDropdownOpen,
    })
  }

  onLanguageDropdownToggle = (isLanguageDropdownOpen) => {
    this.setState({
      isLanguageDropdownOpen,
    })
  }

  onLanguageDropdownSelect = (_event) => {
    this.setState({
      isLanguageDropdownOpen: !this.state.isLanguageDropdownOpen,
    })
  }

  onThemeDropdownToggle = (isThemeDropdownOpen) => {
    this.setState({
      isThemeDropdownOpen,
    })
  }

  onThemeDropdownSelect = (_event) => {
    this.setState({
      isThemeDropdownOpen: !this.state.isThemeDropdownOpen,
    })
  }

  onNavSelect = (result) => {
    this.setState({
      activeItem: result.itemId,
      activeGroup: result.groupId,
    })
  }

  onMenuSelect = (data) => {
    //console.log(data);
    //console.log(item, key, keyPath, domEvent)
    //this.setState({
    //selectedMenuKey: data.key
    //});
    const { history } = this.props
    history.push(data.to)
  }

  renderContent = () => {
    const allRoutes = []
    routes.forEach((item) => {
      if (item.children && item.children.length > 0) {
        item.children.forEach((sItem) => {
          if (!sItem.isSeparator) {
            allRoutes.push(<Route key={sItem.to} exact path={sItem.to} component={sItem.component} />)
          }
        })
      } else if (!item.isSeparator) {
        allRoutes.push(<Route key={item.to} exact path={item.to} component={item.component} />)
      }
    })
    hiddenRoutes.forEach((item) => {
      allRoutes.push(<Route key={item.to} exact path={item.to} component={item.component} />)
    })
    return (
      <Switch>
        {allRoutes}
        <Redirect from="*" to="/" key="default-route" />
      </Switch>
    )
  }

  callLogout = () => {
    api.auth.logout().then((_res) => {
      this.props.doLogout()
    })
  }

  render() {
    const {
      location,
      t,
      languageSelected,
      themeSelected,
      websocketConnected,
      websocketMessage,
      metricsDBDisabled,
    } = this.props
    const { themes } = this.state

    // selected menu
    let menuSelection = ""
    for (let rIndex = 0; rIndex < routes.length; rIndex++) {
      const r = routes[rIndex]
      if (r.children && r.children.length > 0) {
        for (let cIndex = 0; cIndex < r.children.length; cIndex++) {
          const sr = r.children[cIndex]
          if (location.pathname.startsWith(sr.to)) {
            menuSelection = sr.id
            break
          }
        }
      } else if (location.pathname.startsWith(r.to)) {
        menuSelection = r.id
        break
      }
      if (menuSelection !== "") {
        break
      }
    }

    const getSubMenu = (sm) => {
      const sMenus = sm.map((m) => {
        if (m.isSeparator) {
          return <NavItemSeparator key={m.id} />
        } else {
          return (
            <NavItem
              key={m.id}
              itemId={m.id}
              to={m.to}
              isActive={menuSelection === m.id}
              preventDefault={true}
            >
              {t(m.title)}
              {m.isExperimental ? (
                <span className="mc-badge-experimental">
                  <Badge>{t("experimental")}</Badge>
                </span>
              ) : null}
            </NavItem>
          )
        }
      })
      return sMenus
    }

    const PageNav = (
      <Nav onSelect={this.onMenuSelect} aria-label="Nav" theme="dark">
        <NavList>
          {routes.map((m) => {
            if (m.children && m.children.length > 0) {
              return (
                <NavExpandable key={m.id} groupId={m.id} title={t(m.title)}>
                  {getSubMenu(m.children)}
                </NavExpandable>
              )
            } else {
              return (
                <NavItem
                  key={m.id}
                  itemId={m.id}
                  to={m.to}
                  isActive={menuSelection === m.id}
                  preventDefault={true}
                >
                  {t(m.title)}
                  {m.isExperimental ? (
                    <span className="mc-badge-experimental">
                      <Badge>{t("experimental")}</Badge>
                    </span>
                  ) : null}
                </NavItem>
              )
            }
          })}
        </NavList>
      </Nav>
    )

    const notificationDrawer = <NotificationContainer />

    const helpDropdownItems = [
      <DropdownItem key="documentation" href={this.props.documentationUrl} target="_blank">
        <BookIcon /> {t("documentation")} <ExternalLinkAltIcon />
      </DropdownItem>,
      <DropdownItem key="forum" href={URL_FORUM} target="_blank">
        <ChatIcon /> {t("forum")} <ExternalLinkAltIcon />
      </DropdownItem>,
      <DropdownItem key="source-code" href={URL_SOURCE_CODE} target="_blank">
        <GithubIcon /> {t("source_code")} <ExternalLinkAltIcon />
      </DropdownItem>,

      <DropdownSeparator key="separator" />,
      <DropdownItem
        key="settings"
        onClick={() => {
          r(this.props.history, rMap.settings.system)
        }}
      >
        <CogIcon /> {t("settings")}
      </DropdownItem>,
      <DropdownItem key="about" onClick={this.props.showAbout}>
        <InfoAltIcon /> {t("about")}
      </DropdownItem>,
    ]

    const languageItems = languages.map((l) => {
      const lngSelected = l.lng === languageSelected ? "language_selected" : ""
      return (
        <DropdownItem
          key={`lang_${l.lng}`}
          className={`language_item ${lngSelected}`}
          onClick={() => this.props.updateLocale({ language: l.lng })}
        >
          <Tooltip position="left" content={l.country_code}>
            <span>{l.flag} </span>
          </Tooltip>
          {l.title}
        </DropdownItem>
      )
    })

    const themeDropdownItems = themes.map((theme) => {
      const thSelected = theme.name === themeSelected ? "language_selected" : ""
      return (
        <DropdownItem
          key={`theme_${theme.name}`}
          className={`language_item ${thSelected}`}
          onClick={() => this.props.updateTheme({ theme: theme.name })}
          description={theme.description}
        >
          {theme.name}
        </DropdownItem>
      )
    })

    const userDropdownItems = [
      <DropdownGroup key="group2">
        <DropdownItem
          key="group2_profile"
          onClick={() => {
            r(this.props.history, rMap.settings.profile)
          }}
        >
          <UserIcon /> {t("profile")}
        </DropdownItem>
        <DropdownItem key="group2_logout" onClick={this.callLogout}>
          <PowerOffIcon /> {t("logout")}
        </DropdownItem>
      </DropdownGroup>,
    ]

    // show websocket status if disconnected
    const websocketStatus = !websocketConnected ? (
      <Popover
        aria-label="websocket_disconnected"
        showClose={true}
        position="bottom"
        bodyContent={(_hide) => (
          <div>
            {t("websocket_disconnected")}
            <br />
            {websocketMessage}
          </div>
        )}
      >
        <Button variant="warning">
          <UnpluggedIcon />
        </Button>
      </Popover>
    ) : null

    const headerTools = (
      <PageHeaderTools>
        <PageHeaderToolsGroup key="spinner">
          {this.props.showGlobalSpinner ? <HeaderSpinner size="lg" /> : null}
        </PageHeaderToolsGroup>

        <PageHeaderToolsGroup key="others">
          <PageHeaderToolsItem key="web_socket">{websocketStatus}</PageHeaderToolsItem>
          <PageHeaderToolsItem visibility={{ default: "visible" }} isSelected={this.props.isDrawerExpanded}>
            <NotificationBadge
              variant={this.props.notificationDisplayVariant}
              onClick={this.props.onNotificationBadgeClick}
              aria-label={t("notifications")}
              count={this.props.notificationCount}
            >
              <BellIcon />
            </NotificationBadge>
          </PageHeaderToolsItem>
          <PageHeaderToolsItem>
            <Dropdown
              isPlain
              position="right"
              onSelect={this.onHelpDropdownSelect}
              toggle={
                <DropdownToggle
                  toggleIndicator={null}
                  onToggle={this.onHelpDropdownToggle}
                  icon={<HelpIcon />}
                />
              }
              isOpen={this.state.isHelpDropdownOpen}
              dropdownItems={helpDropdownItems}
            />
          </PageHeaderToolsItem>
          <PageHeaderToolsItem>
            <Dropdown
              isPlain
              position="right"
              onSelect={this.onLanguageDropdownSelect}
              isOpen={this.state.isLanguageDropdownOpen}
              toggle={
                <DropdownToggle
                  className="language_icon"
                  toggleIndicator={null}
                  onToggle={this.onLanguageDropdownToggle}
                  icon={<LanguageIcon size="md" />}
                />
              }
              dropdownItems={languageItems}
            />
          </PageHeaderToolsItem>

          <PageHeaderToolsItem>
            <Dropdown
              isPlain
              position="right"
              onSelect={this.onThemeDropdownSelect}
              isOpen={this.state.isThemeDropdownOpen}
              toggle={
                <DropdownToggle
                  toggleIndicator={null}
                  onToggle={this.onThemeDropdownToggle}
                  icon={<PaletteIcon size="sm" />}
                />
              }
              dropdownItems={themeDropdownItems}
            />
          </PageHeaderToolsItem>

          <PageHeaderToolsItem>
            <Dropdown
              isPlain
              position="right"
              onSelect={this.onUserDropdownSelect}
              isOpen={this.state.isUserDropdownOpen}
              toggle={
                <DropdownToggle onToggle={this.onUserDropdownToggle} icon={<UserIcon />}>
                  <span className="mc-mobile-view">{this.props.userDetail.fullName}</span>
                </DropdownToggle>
              }
              dropdownItems={userDropdownItems}
            />
          </PageHeaderToolsItem>
        </PageHeaderToolsGroup>
      </PageHeaderTools>
    )

    const Header = (
      <PageHeader
        //logo={<Brand src={imgBrand} alt="Patternfly Logo" />}
        logo={
          <>
            <Brand className="mc-header-logo" src={logoMain} alt="MyController" />
            <TextContent className="rb-logo-text mc-mobile-view">
              <Text component={TextVariants.h3}>
                MyController.org
                <div
                  style={{
                    marginTop: "-5px",
                    fontSize: "10px",
                    fontWeight: "100",
                  }}
                >
                  {t("the_open_source_controller")}
                </div>
              </Text>
            </TextContent>
          </>
        }
        headerTools={headerTools}
        //avatar={<Avatar src={imgAvatar} alt="Avatar image" />}
        showNavToggle
      />
    )
    const Sidebar = <PageSidebar nav={PageNav} theme="dark" />

    const pageId = "main-content-page-layout-expandable-nav"
    const PageSkipToContent = <SkipToContent href={`#${pageId}`}>Skip to content</SkipToContent>

    const banner = metricsDBDisabled ? (
      <Banner key="banner-metrics-database" variant="warning">
        <strong>{t("metrics_database_disabled")}</strong>
      </Banner>
    ) : null

    return (
      <React.Fragment>
        <Theme />
        <AboutPage />
        <Toaster />
        <Page
          onPageResize={() => {
            window.dispatchEvent(new Event("resize"))
          }}
          header={Header}
          sidebar={Sidebar}
          isManagedSidebar={true}
          skipToContent={PageSkipToContent}
          mainContainerId={pageId}
          notificationDrawer={notificationDrawer}
          isNotificationDrawerExpanded={this.props.isDrawerExpanded}
        >
          <ErrorBoundary hasMargin>
            {banner}
            {this.renderContent()}
          </ErrorBoundary>
        </Page>
      </React.Fragment>
    )
  }
}

const mapStateToProps = (state) => ({
  isDrawerExpanded: state.entities.notification.isDrawerExpanded,
  notificationDisplayVariant: state.entities.notification.displayVariant,
  notificationCount: state.entities.notification.unreadCount,
  showGlobalSpinner: state.entities.spinner.show,
  userDetail: state.entities.auth.user,
  documentationUrl: state.entities.about.documentationUrl,
  languageSelected: state.entities.locale.language,
  themeSelected: state.entities.theme.selection,
  websocketConnected: state.entities.websocket.connected,
  websocketMessage: state.entities.websocket.message,
  metricsDBDisabled: state.entities.about.metricsDBDisabled,
})

const mapDispatchToProps = (dispatch) => ({
  onNotificationBadgeClick: () => dispatch(notificationDrawerToggle()),
  showAbout: () => dispatch(aboutShow()),
  doLogout: () => dispatch(clearAuth()),
  updateLocale: (data) => dispatch(updateLocale(data)),
  updateTheme: (data) => dispatch(updateTheme(data)),
})

export default withRouter(
  connect(mapStateToProps, mapDispatchToProps)(withTranslation()(PageLayoutExpandableNav))
)
