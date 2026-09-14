import { createSlice } from "@reduxjs/toolkit"
import { URL_DOCUMENTATION } from "../../Constants/Common"

const slice = createSlice({
  name: "about",
  initialState: {
    show: false,
    documentationUrl: URL_DOCUMENTATION,
    metricsDBDisabled: false,
  },
  reducers: {
    aboutShow: (about, action) => {
      about.show = true
    },

    aboutHide: (about, action) => {
      about.show = false
    },

    updateDocumentationUrl: (about, action) => {
      about.documentationUrl = action.payload.documentationUrl
    },

    updateMetricsDBStatus: (about, action) => {
      about.metricsDBDisabled = action.payload.metricsDBDisabled
    },
  },
})

export default slice.reducer

export const { aboutShow, aboutHide, updateDocumentationUrl, updateMetricsDBStatus } = slice.actions
