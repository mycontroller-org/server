import { createSlice } from "@reduxjs/toolkit"
import { DEFAULT_THEME } from "../../Constants/Common"

const slice = createSlice({
  name: "theme",
  initialState: {
    selection: DEFAULT_THEME,
  },
  reducers: {
    updateTheme: (theme, action) => {
      theme.selection = action.payload.theme
    },
  },
})

export default slice.reducer

export const { updateTheme } = slice.actions
