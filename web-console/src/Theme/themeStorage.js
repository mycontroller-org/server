import { DEFAULT_THEME, THEME_DARK, normalizeTheme, resolveTheme } from "../Constants/Common"

export const THEME_STORAGE_KEY = "mc-theme"

export const loadStoredTheme = () => {
  try {
    const raw = window.localStorage.getItem(THEME_STORAGE_KEY)
    if (raw == null || raw === "") {
      return null
    }
    return normalizeTheme(raw)
  } catch (_e) {
    return null
  }
}

export const saveStoredTheme = (theme) => {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, normalizeTheme(theme))
  } catch (_e) {
    // private mode / quota
  }
}

export const applyThemeClass = (selection) => {
  document.documentElement.classList.toggle("pf-theme-dark", resolveTheme(selection) === THEME_DARK)
}

export const initialTheme = () => loadStoredTheme() || DEFAULT_THEME
