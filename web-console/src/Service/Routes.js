import { TachometerAltIcon } from "@patternfly/react-icons"
import React from "react"
import Dashboard from "../Pages/Dashboard/Dashboard"
import GatewayUpdatePage from "../Pages/Resources/Gateway/UpdatePage"
import GatewayDetailPage from "../Pages/Resources/Gateway/DetailsRootPage"
import GatewayListPage from "../Pages/Resources/Gateway/ListPage"
import NodeUpdatePage from "../Pages/Resources/Node/UpdatePage"
import NodeDetailPage from "../Pages/Resources/Node/DetailsRootPage"
import NodeListPage from "../Pages/Resources/Node/ListPage"
import SourceDetailPage from "../Pages/Resources/Source/DetailsRootPage"
import SourceListPage from "../Pages/Resources/Source/ListPage"
import SourceUpdatePage from "../Pages/Resources/Source/UpdatePage"
import FieldDetailPage from "../Pages/Resources/Field/DetailsRootPage"
import FieldListPage from "../Pages/Resources/Field/ListPage"
import FieldUpdatePage from "../Pages/Resources/Field/UpdatePage"
import VirtualDeviceDetailPage from "../Pages/Resources/VirtualDevice/DetailsRootPage"
import VirtualDeviceListPage from "../Pages/Resources/VirtualDevice/ListPage"
import VirtualDeviceUpdatePage from "../Pages/Resources/VirtualDevice/UpdatePage"
import VirtualAssistantDetailPage from "../Pages/Resources/VirtualAssistant/DetailsRootPage"
import VirtualAssistantListPage from "../Pages/Resources/VirtualAssistant/ListPage"
import VirtualAssistantUpdatePage from "../Pages/Resources/VirtualAssistant/UpdatePage"
import ProfilePage from "../Pages/Settings/Profile/DetailsPage"
import ForwardPayloadListPage from "../Pages/Operations/ForwardPayload/ListPage"
import ForwardPayloadUpdatePage from "../Pages/Operations/ForwardPayload/UpdatePage"
import ForwardPayloadDetailPage from "../Pages/Operations/ForwardPayload/DetailsRootPage"
import FirmwareListPage from "../Pages/Resources/Firmware/ListPage"
import FirmwareUpdatePage from "../Pages/Resources/Firmware/UpdatePage"
import FirmwareDetailPage from "../Pages/Resources/Firmware/DetailsRootPage"
import DataRepositoryListPage from "../Pages/Resources/DataRepository/ListPage"
import DataRepositoryUpdatePage from "../Pages/Resources/DataRepository/UpdatePage"
import DataRepositoryDetailPage from "../Pages/Resources/DataRepository/DetailsRootPage"
import TaskListPage from "../Pages/Operations/Task/ListPage"
import TaskUpdatePage from "../Pages/Operations/Task/UpdatePage"
import TaskDetailPage from "../Pages/Operations/Task/DetailsRootPage"
import HandlerListPage from "../Pages/Operations/Handler/ListPage"
import HandlerUpdatePage from "../Pages/Operations/Handler/UpdatePage"
import HandlerDetailPage from "../Pages/Operations/Handler/DetailsRootPage"
import ScheduleListPage from "../Pages/Operations/Schedule/ListPage"
import ScheduleUpdatePage from "../Pages/Operations/Schedule/UpdatePage"
import ScheduleDetailPage from "../Pages/Operations/Schedule/DetailsRootPage"
import SystemDetailsPage from "../Pages/Settings/System/DetailsRootPage"
import SystemBackupPage from "../Pages/Settings/Backup/ListPage"
import ServiceAccountDetailPage from "../Pages/Settings/ServiceAccount/DetailsRootPage"
import ServiceAccountListPage from "../Pages/Settings/ServiceAccount/ListPage"
import ServiceAccountUpdatePage from "../Pages/Settings/ServiceAccount/UpdatePage"
import UserDetailPage from "../Pages/Settings/User/DetailsRootPage"
import UserListPage from "../Pages/Settings/User/ListPage"
import UserUpdatePage from "../Pages/Settings/User/UpdatePage"
import PolicyDetailPage from "../Pages/Settings/Policy/DetailsRootPage"
import PolicyListPage from "../Pages/Settings/Policy/ListPage"
import PolicyUpdatePage from "../Pages/Settings/Policy/UpdatePage"

// import dummyPage from "../Pages/DummyPage/DummyPage"
import TopologyPage from "../Pages/Topology/Topology"

