import axios from "axios"
import qs from "qs"
import { SECURE_SHARE_PREFIX } from "../Constants/Common"
import { clearAuth } from "../store/entities/auth"
import { spinnerHide, spinnerShow } from "../store/entities/globalSpinner"
import { notificationAdd } from "../store/entities/notification"
import { toasterAdd } from "../store/entities/toaster"
//import store from "../store/configureStore"
import { store } from "../store/persister"

export const HTTP_CODES = {
  OK: 200,
  ACCEPTED: 202,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  NOT_FOUND: 404,
  REQUEST_FAILED: 422,
  INTERNAL_SERVER: 500,
  SERVICE_UNAVAILABLE: 503,
  GATEWAY_TIMEOUT: 504,
}

export const HTTP_VERBS = {
  DELETE: "delete",
  GET: "get",
  PATCH: "patch",
  POST: "post",
  PUT: "put",
}

const myAxios = axios.create({
  paramsSerializer: (params) => qs.stringify(params, { arrayFormat: "repeat" }),
})

// Add a request interceptor
myAxios.interceptors.request.use(
  (request) => {
    // console.log("request:", request)
    store.dispatch(spinnerShow())
    return request
  },
  (error) => {
    store.dispatch(spinnerHide())
    console.log("REQ-Error:", error)
    return Promise.reject(error)
  }
)

// Add a response interceptor
myAxios.interceptors.response.use(
  (response) => {
    //console.log('Response:', response)
    store.dispatch(spinnerHide())
    return response
  },
  (error) => {
    const alert = {
      type: "danger",
      title: error.message,
      description: ["URL: " + error.config.url],
    }
    if (error.response && error.response.data) {
      alert.description.push("Message: " + JSON.stringify(error.response.data))
    }
    // update redux, if authenticated
    if (store.getState().entities.auth.authenticated) {
      // if unauthorized clear auth data
      if (error.response.status === 401) {
        store.dispatch(clearAuth())
      }
      store.dispatch(spinnerHide())
      store.dispatch(notificationAdd(alert))
      store.dispatch(toasterAdd(alert))
    }

    return Promise.reject(error)
  }
)

// Default headers will be included on all requests
const getHeaders = () => {
  return {
    "X-Auth-Type-Browser-UI": "1",
  }
}

// Converts object to JSON string
const updateFilter = (qp) => {
  const keys = ["filter", "sortBy"]
  keys.forEach((k) => {
    if (qp[k]) {
      // qp[k] = u.toString(qp[k])
      qp[k] = JSON.stringify(qp[k])
    }
  })
  return qp
}

const newRequest = (method, url, queryParams = {}, data, headers = {}) => {
  // Do not include token in header, now it is managed my cookie
  // grab current state
  // const auth = store.getState().entities.auth
  // if (auth.authenticated) {
  //   const token = "Bearer " + (auth.authenticated ? auth.token : "")
  //   headers = { ...headers, Authorization: token }
  // }

  return myAxios.request({
    method: method,
    url: getFinalUrl(url),
    data: data,
    headers: { ...getHeaders(), ...headers },
    params: updateFilter(queryParams),
  })
}

const getFinalUrl = (url) => {
  if (url.startsWith(SECURE_SHARE_PREFIX)) {
    return url
  }
  return "/api" + url
}

