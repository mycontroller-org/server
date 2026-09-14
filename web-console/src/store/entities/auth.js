import { createSlice } from "@reduxjs/toolkit"

const slice = createSlice({
  name: "auth",
  initialState: {
    authenticated: false,
    token: "",
    user: {},
  },
  reducers: {
    authSuccess: (state, action) => {
      state.authenticated = true
      state.user = {
        ...action.payload,
      }
      state.token = "" // do not keep token here, now managed via cookie, remove this field sometime later
    },
    clearAuth: (state, action) => {
      state.authenticated = false
      state.token = ""
      state.user = {}
    },
    updateUser: (state, action) => {
      state.user = {
        ...action.payload,
      }
      state.token = ""
    },
  },
})

export default slice.reducer

export const { authSuccess, clearAuth, updateUser } = slice.actions
