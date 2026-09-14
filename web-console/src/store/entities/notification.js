import { createSlice } from "@reduxjs/toolkit"
import { NOTIFICATION_LIMIT, ALERT_TYPE_ERROR } from "../../Constants/Common"
import { getRandomId } from "../../Util/Util"

const DISPLAY_VARIANT_READ = "read"
const DISPLAY_VARIANT_UNREAD = "unread"
const DISPLAY_VARIANT_ATTENTION = "attention"

const slice = createSlice({
  name: "notification",
  initialState: {
    isDrawerExpanded: false,
    displayVariant: DISPLAY_VARIANT_READ, // can be one of: read, unread, attention
    unreadCount: 0,
    items: [],
  },
  reducers: {
    notificationAdd: (state, action) => {
      const { type, title, description } = action.payload
      state.items.push({
        id: getRandomId(5),
        type,
        title,
        description,
        unread: true,
        timestamp: Date.now(),
      })
      state.unreadCount++
      if (state.items.length > NOTIFICATION_LIMIT) {
        state.items = state.items.slice(-NOTIFICATION_LIMIT)
      }
      if (state.unreadCount > NOTIFICATION_LIMIT) {
        state.unreadCount = NOTIFICATION_LIMIT
      }
      for (let index = 0; index < state.items.length; index++) {
        const item = state.items[index]
        if (item.unread) {
          if (item.type === ALERT_TYPE_ERROR) {
            state.displayVariant = DISPLAY_VARIANT_ATTENTION
            break
          } else {
            state.displayVariant = DISPLAY_VARIANT_UNREAD
          }
        }
      }
    },

    notificationMarkAsRead: (state, action) => {
      let unreadCount = 0
      let displayVariant = DISPLAY_VARIANT_READ
      for (let index = 0; index < state.items.length; index++) {
        const item = state.items[index]
        if (item.id === action.payload.id) {
          item.unread = false
        }
        if (item.unread) {
          unreadCount++
          if (displayVariant !== DISPLAY_VARIANT_ATTENTION) {
            if (item.type === ALERT_TYPE_ERROR) {
              displayVariant = DISPLAY_VARIANT_ATTENTION
            } else {
              displayVariant = DISPLAY_VARIANT_UNREAD
            }
          }
        }
      }
      state.unreadCount = unreadCount
      state.displayVariant = displayVariant
    },

    notificationMarkAllRead: (state, action) => {
      state.items.forEach((n) => {
        n.unread = false
      })
      state.unreadCount = 0
      state.displayVariant = DISPLAY_VARIANT_READ
    },

    notificationClearAll: (state, action) => {
      state.items = []
      state.unreadCount = 0
      state.displayVariant = DISPLAY_VARIANT_READ
    },

    notificationDrawerToggle: (state, action) => {
      state.isDrawerExpanded = !state.isDrawerExpanded
    },
  },
})

export default slice.reducer

export const {
  notificationAdd,
  notificationMarkAsRead,
  notificationMarkAllRead,
  notificationClearAll,
  notificationDrawerToggle,
} = slice.actions