export const api = {
  status: {
    get: () => newRequest(HTTP_VERBS.GET, "/status", {}, {}),
  },
  version: {
    get: () => newRequest(HTTP_VERBS.GET, "/version", {}, {}),
  },
  auth: {
    login: (data) => newRequest(HTTP_VERBS.POST, "/user/login", {}, data),
    profile: () => newRequest(HTTP_VERBS.GET, "/user/profile", {}, {}),
    updateProfile: (data) => newRequest(HTTP_VERBS.POST, "/user/profile", {}, data),
    logout: () => newRequest(HTTP_VERBS.GET, "/mc_auth/sign_out", {}, {}),
  },
  gateway: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/gateway", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/gateway/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/gateway", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/gateway/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/gateway/disable", {}, data),
    reload: (data) => newRequest(HTTP_VERBS.POST, "/gateway/reload", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/gateway", {}, data),
    getSleepingQueue: (gwID, nodeID) =>
      newRequest(HTTP_VERBS.GET, "/gateway-sleeping-queue", { gatewayId: gwID, nodeId: nodeID }, {}),
    clearSleepingQueue: (gwID, nodeID) =>
      newRequest(HTTP_VERBS.GET, "/gateway-sleeping-queue/clear", { gatewayId: gwID, nodeId: nodeID }, {}),
  },
  node: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/node", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/node/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/node", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/node", {}, data),
  },
  source: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/source", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/source/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/source", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/source", {}, data),
  },
  field: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/field", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/field/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/field", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/field", {}, data),
  },
  virtualDevice: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/virtualdevice", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/virtualdevice/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/virtualdevice", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/virtualdevice", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/virtualdevice/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/virtualdevice/disable", {}, data),
  },
  virtualAssistant: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/virtualassistant", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/virtualassistant/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/virtualassistant", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/virtualassistant", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/virtualassistant/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/virtualassistant/disable", {}, data),
  },
  metric: {
    fetch: (queries) => newRequest(HTTP_VERBS.POST, "/metric", {}, queries),
  },
  action: {
    send: (queries) => newRequest(HTTP_VERBS.GET, "/action", queries, {}),
    post: (actions = []) => newRequest(HTTP_VERBS.POST, "/action", {}, actions),
    nodeAction: (queries) => newRequest(HTTP_VERBS.GET, "/action/node", queries, {}),
    gatewayAction: (queries) => newRequest(HTTP_VERBS.GET, "/action/gateway", queries, {}),
  },
  dashboard: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/dashboard", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/dashboard/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/dashboard", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/dashboard", {}, data),
  },
  forwardPayload: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/forwardpayload", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/forwardpayload/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/forwardpayload", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/forwardpayload", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/forwardpayload/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/forwardpayload/disable", {}, data),
  },
  firmware: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/firmware", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/firmware/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/firmware", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/firmware", {}, data),
    upload: (id, data, headers) => newRequest(HTTP_VERBS.POST, "/firmware/upload/" + id, {}, data, headers),
  },
  task: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/task", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/task/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/task", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/task", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/task/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/task/disable", {}, data),
  },
  handler: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/handler", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/handler/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/handler", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/handler/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/handler/disable", {}, data),
    reload: (data) => newRequest(HTTP_VERBS.POST, "/handler/reload", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/handler", {}, data),
  },
  schedule: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/schedule", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/schedule/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/schedule", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/schedule", {}, data),
    enable: (data) => newRequest(HTTP_VERBS.POST, "/schedule/enable", {}, data),
    disable: (data) => newRequest(HTTP_VERBS.POST, "/schedule/disable", {}, data),
  },
  dataRepository: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/datarepository", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/datarepository/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/datarepository", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/datarepository", {}, data),
  },
  settings: {
    updateSystem: (data) => newRequest(HTTP_VERBS.POST, "/settings", {}, data),
    getSystemSettings: () => newRequest(HTTP_VERBS.GET, "/settings/system", {}, {}),
    getJobs: () => newRequest(HTTP_VERBS.GET, "/settings/jobs", {}, {}),
    getBackupLocations: () => newRequest(HTTP_VERBS.GET, "/settings/backuplocations", {}, {}),
    resetJwtSecret: () => newRequest(HTTP_VERBS.GET, "/settings/system/jwtsecret/reset", {}, {}),
  },
  backup: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/backup", filter, {}),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/backup", {}, data),
    runOnDemandBackup: (data) => newRequest(HTTP_VERBS.POST, "/backup/run", {}, data),
    runRestore: (query) => newRequest(HTTP_VERBS.GET, "/restore/run", query, {}),
  },
  serviceAccount: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/serviceaccount", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/serviceaccount/" + id, {}, {}),
    create: (data) => newRequest(HTTP_VERBS.POST, "/serviceaccount/create", {}, data),
    update: (data) => newRequest(HTTP_VERBS.POST, "/serviceaccount/update", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/serviceaccount", {}, data),
  },
  user: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/user", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/user/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/user", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/user", {}, data),
  },
  policy: {
    list: (filter) => newRequest(HTTP_VERBS.GET, "/policy", filter, {}),
    get: (id) => newRequest(HTTP_VERBS.GET, "/policy/" + id, {}, {}),
    update: (data) => newRequest(HTTP_VERBS.POST, "/policy", {}, data),
    delete: (data) => newRequest(HTTP_VERBS.DELETE, "/policy", {}, data),
  },
  quickId: {
    getResources: (queries) => newRequest(HTTP_VERBS.GET, "/quickid", queries, {}),
  },
  secureShare: {
    get: (suffixUrl = "", queries = {}) => newRequest(HTTP_VERBS.GET, `/secure_share/${suffixUrl}`, queries),
  },
  websocket: {
    get: () => newRequest(HTTP_VERBS.GET, "/ws", {}, {}),
  },
}

// This key used to update system settings on a api
// api.settings.updateSystem == POST /settings
export const SystemApiKey = {
  SystemSettings: "system_settings",
  BackupLocations: "system_backup_locations",
}
