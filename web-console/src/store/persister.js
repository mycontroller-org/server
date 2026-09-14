import { createStore } from "redux"
// import logger from "redux-logger"
import { persistReducer, persistStore } from "redux-persist"
import storage from "redux-persist/lib/storage"
// import thunk from "redux-thunk"
import rootReducer from "./reducer"

const persistConfig = {
  key: "root",
  storage: storage,
  //whitelist: ["entities.auth"], // which reducer want to store
}
const pReducer = persistReducer(persistConfig, rootReducer)
//const middleware = applyMiddleware(thunk, logger)
//const store = createStore(pReducer, middleware)
const store = createStore(pReducer)
const persistor = persistStore(store)
export { persistor, store }
