import { createSlice } from "@reduxjs/toolkit"

const slice = createSlice({
  name: "dashboard",
  initialState: {
    selectionId: "",
  },
  reducers: {
    updateSelectionId: (state, action) => {
      state.selectionId = action.payload
    },
  },
})

export default slice.reducer

export const { updateSelectionId } = slice.actions
