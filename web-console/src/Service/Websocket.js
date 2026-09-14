import { api } from "../Service/Api"
import { w3cwebsocket as W3CWebSocket } from "websocket"
import { store } from "../store/persister"
import { connected, disconnected, updateEvent } from "../store/entities/websocket"

let wsClient = null
let wsUrl = null

const reconnectDelay = 5000 // 5 seconds

const getWebsocketUrl = () => {
  // workaround to websocket proxy on development mode
  // call websocket http url once
  if (process.env.REACT_APP_IS_DEV_ENV === "true" && wsUrl === null) {
    api.websocket.get().catch((_e) => {})
  }

  wsUrl = new URL("/api/ws", window.location.href)
  wsUrl.protocol = wsUrl.protocol.replace("http", "ws")

  // do not include token in request, now it is managed by cookie
  // add authorization token
  // let token = ""
  // const auth = store.getState().entities.auth
  // if (auth.authenticated) {
  //   token = auth.token
  // }
  // return `${wsUrl.href}?access_token=${token}`

  return wsUrl.href
}

export const wsConnect = () => {
  wsClient = new W3CWebSocket(getWebsocketUrl())

  wsClient.onopen = () => {
    store.dispatch(connected())
    // console.log("WebSocket Client Connected")
  }

  wsClient.onmessage = (message) => {
    store.dispatch(updateEvent({ response: message.data }))
  }

  wsClient.onclose = (event) => {
    store.dispatch(disconnected({ message: event.reason }))
    // console.log("websocket is closed", event.reason)
    // if (wsClient !== null) {
    //   console.log("websocket reconnect will be attempted in 5 seconds")
    // }
    setTimeout(function () {
      if (wsClient != null) {
        wsConnect()
      }
    }, reconnectDelay)
  }

  wsClient.onerror = (err) => {
    store.dispatch(disconnected({ message: err.message }))
    // console.error("websocket encountered error: ", err.message, "closing socket")
    if (wsClient !== null) {
      wsClient.close()
    }
  }
}

export const wsDisconnect = () => {
  if (wsClient != null) {
    const _wsClient = wsClient
    wsClient = null
    if (_wsClient !== null) {
      _wsClient.close()
    }
  }
}
