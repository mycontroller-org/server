import {
  ListItem,
  LoginFooterItem,
  LoginForm,
  LoginMainFooterBandItem,
  LoginPage,
} from "@patternfly/react-core"
import { ExclamationCircleIcon } from "@patternfly/react-icons"
import htmlParser from "html-react-parser"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import {
  DEFAULT_LANGUAGE,
  MAXIMUM_LOGIN_EXPIRES_IN_DAYS,
  MINIMUM_LOGIN_EXPIRES_IN_DAYS,
  URL_DOCUMENTATION,
} from "../Constants/Common"
import logoBackground from "../Logo/mc-black-login-page.svg"
import displayLogo from "../Logo/mc-white-full.svg"
import { api } from "../Service/Api"
import { updateDocumentationUrl, updateMetricsDBStatus } from "../store/entities/about"
import { authSuccess } from "../store/entities/auth"
import { updateLocale } from "../store/entities/locale"
import { getValue } from "../Util/Util"
import "./Login.scss"
import { notificationClearAll } from "../store/entities/notification"

class SimpleLoginPage extends React.Component {
  state = {
    showHelperText: false,
    usernameValue: "",
    isValidUsername: true,
    passwordValue: "",
    isValidPassword: true,
    isRememberMeChecked: false,
    isLoginButtonDisabled: false,
    loginData: { message: `${this.props.t("fetching")}...`, serverMessage: "" },
  }

  componentDidMount() {
    const { t, updateMetricsDB, updateDocUrl, updateLocale } = this.props
    api.status
      .get()
      .then((res) => {
        const { documentationUrl, language, metricsDBDisabled } = res.data
        // update documentation url
        updateDocUrl({
          documentationUrl: documentationUrl !== "" ? documentationUrl : URL_DOCUMENTATION,
        })
        updateMetricsDB({ metricsDBDisabled: metricsDBDisabled })

        // update locale
        const lng = language && language !== "" ? language : DEFAULT_LANGUAGE
        updateLocale({ language: lng })

        const loginData = getValue(res.data, "login", { message: t("no_login_message") })
        this.setState({ loginData: loginData })
      })
      .catch((_e) => {
        this.setState({ loginData: { message: t("error.console.login_msg_loading_failed") } })
      })
  }

  handleUsernameChange = (value) => {
    this.setState({ usernameValue: value })
  }

  handlePasswordChange = (passwordValue) => {
    this.setState({ passwordValue })
  }

  onRememberMeClick = () => {
    this.setState({ isRememberMeChecked: !this.state.isRememberMeChecked })
  }

  onLoginButtonClick = (event) => {
    event.preventDefault()
    this.setState({ isLoginButtonDisabled: true })

    const { usernameValue, passwordValue, isRememberMeChecked } = this.state
    const expiresIn = isRememberMeChecked ? MAXIMUM_LOGIN_EXPIRES_IN_DAYS : MINIMUM_LOGIN_EXPIRES_IN_DAYS
    const loginData = {
      username: usernameValue,
      password: passwordValue,
      expiresIn: expiresIn,
    }
    api.auth
      .login(loginData)
      .then((res) => {
        const user = { ...res.data }
        // when login success, clears all the existing notifications
        this.props.clearAllNotifications()
        // and moves into index page
        this.props.updateSuccessLogin(user)
      })
      .catch((e) => {
        this.setState({
          isValidUsername: false,
          isValidPassword: false,
          showHelperText: true,
          isLoginButtonDisabled: false,
        })
      })
  }

  render() {
    const { message: loginMessage, serverMessage } = this.state.loginData
    const { t } = this.props
    let loginText = loginMessage
    if (serverMessage !== undefined && serverMessage !== "") {
      loginText = `
        <p>${loginMessage}</p><br>
        <p><b>${t("system_message")}</b></p>
        <p>${serverMessage}</p>
      `
    }
    loginText = htmlParser(loginText)

    const helperText = (
      <React.Fragment>
        <ExclamationCircleIcon />
        &nbsp;{t("error.console.invalid_login_credentials")}
      </React.Fragment>
    )

    const forgotCredentials = (
      <LoginMainFooterBandItem>
        <a href="#">{t("forgot_username_or_password")}</a>
      </LoginMainFooterBandItem>
    )

    const listItem = (
      <React.Fragment>
        <ListItem>
          <LoginFooterItem href={this.props.documentationUrl}>{t("documentation")}</LoginFooterItem>
        </ListItem>
      </React.Fragment>
    )

    const loginForm = (
      <LoginForm
        showHelperText={this.state.showHelperText}
        helperText={helperText}
        //helperTextIcon={<ExclamationCircleIcon />}
        usernameLabel={t("username")}
        usernameValue={this.state.usernameValue}
        onChangeUsername={this.handleUsernameChange}
        isValidUsername={this.state.isValidUsername}
        passwordLabel={t("password")}
        passwordValue={this.state.passwordValue}
        onChangePassword={this.handlePasswordChange}
        isValidPassword={this.state.isValidPassword}
        rememberMeLabel={t("keep_me_logged_in_30_days")}
        isRememberMeChecked={this.state.isRememberMeChecked}
        onChangeRememberMe={this.onRememberMeClick}
        onLoginButtonClick={this.onLoginButtonClick}
        isLoginButtonDisabled={this.state.isLoginButtonDisabled}
        loginButtonLabel={t("log_in")}
      />
    )

    return (
      <LoginPage
        footerListVariants="inline"
        brandImgSrc={displayLogo}
        brandImgAlt="MyController.org"
        className="mc-login"
        backgroundImgSrc={logoBackground}
        backgroundImgAlt="Images"
        footerListItems={listItem}
        textContent={loginText}
        loginTitle={t("welcome_to_mycontroller")}
        loginSubtitle={t("login_to_your_account")}
        // forgotCredentials={forgotCredentials}
      >
        {loginForm}
      </LoginPage>
    )
  }
}

const mapStateToProps = (state) => ({
  documentationUrl: state.entities.about.documentationUrl,
})

const mapDispatchToProps = (dispatch) => ({
  clearAllNotifications: () => dispatch(notificationClearAll()),
  updateSuccessLogin: (data) => dispatch(authSuccess(data)),
  updateDocUrl: (data) => dispatch(updateDocumentationUrl(data)),
  updateMetricsDB: (data) => dispatch(updateMetricsDBStatus(data)),
  updateLocale: (data) => dispatch(updateLocale(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(SimpleLoginPage))
