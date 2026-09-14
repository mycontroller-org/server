import { createSlice } from "@reduxjs/toolkit"

const slice = createSlice({
  name: "spinner",
  initialState: {
    show: false,
  },
  reducers: {
    spinnerShow: (spinner, action) => {
      spinner.show = true
    },

    spinnerHide: (spinner, action) => {
      spinner.show = false
    },
  },
})

export default slice.reducer

export const { spinnerShow, spinnerHide } = slice.actions
