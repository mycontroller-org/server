import {
  AboutModal,
  Divider,
  Text,
  TextContent,
  TextList,
  TextListItem,
  TextVariants,
} from "@patternfly/react-core"
import React from "react"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import { COPYRIGHT_YEAR } from "../../Constants/Common"
import brandImg from "../../Logo/mc-white-full.svg"
import { api } from "../../Service/Api"
import { aboutHide } from "../../store/entities/about"
import "./About.scss"
class AboutPage extends React.Component {
  state = {
    backend: {},
    loading: true,
  }

  componentDidMount() {
    api.version
      .get()
      .then((res) => {
        this.setState({ backend: res.data, loading: false })
      })
      .catch((e) => {
        this.setState({ loading: false })
      })
  }

  render() {
    const bk = this.state.backend
    const { t } = this.props
    return (
      <AboutModal
        className="mc-about"
        isOpen={this.props.showModel}
        onClose={this.props.hideAbout}
        trademark={`© ${COPYRIGHT_YEAR} MyController.org. ${t("all_rights_reserved")}. ${t("apache_license_2")}`}
        brandImageSrc={brandImg}
        brandImageAlt="MyController.org"
        //productName="MYCONTROLLER.ORG"
      >
        <TextContent>
          <Text component={TextVariants.h2}>{t("server")}</Text>
          <Divider component="hr" />
          <TextList component="dl">
            <TextListItem component="dt">{t("host_id")}</TextListItem>
            <TextListItem component="dd">{bk.hostId}</TextListItem>
            <TextListItem component="dt">{t("version")}</TextListItem>
            <TextListItem component="dd">{bk.version}</TextListItem>
            <TextListItem component="dt">{t("git_commit")}</TextListItem>
            <TextListItem component="dd">{bk.gitCommit}</TextListItem>
            <TextListItem component="dt">{t("build_date")}</TextListItem>
            <TextListItem component="dd">{bk.buildDate}</TextListItem>
            <TextListItem component="dt">{t("golang_version")}</TextListItem>
            <TextListItem component="dd">
              {bk.goVersion} ({bk.platform}, {bk.arch})
            </TextListItem>
          </TextList>

          <Text component={TextVariants.h2}>{t("web_console")}</Text>
          <Divider component="hr" />
          <TextList component="dl">
            <TextListItem component="dt">{t("version")}</TextListItem>
            <TextListItem component="dd">{process.env.REACT_APP_VERSION}</TextListItem>
            <TextListItem component="dt">{t("git_commit")}</TextListItem>
            <TextListItem component="dd">{process.env.REACT_APP_GIT_SHA}</TextListItem>
            <TextListItem component="dt">{t("build_date")}</TextListItem>
            <TextListItem component="dd">{process.env.REACT_APP_BUILD_DATE}</TextListItem>
            <TextListItem component="dt">{t("react_version")}</TextListItem>
            <TextListItem component="dd">{React.version}</TextListItem>
          </TextList>
        </TextContent>
      </AboutModal>
    )
  }
}

const mapStateToProps = (state) => ({
  showModel: state.entities.about.show,
})

const mapDispatchToProps = (dispatch) => ({
  hideAbout: () => dispatch(aboutHide()),
})

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(AboutPage))
