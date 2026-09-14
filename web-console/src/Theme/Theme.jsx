import React from "react"
import { connect } from "react-redux"
import { DEFAULT_THEME } from "../Constants/Common"
import { api } from "../Service/Api"
import { getValue } from "../Util/Util"

// default Colors
const defaultTheme = {
  "--mc-background-color": "#fff",
  "--mc-components-background-color": "#fff",
  "--mc-dashboard-background-color": "#f7f7f7",
  "--mc-widget-background-color": "#fff",
  "--mc-widget-title-background-color": "#f6f6f6",
  "--mc-widget-utilization-title-color": "var(--pf-global--palette--black-600)",
  "--mc-widget-utilization-value-color": "var(--pf-global--palette--black-800)",
  "--mc-widget-utilization-unit-color": "var(--pf-global--palette--black-600)",
  "--mc-widget-utilization-function-color": "var(--pf-global--palette--black-400)",
  "--mc-widget-title-color": "#434343",
  "--mc-timestamp-color": "var(--pf-global--palette--black-600)",
}

class Theme extends React.Component {
  state = {
    themeSettings: {},
  }

  componentDidMount() {
    this.loadThemeData(this.props.themeSelected)
  }

  componentDidUpdate(prevProps) {
    if (prevProps.themeSelected !== this.props.themeSelected) {
      this.loadThemeData(this.props.themeSelected)
    }
  }

  loadThemeData = (themeSelected = DEFAULT_THEME) => {
    if (themeSelected === DEFAULT_THEME) {
      this.setState({ themeSettings: defaultTheme })
    } else {
      api.dataRepository
        .get(themeSelected)
        .then((res) => {
          const themeData = getValue(res, "data.data", defaultTheme)
          this.setState({ themeSettings: themeData })
        })
        .catch((_e) => {
          this.setState({ themeSettings: defaultTheme })
        })
    }
  }

  updateTheme = (themeSettings = {}) => {
    const settings = { ...defaultTheme, ...themeSettings }
    const keys = Object.keys(settings)
    keys.map((key) => {
      document.documentElement.style.setProperty(key, settings[key])
    })
  }

  render() {
    const { themeSettings } = this.state
    this.updateTheme(themeSettings)
    return null
  }
}

const mapStateToProps = (state) => ({
  themeSelected: state.entities.theme.selection,
})

export default connect(mapStateToProps, null)(Theme)