const routeMap = {
  home: "/",
  dashboard: "/dashboard",
  resources: {
    gateway: {
      list: "/resources/gateway",
      detail: "/resources/gateway/list/:id",
      update: "/resources/gateway/update/:id",
      add: "/resources/gateway/add",
    },
    node: {
      list: "/resources/node",
      detail: "/resources/node/list/:id",
      update: "/resources/node/update/:id",
      add: "/resources/node/add",
    },
    source: {
      list: "/resources/source",
      detail: "/resources/source/list/:id",
      update: "/resources/source/update/:id",
      add: "/resources/source/add",
    },
    field: {
      list: "/resources/field",
      detail: "/resources/field/list/:id",
      update: "/resources/field/update/:id",
      add: "/resources/field/add",
    },
    virtualDevice: {
      list: "/resources/virtualdevice",
      detail: "/resources/virtualdevice/list/:id",
      update: "/resources/virtualdevice/update/:id",
      add: "/resources/virtualdevice/add",
    },
    virtualAssistant: {
      list: "/resources/virtualassistant",
      detail: "/resources/virtualassistant/list/:id",
      update: "/resources/virtualassistant/update/:id",
      add: "/resources/virtualassistant/add",
    },
    firmware: {
      list: "/resources/firmware",
      detail: "/resources/firmware/list/:id",
      update: "/resources/firmware/update/:id",
      add: "/resources/firmware/add",
    },
    dataRepository: {
      list: "/resources/datarepository",
      detail: "/resources/datarepository/list/:id",
      update: "/resources/datarepository/update/:id",
      add: "/resources/datarepository/add",
    },
  },
  settings: {
    profile: "/settings/profile",
    system: "/settings/system",
    backup: "/settings/backup",
    serviceAccount: {
      list: "/settings/serviceaccount",
      detail: "/settings/serviceaccount/list/:id",
      update: "/settings/serviceaccount/update/:id",
      add: "/settings/serviceaccount/add",
    },
    user: {
      list: "/settings/user",
      detail: "/settings/user/list/:id",
      update: "/settings/user/update/:id",
      add: "/settings/user/add",
    },
    policy: {
      list: "/settings/policy",
      detail: "/settings/policy/list/:id",
      update: "/settings/policy/update/:id",
      add: "/settings/policy/add",
    },
  },
  operations: {
    forwardPayload: {
      list: "/operations/forwardpayload",
      detail: "/operations/forwardpayload/list/:id",
      update: "/operations/forwardpayload/update/:id",
      add: "/operations/forwardpayload/add",
    },
    task: {
      list: "/operations/task",
      detail: "/operations/task/list/:id",
      update: "/operations/task/update/:id",
      add: "/operations/task/add",
    },
    handler: {
      list: "/operations/handler",
      detail: "/operations/handler/list/:id",
      update: "/operations/handler/update/:id",
      add: "/operations/handler/add",
    },
    scheduler: {
      list: "/operations/schedule",
      detail: "/operations/schedule/list/:id",
      update: "/operations/schedule/update/:id",
      add: "/operations/schedule/add",
    },
  },
}

const routes = [
  {
    id: "dashboard",
    title: "dashboard",
    icon: <TachometerAltIcon />,
    children: [
      {
        id: "dashboardMain",
        title: "dashboard",
        to: "/dashboard/main",
        component: Dashboard,
        icon: <TachometerAltIcon />,
      },
      {
        id: "dashboardTopology",
        title: "Topology",
        to: "/dashboard/topology",
        component: TopologyPage,
      },
    ],
  },
  {
    id: "resources",
    title: "resources",
    children: [
      {
        id: "gateway",
        title: "gateways",
        to: routeMap.resources.gateway.list,
        component: GatewayListPage,
      },
      {
        id: "node",
        title: "nodes",
        to: routeMap.resources.node.list,
        component: NodeListPage,
      },
      {
        id: "source",
        title: "sources",
        to: routeMap.resources.source.list,
        component: SourceListPage,
      },
      {
        id: "field",
        title: "fields",
        to: routeMap.resources.field.list,
        component: FieldListPage,
      },
      {
        id: "separator_1",
        isSeparator: true,
      },
      {
        id: "firmware",
        title: "firmwares",
        to: routeMap.resources.firmware.list,
        component: FirmwareListPage,
      },
      {
        id: "dataRepository",
        title: "data_repository",
        to: routeMap.resources.dataRepository.list,
        component: DataRepositoryListPage,
      },
      {
        id: "separator_2",
        isSeparator: true,
      },
      {
        id: "virtual_device",
        title: "virtual_devices",
        to: routeMap.resources.virtualDevice.list,
        component: VirtualDeviceListPage,
        isExperimental: true,
      },
      {
        id: "virtual_assistant",
        title: "virtual_assistants",
        to: routeMap.resources.virtualAssistant.list,
        component: VirtualAssistantListPage,
        isExperimental: true,
      },
    ],
  },
  {
    id: "operations",
    title: "operations",
    children: [
      {
        id: "task",
        title: "tasks",
        to: routeMap.operations.task.list,
        component: TaskListPage,
      },
      {
        id: "scheduler",
        title: "schedules",
        to: routeMap.operations.scheduler.list,
        component: ScheduleListPage,
      },
      {
        id: "handler",
        title: "handlers",
        to: routeMap.operations.handler.list,
        component: HandlerListPage,
      },
      {
        id: "forwardpayload",
        title: "forward_payload",
        to: routeMap.operations.forwardPayload.list,
        component: ForwardPayloadListPage,
      },
    ],
  },
  {
    id: "settings",
    title: "settings",
    children: [
      {
        id: "profile",
        title: "profile",
        to: routeMap.settings.profile,
        component: ProfilePage,
      },
      {
        id: "system",
        title: "system",
        to: routeMap.settings.system,
        component: SystemDetailsPage,
      },
      {
        id: "backup-restore",
        title: "backup_and_restore",
        to: routeMap.settings.backup,
        component: SystemBackupPage,
      },
      {
        id: "separator_settings_1",
        isSeparator: true,
      },
      {
        id: "policies",
        title: "policies",
        to: routeMap.settings.policy.list,
        component: PolicyListPage,
      },
      {
        id: "users",
        title: "users",
        to: routeMap.settings.user.list,
        component: UserListPage,
      },
      {
        id: "service-account",
        title: "service_accounts",
        to: routeMap.settings.serviceAccount.list,
        component: ServiceAccountListPage,
      },
    ],
  },
]

