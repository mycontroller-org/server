import { createSlice } from "@reduxjs/toolkit"

let lastId = 0

const slice = createSlice({
  name: "toaster",
  initialState: {
    items: [],
  },
  reducers: {
    toasterAdd: (toaster, action) => {
      const { type, title, description } = action.payload
      toaster.items.push({
        id: ++lastId,
        type,
        title,
        description,
        timestamp: Date.now(),
      })
    },

    toasterRemove: (toaster, action) => {
      toaster.items = toaster.items.filter((a) => a.id !== action.payload.id)
    },
  },
})

export default slice.reducer

export const { toasterAdd, toasterRemove } = slice.actions
