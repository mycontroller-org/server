import { createSlice } from "@reduxjs/toolkit"
import { DEFAULT_LANGUAGE } from "../../Constants/Common"

const slice = createSlice({
  name: "locale",
  initialState: {
    language: DEFAULT_LANGUAGE,
  },
  reducers: {
    updateLocale: (locale, action) => {
      locale.language = action.payload.language
    },
  },
})

export default slice.reducer

export const { updateLocale } = slice.actions
