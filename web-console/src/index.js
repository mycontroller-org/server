import "@patternfly/react-core/dist/styles/base.css"
import React from "react"
import ReactDOM from "react-dom"
import "react-grid-layout/css/styles.css"
import "react-resizable/css/styles.css"
import { HashRouter as Router } from "react-router-dom"
import App from "./App"
import "./i18n/i18n"
import "./index.scss"
import * as serviceWorker from "./serviceWorker"

ReactDOM.render(
  //<React.StrictMode>
  <Router>
    <App />
  </Router>,
  //</React.StrictMode>,
  document.getElementById("root")
)

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister()
