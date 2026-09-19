import React from "react"
import { connect } from "react-redux"
import { THEME_SYSTEM, normalizeTheme } from "../Constants/Common"
import { updateTheme } from "../store/entities/theme"
import { applyThemeClass, loadStoredTheme, saveStoredTheme } from "./themeStorage"

class Theme extends React.Component {
  mediaQuery = null

  componentDidMount() {
    const stored = loadStoredTheme()
    const normalized = stored || normalizeTheme(this.props.themeSelected)
    if (normalized !== this.props.themeSelected) {
      this.props.updateTheme({ theme: normalized })
    }
    saveStoredTheme(normalized)
    this.applySelection(normalized)
    this.mediaQuery = window.matchMedia("(prefers-color-scheme: dark)")
    if (this.mediaQuery.addEventListener) {
      this.mediaQuery.addEventListener("change", this.onSystemChange)
    } else if (this.mediaQuery.addListener) {
      this.mediaQuery.addListener(this.onSystemChange)
    }
  }

  componentDidUpdate(prevProps) {
    if (prevProps.themeSelected !== this.props.themeSelected) {
      this.applySelection(this.props.themeSelected)
    }
  }

  componentWillUnmount() {
    if (!this.mediaQuery) {
      return
    }
    if (this.mediaQuery.removeEventListener) {
      this.mediaQuery.removeEventListener("change", this.onSystemChange)
    } else if (this.mediaQuery.removeListener) {
      this.mediaQuery.removeListener(this.onSystemChange)
    }
  }

  onSystemChange = () => {
    if (normalizeTheme(this.props.themeSelected) === THEME_SYSTEM) {
      applyThemeClass(THEME_SYSTEM)
    }
  }

  applySelection = (selection) => {
    const normalized = normalizeTheme(selection)
    saveStoredTheme(normalized)
    applyThemeClass(normalized)
  }

  render() {
    return null
  }
}

const mapStateToProps = (state) => ({
  themeSelected: state.entities.theme.selection,
})

export default connect(mapStateToProps, { updateTheme })(Theme)
