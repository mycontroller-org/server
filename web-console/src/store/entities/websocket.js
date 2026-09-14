import { createSlice } from "@reduxjs/toolkit"
import { getValue } from "../../Util/Util"

const slice = createSlice({
  name: "websocket",
  initialState: {
    subscription: {},
    data: {},
    connected: false,
    message: "",
  },
  reducers: {
    subscribe: (state, action) => {
      const { subscribe = [] } = action.payload
      state.subscription = subscribe
    },

    unsubscribe: (state, action) => {
      const { unsubscribe = [] } = action.payload
      if (unsubscribe.length === 0) {
        state.subscription = []
      } else {
        // TODO: remove specific subscriptions
      }
    },

    loadData: (state, action) => {
      const { key, resources = [] } = action.payload
      let data = { ...state.data }
      data[key] = resources
      state.data = data
    },

    unloadData: (state, action) => {
      const { key } = action.payload
      let data = { ...state.data }
      delete data[key]
      state.data = data
    },

    updateEvent: (state, action) => {
      const { response } = action.payload
      let data = { ...state.data }
      state.data = updateItem(data, response)
    },
    connected: (state, _action) => {
      state.connected = true
    },
    disconnected: (state, action) => {
      const { message } = action.payload
      state.connected = false
      state.message = message
    },
  },
})

export default slice.reducer

export const { connected, disconnected, subscribe, unsubscribe, loadData, unloadData, updateEvent } =
  slice.actions

// helpers
const updateItem = (data = {}, responseString = "{}") => {
  // console.log("received:", responseString)
  const keys = Object.keys(data)
  const response = JSON.parse(responseString)
  const event = response.data
  if (response.type !== "event" || event === undefined) {
    return data
  }

  const quickId = event.entityQuickId
  const entity = event.entity

  keys.forEach((key) => {
    const items = getValue(data, key, {})
    if (items[quickId] !== undefined) {
      //console.log("updated:", responseString)
      items[quickId] = entity
      data[key] = items
    }
  })
  return data
}
