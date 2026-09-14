import { combineReducers } from "redux"
import aboutReducer from "./entities/about"
import authReducer from "./entities/auth"
import dashboardReducer from "./entities/dashboard"
import spinnerReducer from "./entities/globalSpinner"
import localeReducer from "./entities/locale"
import notificationReducer from "./entities/notification"
import operationForwardPayloadReducer from "./entities/operations/forwardPayload"
import operationHandlerReducer from "./entities/operations/handler"
import operationSchedulerReducer from "./entities/operations/scheduler"
import operationTaskReducer from "./entities/operations/task"
import resourceDataRepositoryReducer from "./entities/resources/dataRepository"
import resourceFieldReducer from "./entities/resources/field"
import resourceFirmwareReducer from "./entities/resources/firmware"
import resourceGatewayReducer from "./entities/resources/gateway"
import resourceNodeReducer from "./entities/resources/node"
import resourceSourceReducer from "./entities/resources/source"
import resourceVirtualAssistantReducer from "./entities/resources/virtualAssistant"
import resourceVirtualDeviceReducer from "./entities/resources/virtualDevice"
import systemBackupReducer from "./entities/system/backup"
import themeReducer from "./entities/theme"
import toasterReducer from "./entities/toaster"
import websocketReducer from "./entities/websocket"
import settingsServiceAccountReducer from "./entities/system/serviceAccount"
import settingsUserReducer from "./entities/system/user"
import settingsPolicyReducer from "./entities/system/policy"

export default combineReducers({
  notification: notificationReducer,
  toaster: toasterReducer,
  about: aboutReducer,
  auth: authReducer,
  dashboard: dashboardReducer,
  spinner: spinnerReducer,
  resourceGateway: resourceGatewayReducer,
  resourceNode: resourceNodeReducer,
  resourceSource: resourceSourceReducer,
  resourceField: resourceFieldReducer,
  resourceFirmware: resourceFirmwareReducer,
  resourceDataRepository: resourceDataRepositoryReducer,
  operationForwardPayload: operationForwardPayloadReducer,
  resourceVirtualAssistant: resourceVirtualAssistantReducer,
  resourceVirtualDevice: resourceVirtualDeviceReducer,
  operationTask: operationTaskReducer,
  operationHandler: operationHandlerReducer,
  operationScheduler: operationSchedulerReducer,
  systemBackup: systemBackupReducer,
  websocket: websocketReducer,
  locale: localeReducer,
  theme: themeReducer,
  settingsServiceAccount: settingsServiceAccountReducer,
  settingsUser: settingsUserReducer,
  settingsPolicy: settingsPolicyReducer,
})