const hiddenRoutes = [
  {
    to: routeMap.home,
    component: Dashboard,
  },
  {
    to: routeMap.resources.gateway.detail,
    component: GatewayDetailPage,
  },
  {
    to: routeMap.resources.gateway.add,
    component: GatewayUpdatePage,
  },
  {
    to: routeMap.resources.node.detail,
    component: NodeDetailPage,
  },
  {
    to: routeMap.resources.node.add,
    component: NodeUpdatePage,
  },
  {
    to: routeMap.resources.source.detail,
    component: SourceDetailPage,
  },
  {
    to: routeMap.resources.source.add,
    component: SourceUpdatePage,
  },
  {
    to: routeMap.resources.field.detail,
    component: FieldDetailPage,
  },
  {
    to: routeMap.resources.field.add,
    component: FieldUpdatePage,
  },
  {
    to: routeMap.operations.forwardPayload.detail,
    component: ForwardPayloadDetailPage,
  },
  {
    to: routeMap.operations.forwardPayload.add,
    component: ForwardPayloadUpdatePage,
  },
  {
    to: routeMap.resources.firmware.detail,
    component: FirmwareDetailPage,
  },
  {
    to: routeMap.resources.firmware.add,
    component: FirmwareUpdatePage,
  },
  {
    to: routeMap.resources.dataRepository.detail,
    component: DataRepositoryDetailPage,
  },
  {
    to: routeMap.resources.dataRepository.add,
    component: DataRepositoryUpdatePage,
  },
  {
    to: routeMap.resources.virtualDevice.detail,
    component: VirtualDeviceDetailPage,
  },
  {
    to: routeMap.resources.virtualDevice.add,
    component: VirtualDeviceUpdatePage,
  },
  {
    to: routeMap.resources.virtualAssistant.detail,
    component: VirtualAssistantDetailPage,
  },
  {
    to: routeMap.resources.virtualAssistant.add,
    component: VirtualAssistantUpdatePage,
  },
  {
    to: routeMap.operations.task.detail,
    component: TaskDetailPage,
  },
  {
    to: routeMap.operations.task.add,
    component: TaskUpdatePage,
  },
  {
    to: routeMap.operations.handler.detail,
    component: HandlerDetailPage,
  },
  {
    to: routeMap.operations.handler.add,
    component: HandlerUpdatePage,
  },
  {
    to: routeMap.operations.scheduler.detail,
    component: ScheduleDetailPage,
  },
  {
    to: routeMap.operations.scheduler.add,
    component: ScheduleUpdatePage,
  },
  {
    to: routeMap.settings.serviceAccount.detail,
    component: ServiceAccountDetailPage,
  },
  {
    to: routeMap.settings.serviceAccount.add,
    component: ServiceAccountUpdatePage,
  },
  {
    to: routeMap.settings.user.detail,
    component: UserDetailPage,
  },
  {
    to: routeMap.settings.user.add,
    component: UserUpdatePage,
  },
  {
    to: routeMap.settings.policy.detail,
    component: PolicyDetailPage,
  },
  {
    to: routeMap.settings.policy.add,
    component: PolicyUpdatePage,
  },
]

const redirect = (history, to = routeMap.home, urlParams = {}) => {
  if (history === null) {
    return
  }
  //console.log(history, to, urlParams)
  //const to = t(routeMap, name).safeString
  if (to) {
    let finalPath = to
    Object.keys(urlParams).forEach((key) => {
      finalPath = finalPath.replace(":" + key, urlParams[key])
    })
    history.push(finalPath)
  }
}

export { routes, hiddenRoutes, routeMap, redirect }
