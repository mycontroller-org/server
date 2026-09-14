import React, { Suspense } from "react"
import { Provider } from "react-redux"
import { PersistGate } from "redux-persist/integration/react"
import Moment from "react-moment"
import IndexPage from "./Layout/IndexPage"
//import store from "./store/configureStore"
import { persistor, store } from "./store/persister"

// https://github.com/headzoo/react-moment#pooled-timer
Moment.startPooledTimer(30000)

function App() {
  return (
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <Suspense fallback={<div>Loading...</div>}>
          <IndexPage />
        </Suspense>
      </PersistGate>
    </Provider>
  )
}

export default App
