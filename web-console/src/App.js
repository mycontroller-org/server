import React, { Suspense } from "react"
import { Provider } from "react-redux"
import { PersistGate } from "redux-persist/integration/react"
import IndexPage from "./Layout/IndexPage"
import { persistor, store } from "./store/persister"
import Theme from "./Theme/Theme"

function App() {
  return (
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <Theme />
        <Suspense fallback={<div>Loading...</div>}>
          <IndexPage />
        </Suspense>
      </PersistGate>
    </Provider>
  )
}

export default App
