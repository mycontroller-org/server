import { createSlice } from "@reduxjs/toolkit"
import { normalizeTheme } from "../../Constants/Common"
import { initialTheme } from "../../Theme/themeStorage"

const slice = createSlice({
  name: "theme",
  initialState: {
    selection: initialTheme(),
  },
  reducers: {
    updateTheme: (theme, action) => {
      theme.selection = normalizeTheme(action.payload.theme)
    },
  },
})

export default slice.reducer

export const { updateTheme } = slice.actions
