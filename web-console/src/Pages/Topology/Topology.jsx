import "./Topology.scss"

import {
  Checkbox,
  EmptyState,
  EmptyStateBody,
  EmptyStateIcon,
  EmptyStateVariant,
  Spinner,
  Title,
  Tooltip,
  Toolbar,
  ToolbarContent,
  ToolbarGroup,
  ToolbarItem,
} from "@patternfly/react-core"
import {
  DagreLayout,
  DefaultEdge,
  DefaultGroup,
  DefaultNode,
  EdgeStyle,
  ForceLayout,
  GraphComponent,
  ModelKind,
  NodeShape,
  NodeStatus,
  Point,
  SELECTION_EVENT,
  TopologyControlBar,
  TopologyView,
  Visualization,
  VisualizationProvider,
  VisualizationSurface,
  action,
  createTopologyControlButtons,
  defaultControlButtonsOptions,
  withDragNode,
  withPanZoom,
  withSelection,
} from "@patternfly/react-topology"
import { RefreshButton, ResetButton } from "../../Components/Buttons/Buttons"
import { getFieldValue, getValue } from "../../Util/Util"

import { CircleNotchIcon as FieldIcon } from "@patternfly/react-icons"
import { PlugIcon as GatewayIcon } from "@patternfly/react-icons"
import { ContainerNodeIcon as NodeIcon } from "@patternfly/react-icons"
import PageContent from "../../Components/PageContent/PageContent"
import PageTitle from "../../Components/PageTitle/PageTitle"
import React from "react"
import Select from "../../Components/Form/Select"
import { MicrochipIcon as SourceIcon } from "@patternfly/react-icons"
import { TopologyIcon } from "@patternfly/react-icons"
import { getQuickId, ResourceType } from "../../Constants/ResourcePicker"
import { api } from "../../Service/Api"
import { redirect as r, routeMap as rMap } from "../../Service/Routes"
import { connect, useSelector } from "react-redux"
import { loadData, unloadData } from "../../store/entities/websocket"
import { useHistory } from "react-router-dom"
import { withTranslation } from "react-i18next"
import ItemDetails, { TOPOLOGY_WS_KEY } from "./ItemDetails"

const ALL_GATEWAYS = "__all__"
const ALL_NODES = "__all_nodes__"
const LAYOUT_FORCE = "Force"
const LAYOUT_DAGRE = "Dagre"
const PREFS_KEY = "mc.topology.prefs"
const WS_KEY = TOPOLOGY_WS_KEY
const CLICK_MOVE_PX = 5

const DEFAULT_PREFS = {
  selectedGatewayId: ALL_GATEWAYS,
  selectedNodeKey: ALL_NODES,
  showSources: false,
  showFields: false,
  includeChildNodes: false,
  selectedLayout: LAYOUT_FORCE,
}

const loadPrefs = () => {
  try {
    const raw = window.localStorage.getItem(PREFS_KEY)
    if (!raw) {
      return { ...DEFAULT_PREFS }
    }
    const parsed = JSON.parse(raw)
    const selectedLayout = parsed.selectedLayout === LAYOUT_DAGRE ? LAYOUT_DAGRE : LAYOUT_FORCE
    const selectedGatewayId = parsed.selectedGatewayId || ALL_GATEWAYS
    const allGateways = selectedGatewayId === ALL_GATEWAYS
    return {
      ...DEFAULT_PREFS,
      ...parsed,
      selectedLayout,
      selectedGatewayId,
      selectedNodeKey: allGateways ? ALL_NODES : parsed.selectedNodeKey || ALL_NODES,
      showSources: allGateways ? false : !!parsed.showSources,
      showFields: allGateways ? false : !!parsed.showFields,
      includeChildNodes:
        allGateways || !parsed.selectedNodeKey || parsed.selectedNodeKey === ALL_NODES
          ? false
          : !!parsed.includeChildNodes,
    }
  } catch (_err) {
    return { ...DEFAULT_PREFS }
  }
}
const MIN_READABLE_SCALE = 0.85
const LAYOUT_PADDING = 48
const NODE_DIAMETER = 64
const SOURCE_DIAMETER = 48
const FIELD_DIAMETER = 44

const applyReadableView = (graph) => {
  const nodes = graph.getNodes()
  if (!nodes.length) {
    return
  }
  let rect
  nodes.forEach((node) => {
    const bounds = node.getBounds()
    rect = rect ? rect.union(bounds) : bounds.clone()
  })
  if (!rect || rect.width === 0 || rect.height === 0) {
    return
  }
  const { width: viewWidth, height: viewHeight } = graph.getDimensions()
  if (viewWidth < 1 || viewHeight < 1) {
    return
  }
  const fitScale = Math.min(
    (viewWidth - LAYOUT_PADDING) / rect.width,
    (viewHeight - LAYOUT_PADDING) / rect.height
  )
  if (fitScale >= MIN_READABLE_SCALE) {
    graph.fit(LAYOUT_PADDING)
    return
  }
  graph.setScale(MIN_READABLE_SCALE)
  graph.setPosition(
    new Point(LAYOUT_PADDING / 2 - rect.x * MIN_READABLE_SCALE, LAYOUT_PADDING / 2 - rect.y * MIN_READABLE_SCALE)
  )
}

const baselineLayoutFactory = (type, graph) => {
  if (type === LAYOUT_DAGRE) {
    return new DagreLayout(graph, {
      layoutOnDrag: false,
      rankdir: "TB",
      ranker: "tight-tree",
      nodesep: 48,
      ranksep: 72,
      edgesep: 16,
      marginx: 24,
      marginy: 24,
    })
  }
  return new ForceLayout(graph, {
    nodeDistance: 50,
    linkDistance: 80,
    collideDistance: 16,
  })
}

const resourceData = (kind, id, extra = {}) => ({
  resourceKind: kind,
  resourceId: id,
  ...extra,
})

const secondaryName = (id, name) => {
  if (name === undefined || name === null) {
    return undefined
  }
  const text = String(name).trim()
  if (text === "" || text === String(id)) {
    return undefined
  }
  return text
}

const truncateLabel = (text, max) => {
  const value = String(text || "")
  if (!max || value.length <= max) {
    return value
  }
  const keep = Math.max(1, max - 1)
  const front = Math.ceil(keep / 2)
  const back = keep - front
  return `${value.slice(0, front)}…${value.slice(value.length - back)}`
}

const formatFieldValue = (field) => {
  const raw = getValue(field, "current.value", "")
  const value = getFieldValue(raw)
  const unit = field.unit && field.unit !== "none" ? ` ${field.unit}` : ""
  if (value === "") {
    return ""
  }
  return `${value}${unit}`
}

const openResourceDetails = (history, data) => {
  if (!history || !data || !data.resourceKind || !data.resourceId) {
    return
  }
  switch (data.resourceKind) {
    case "gateway":
      r(history, rMap.resources.gateway.detail, { id: data.resourceId })
      return
    case "node":
      r(history, rMap.resources.node.detail, { id: data.resourceId })
      return
    case "source":
      r(history, rMap.resources.source.detail, { id: data.resourceId })
      return
    case "field":
      r(history, rMap.resources.field.detail, { id: data.resourceId })
      return
    default:
      return
  }
}

const nodeIcon = (type) => {
  if (type === "gateway") {
    return GatewayIcon
  }
  if (type === "source") {
    return SourceIcon
  }
  if (type === "field") {
    return FieldIcon
  }
  return NodeIcon
}

const LABEL_PAD_X = 8
const LABEL_PAD_Y = 3
const LABEL_LINE_GAP = 0

const TopologyNodeCaption = ({ x, y, primary, value, secondary, maxLen }) => {
  const primaryRef = React.useRef(null)
  const secondaryRef = React.useRef(null)
  const idMax = value ? Math.max(8, maxLen - 8) : maxLen
  const primaryText = truncateLabel(primary, idMax)
  const valueText = value ? truncateLabel(value, 14) : ""
  const secondaryText = secondary ? truncateLabel(secondary, maxLen) : ""
  const [box, setBox] = React.useState({ width: 0, height: 18, primaryY: 9, secondaryY: 0 })

  React.useLayoutEffect(() => {
    const primaryEl = primaryRef.current
    if (!primaryEl) {
      return
    }
    const primaryBox = primaryEl.getBBox()
    const secondaryEl = secondaryRef.current
    const secondaryBox = secondaryEl ? secondaryEl.getBBox() : { width: 0, height: 0 }
    const width = Math.ceil(Math.max(primaryBox.width, secondaryBox.width) + LABEL_PAD_X * 2)
    const primaryY = LABEL_PAD_Y + primaryBox.height / 2
    const secondaryY = secondaryEl
      ? LABEL_PAD_Y + primaryBox.height + LABEL_LINE_GAP + secondaryBox.height / 2
      : 0
    const height = Math.ceil(
      LABEL_PAD_Y +
        primaryBox.height +
        (secondaryEl ? LABEL_LINE_GAP + secondaryBox.height : 0) +
        LABEL_PAD_Y
    )
    setBox((prev) => {
      if (
        prev.width === width &&
        prev.height === height &&
        prev.primaryY === primaryY &&
        prev.secondaryY === secondaryY
      ) {
        return prev
      }
      return { width, height, primaryY, secondaryY }
    })
  }, [primaryText, valueText, secondaryText])

  return (
    <g
      className="pf-topology__node__label topology-node-caption"
      transform={`translate(${x - box.width / 2}, ${y})`}
    >
      {box.width > 0 ? (
        <rect
          className="pf-topology__node__label__background"
          x={0}
          y={0}
          width={box.width}
          height={box.height}
          rx={4}
          ry={4}
        />
      ) : null}
      <text ref={primaryRef} textAnchor="middle" x={box.width / 2} y={box.primaryY} dy="0.35em">
        <tspan className="topology-label-id">{primaryText}</tspan>
        {valueText ? <tspan className="topology-label-value">{`  ${valueText}`}</tspan> : null}
      </text>
      {secondaryText ? (
        <text
          ref={secondaryRef}
          className="pf-m-secondary"
          textAnchor="middle"
          x={box.width / 2}
          y={box.secondaryY}
          dy="0.35em"
        >
          {secondaryText}
        </text>
      ) : null}
    </g>
  )
}

const graphIdOf = (kind, resource) => {
  if (!resource) {
    return null
  }
  if (kind === "gateway") {
    return `gw-${String(resource.id)}`
  }
  if (kind === "node") {
    return `node-${String(resource.gatewayId)}-${String(resource.nodeId)}`
  }
  if (kind === "source") {
    return `src-${String(resource.gatewayId)}-${String(resource.nodeId)}-${String(resource.sourceId)}`
  }
  if (kind === "field") {
    return `fld-${String(resource.gatewayId)}-${String(resource.nodeId)}-${String(resource.sourceId)}-${String(
      resource.fieldId
    )}`
  }
  return null
}

const CustomNode = ({ element, dragNodeRef, onSelectResource, selected }) => {
  const history = useHistory()
  const data = (element.getData && element.getData()) || {}
  const type = element.getType ? element.getType() : element.type
  const quickId = data.quickId
  const movedRef = React.useRef(false)
  const startRef = React.useRef(null)
  const liveField = useSelector((state) => {
    if (!quickId) {
      return undefined
    }
    const bucket = state.entities.websocket.data[WS_KEY]
    return bucket ? bucket[quickId] : undefined
  })
  const Icon = nodeIcon(type)
  const bounds = element.getBounds ? element.getBounds() : { width: NODE_DIAMETER, height: NODE_DIAMETER }
  const size = Math.min(bounds.width, bounds.height)
  const iconSize = Math.round(size * 0.5)
  const offset = (size - iconSize) / 2
  const onPointerDown = (event) => {
    movedRef.current = false
    startRef.current = { x: event.clientX, y: event.clientY }
  }
  const onPointerMove = (event) => {
    if (!startRef.current || movedRef.current) {
      return
    }
    const dx = event.clientX - startRef.current.x
    const dy = event.clientY - startRef.current.y
    if (dx * dx + dy * dy > CLICK_MOVE_PX * CLICK_MOVE_PX) {
      movedRef.current = true
    }
  }
  const onClick = (event) => {
    if (event.detail !== 1 || movedRef.current || !data.resourceKind || !data.resourceId) {
      return
    }
    event.stopPropagation()
    if (onSelectResource) {
      onSelectResource(data.resourceKind, data.resourceId)
    }
  }
  const onDoubleClick = (event) => {
    event.stopPropagation()
    event.preventDefault()
    openResourceDetails(history, data)
  }
  const isField = type === "field"
  const primaryLabel = isField
    ? (liveField && liveField.fieldId) || data.fieldId || (element.getLabel && element.getLabel()) || ""
    : (element.getLabel && element.getLabel()) || ""
  const valueLabel = isField ? (liveField ? formatFieldValue(liveField) : data.fieldValue || "") : ""
  return (
    <g
      className={data.resourceId ? "topology-resource-node" : undefined}
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onClick={onClick}
      onDoubleClick={onDoubleClick}
    >
      <DefaultNode
        element={element}
        dragNodeRef={dragNodeRef}
        selected={!!selected}
        showStatusDecorator={false}
        showLabel={false}
      >
        <g transform={`translate(${offset},${offset})`}>
          <Icon style={{ color: "#393F44" }} width={iconSize} height={iconSize} />
        </g>
        <TopologyNodeCaption
          x={bounds.width / 2}
          y={bounds.height + 6}
          primary={primaryLabel}
          value={valueLabel}
          secondary={data.secondaryLabel}
          maxLen={data.truncateLength || 18}
        />
      </DefaultNode>
    </g>
  )
}

const DraggableNode = withSelection()(withDragNode()(CustomNode))
const PannableGraph = withPanZoom()(GraphComponent)

const graphSignature = (model) => {
  const nodeIds = (model.nodes || []).map((n) => n.id).sort().join("\n")
  const edgeIds = (model.edges || []).map((e) => e.id).sort().join("\n")
  return `${(model.graph && model.graph.layout) || ""}\n${nodeIds}\n${edgeIds}`
}

class SurfaceReady extends React.Component {
  componentDidMount() {
    this.props.onReady()
  }

  render() {
    return this.props.children
  }
}

const nodeStatus = (item) => {
  const itemStatus = getValue(item, "state.status", "unknown")
  if (itemStatus == "up") {
    return NodeStatus.success
  }
  if (itemStatus == "down") {
    return NodeStatus.danger
  }
  return NodeStatus.default
}

const nodeKey = (item) => `${String(item.gatewayId)}::${String(item.nodeId)}`
const sourceKey = (item) => `${String(item.gatewayId)}::${String(item.nodeId)}::${String(item.sourceId)}`

const parentIdOf = (item) => {
  let raw = getValue(item, "labels.parent_id", undefined)
  if (raw === undefined || raw === null || raw === "") {
    raw = getValue(item, "others.parent_id", undefined)
  }
  if (raw === undefined || raw === null || raw === "") {
    return null
  }
  return String(raw)
}

const collectChildNodes = (nodes, rootKey) => {
  const root = nodes.find((n) => nodeKey(n) === rootKey)
  if (!root) {
    return []
  }
  const children = []
  const seen = new Set()
  const walk = (parentId) => {
    nodes.forEach((n) => {
      if (String(n.gatewayId) !== String(root.gatewayId)) {
        return
      }
      if (parentIdOf(n) !== String(parentId) || String(n.nodeId) === String(parentId)) {
        return
      }
      const key = nodeKey(n)
      if (seen.has(key)) {
        return
      }
      seen.add(key)
      children.push(n)
      walk(n.nodeId)
    })
  }
  walk(root.nodeId)
  return children
}

const listPayload = (response) => {
  const data = response && response.data ? response.data.data : null
  return Array.isArray(data) ? data : []
}

const eqFilter = (key, value) => ({ k: key, o: "eq", v: String(value) })

const scopedQuery = (filters) => ({
  limit: -1,
  offset: 0,
  filter: filters,
})

const buildTopologyModel = ({
  gateways,
  nodes,
  sources,
  fields,
  selectedGatewayId,
  selectedNodeKey,
  showSources,
  showFields,
  includeChildNodes,
  selectedLayout,
}) => {
  const visibleGateways =
    selectedGatewayId === ALL_GATEWAYS
      ? gateways
      : gateways.filter((gw) => String(gw.id) === String(selectedGatewayId))
  let visibleNodes =
    selectedGatewayId === ALL_GATEWAYS
      ? nodes
      : nodes.filter((n) => String(n.gatewayId) === String(selectedGatewayId))
  if (selectedNodeKey !== ALL_NODES) {
    const selected = visibleNodes.filter((n) => nodeKey(n) === selectedNodeKey)
    visibleNodes = includeChildNodes
      ? selected.concat(collectChildNodes(visibleNodes, selectedNodeKey))
      : selected
  }
  const visibleNodeKeys = new Set(visibleNodes.map(nodeKey))

  const visibleSources = showSources
    ? sources.filter((s) => visibleNodeKeys.has(nodeKey(s)))
    : []
  const visibleSourceKeys = new Set(visibleSources.map(sourceKey))

  const visibleFields = showFields
    ? fields.filter((f) => {
        if (!visibleNodeKeys.has(nodeKey(f))) {
          return false
        }
        if (showSources) {
          return visibleSourceKeys.has(sourceKey(f))
        }
        return true
      })
    : []

  const items = visibleGateways.map((gw) => ({
    id: `gw-${String(gw.id)}`,
    type: "gateway",
    label: gw.id,
    width: NODE_DIAMETER,
    height: NODE_DIAMETER,
    shape: NodeShape.hexagon,
    status: nodeStatus(gw),
    data: resourceData("gateway", gw.id, { secondaryLabel: secondaryName(gw.id, gw.name) }),
  }))

  const itemIds = new Set(items.map((item) => item.id))
  const edges = []

  const addItem = (item) => {
    items.push(item)
    itemIds.add(item.id)
  }

  const addEdge = (id, source, target, edgeStyle) => {
    if (!itemIds.has(source) || !itemIds.has(target)) {
      return
    }
    edges.push({ id, type: "edge", source, target, edgeStyle })
  }

  visibleNodes.forEach((n) => {
    addItem({
      id: `node-${String(n.gatewayId)}-${String(n.nodeId)}`,
      type: "node",
      label: String(n.nodeId),
      width: NODE_DIAMETER,
      height: NODE_DIAMETER,
      shape: NodeShape.circle,
      status: nodeStatus(n),
      data: resourceData("node", n.id, { secondaryLabel: secondaryName(n.nodeId, n.name) }),
    })
  })

  // Parent may appear after the child in API order; edges need both ends present.
  visibleNodes.forEach((n, index) => {
    const id = `node-${String(n.gatewayId)}-${String(n.nodeId)}`
    const parentId = parentIdOf(n)
    const parentNodeId = parentId !== null ? `node-${String(n.gatewayId)}-${parentId}` : ""
    if (parentId !== null && parentNodeId !== id && itemIds.has(parentNodeId)) {
      addEdge(`edge-node-${index}`, parentNodeId, id, EdgeStyle.default)
    } else {
      addEdge(`edge-node-${index}`, `gw-${String(n.gatewayId)}`, id, EdgeStyle.default)
    }
  })

  visibleSources.forEach((s, index) => {
    const id = `src-${String(s.gatewayId)}-${String(s.nodeId)}-${String(s.sourceId)}`
    addItem({
      id,
      type: "source",
      label: String(s.sourceId),
      width: SOURCE_DIAMETER,
      height: SOURCE_DIAMETER,
      shape: NodeShape.rect,
      status: nodeStatus(s),
      data: resourceData("source", s.id, { secondaryLabel: secondaryName(s.sourceId, s.name) }),
    })
    addEdge(`edge-src-${index}`, `node-${String(s.gatewayId)}-${String(s.nodeId)}`, id, EdgeStyle.default)
  })

  visibleFields.forEach((f, index) => {
    const id = `fld-${String(f.gatewayId)}-${String(f.nodeId)}-${String(f.sourceId)}-${String(f.fieldId)}`
    addItem({
      id,
      type: "field",
      label: String(f.fieldId || ""),
      width: FIELD_DIAMETER,
      height: FIELD_DIAMETER,
      shape: NodeShape.rect,
      status: NodeStatus.default,
      data: resourceData("field", f.id, {
        truncateLength: 28,
        secondaryLabel: secondaryName(f.fieldId, f.name),
        quickId: getQuickId(ResourceType.Field, f),
        fieldId: f.fieldId,
        fieldValue: formatFieldValue(f),
      }),
    })
    const parent = showSources
      ? `src-${String(f.gatewayId)}-${String(f.nodeId)}-${String(f.sourceId)}`
      : `node-${String(f.gatewayId)}-${String(f.nodeId)}`
    addEdge(`edge-fld-${index}`, parent, id, EdgeStyle.dashed)
  })

  return {
    nodes: items,
    edges,
    graph: {
      id: "g1",
      type: "graph",
      layout: selectedLayout || LAYOUT_FORCE,
    },
  }
}

class TopologyPage extends React.Component {
  constructor(props) {
    super(props)
    const prefs = loadPrefs()
    this.state = {
      loading: true,
      selectedGatewayId: prefs.selectedGatewayId,
      selectedNodeKey: prefs.selectedNodeKey,
      selectedLayout: prefs.selectedLayout,
      showSources: prefs.showSources,
      showFields: prefs.showFields,
      includeChildNodes: prefs.includeChildNodes,
      gateways: [],
      nodes: [],
      sources: [],
      fields: [],
      sidebarKind: null,
      sidebarResource: null,
      sidebarExtraSources: [],
      selectedGraphId: null,
    }
    this.alive = false
    this.detailsRequestId = 0
    this.listRequestId = 0
    this.sidebarSourceRequestId = 0
    this.lastGraphSignature = ""
  }

  setSelectedIds = (_id) => {}

  onSelectResource = (kind, id) => {
    const lists = {
      gateway: this.state.gateways,
      node: this.state.nodes,
      source: this.state.sources,
      field: this.state.fields,
    }
    const resource = (lists[kind] || []).find((item) => String(item.id) === String(id))
    if (!resource) {
      return
    }
    this.setState(
      {
        sidebarKind: kind,
        sidebarResource: resource,
        sidebarExtraSources: [],
        selectedGraphId: graphIdOf(kind, resource),
      },
      () => {
        this.ensureSidebarSources(kind, resource)
      }
    )
  }

  ensureSidebarSources = (kind, resource) => {
    if (kind !== "field" || !resource || !resource.gatewayId || !resource.sourceId) {
      return
    }
    const known = this.state.sources.concat(this.state.sidebarExtraSources)
    const found = known.find(
      (s) =>
        String(s.gatewayId) === String(resource.gatewayId) &&
        String(s.nodeId) === String(resource.nodeId) &&
        String(s.sourceId) === String(resource.sourceId)
    )
    if (found) {
      return
    }
    const requestId = this.sidebarSourceRequestId + 1
    this.sidebarSourceRequestId = requestId
    const resourceId = resource.id
    api.source
      .list(
        scopedQuery([
          eqFilter("gatewayId", resource.gatewayId),
          eqFilter("nodeId", resource.nodeId),
          eqFilter("sourceId", resource.sourceId),
        ])
      )
      .then(listPayload)
      .then((sources) => {
        if (!this.alive || requestId !== this.sidebarSourceRequestId) {
          return
        }
        if (!this.state.sidebarResource || this.state.sidebarResource.id !== resourceId) {
          return
        }
        if (sources.length) {
          this.setState({ sidebarExtraSources: sources })
        }
      })
      .catch(() => {})
  }

  closeSidebar = () => {
    this.setState({
      sidebarKind: null,
      sidebarResource: null,
      sidebarExtraSources: [],
      selectedGraphId: null,
    })
  }

  closedSidebarState = () => ({
    sidebarKind: null,
    sidebarResource: null,
    sidebarExtraSources: [],
    selectedGraphId: null,
  })

  persistPrefs = () => {
    try {
      window.localStorage.setItem(
        PREFS_KEY,
        JSON.stringify({
          selectedGatewayId: this.state.selectedGatewayId,
          selectedNodeKey: this.state.selectedNodeKey,
          selectedLayout: this.state.selectedLayout,
          showSources: this.state.showSources,
          showFields: this.state.showFields,
          includeChildNodes: this.state.includeChildNodes,
        })
      )
    } catch (_err) {
      // ignore quota / private mode
    }
  }

  componentDidMount() {
    this.topologyController = new Visualization()
    this.topologyController.registerLayoutFactory(baselineLayoutFactory)
    this.topologyController.registerComponentFactory((_kind, type) => {
      if (type === "group") {
        return DefaultGroup
      }
      switch (_kind) {
        case ModelKind.graph:
          return PannableGraph
        case ModelKind.node:
          return (props) => <DraggableNode {...props} onSelectResource={this.onSelectResource} />
        case ModelKind.edge:
          return DefaultEdge
        default:
          return undefined
      }
    })
    this.topologyController.addEventListener(SELECTION_EVENT, this.setSelectedIds)
    this.alive = true
    const prefs = loadPrefs()
    Promise.all([
      api.gateway.list(scopedQuery([])).catch(() => ({ data: { data: [] } })),
      api.node.list(scopedQuery([])).catch(() => ({ data: { data: [] } })),
    ])
      .then((values) => {
        if (!this.alive) {
          return
        }
        const gateways = listPayload(values[0])
        const nodes = listPayload(values[1])
        const gatewayExists =
          prefs.selectedGatewayId === ALL_GATEWAYS ||
          gateways.some((gw) => String(gw.id) === String(prefs.selectedGatewayId))
        const selectedGatewayId = gatewayExists ? prefs.selectedGatewayId : ALL_GATEWAYS
        const nodeExists = nodes.some((n) => nodeKey(n) === prefs.selectedNodeKey)
        this.setState(
          {
            gateways,
            nodes,
            selectedGatewayId,
            selectedNodeKey:
              selectedGatewayId === ALL_GATEWAYS || !nodeExists ? ALL_NODES : prefs.selectedNodeKey,
            showSources: selectedGatewayId === ALL_GATEWAYS ? false : prefs.showSources,
            showFields: selectedGatewayId === ALL_GATEWAYS ? false : prefs.showFields,
            includeChildNodes:
              selectedGatewayId === ALL_GATEWAYS ||
              !prefs.selectedNodeKey ||
              prefs.selectedNodeKey === ALL_NODES
                ? false
                : prefs.includeChildNodes,
            selectedLayout: prefs.selectedLayout,
            loading: false,
          },
          () => {
            this.persistPrefs()
            this.refreshDetails(this.state)
          }
        )
      })
      .catch(() => {
        if (this.alive) {
          this.setState({ loading: false })
        }
      })
  }

  filtersForSelection = (gatewayId, nodeKey, includeChildNodes, nodes) => {
    const filters = []
    if (gatewayId && gatewayId !== ALL_GATEWAYS) {
      filters.push(eqFilter("gatewayId", gatewayId))
    }
    if (nodeKey && nodeKey !== ALL_NODES) {
      const nodeIds = [nodeKey.split("::")[1]]
      if (includeChildNodes) {
        collectChildNodes(nodes || [], nodeKey).forEach((n) => nodeIds.push(String(n.nodeId)))
      }
      if (nodeIds.length === 1) {
        filters.push(eqFilter("nodeId", nodeIds[0]))
      } else {
        filters.push({ k: "nodeId", o: "in", v: nodeIds })
      }
    }
    return filters
  }

  refreshDetails = (nextState) => {
    const requestId = this.detailsRequestId + 1
    this.detailsRequestId = requestId
    const gatewayId = nextState.selectedGatewayId
    const nodeKey = nextState.selectedNodeKey
    const showSources = nextState.showSources
    const showFields = nextState.showFields
    const includeChildNodes = nextState.includeChildNodes
    const apply = (patch) => {
      if (!this.alive || requestId !== this.detailsRequestId) {
        return
      }
      this.setState(patch)
    }
    if (gatewayId === ALL_GATEWAYS) {
      apply({
        sources: nextState.sources.length ? [] : nextState.sources,
        fields: nextState.fields.length ? [] : nextState.fields,
      })
      return
    }
    if (!showSources && nextState.sources.length) {
      apply({ sources: [] })
    }
    if (!showFields && nextState.fields.length) {
      apply({ fields: [] })
    }
    const filters = this.filtersForSelection(gatewayId, nodeKey, includeChildNodes, nextState.nodes)
    if (showSources) {
      api.source
        .list(scopedQuery(filters))
        .then(listPayload)
        .then((sources) => apply({ sources }))
        .catch(() => apply({ sources: [] }))
    }
    if (showFields) {
      api.field
        .list(scopedQuery(filters))
        .then(listPayload)
        .then((fields) => apply({ fields }))
        .catch(() => apply({ fields: [] }))
    }
  }

  syncFieldEvents = () => {
    const { showFields, fields } = this.state
    if (!showFields || !fields.length) {
      this.props.unloadData({ key: WS_KEY })
      return
    }
    const resources = {}
    fields.forEach((f) => {
      resources[getQuickId(ResourceType.Field, f)] = f
    })
    this.props.loadData({ key: WS_KEY, resources })
  }

  componentDidUpdate(_prevProps, prevState) {
    if (prevState.fields !== this.state.fields || prevState.showFields !== this.state.showFields) {
      this.syncFieldEvents()
    }
    if (
      prevState.loading !== this.state.loading ||
      prevState.gateways !== this.state.gateways ||
      prevState.nodes !== this.state.nodes ||
      prevState.sources !== this.state.sources ||
      prevState.fields !== this.state.fields ||
      prevState.selectedGatewayId !== this.state.selectedGatewayId ||
      prevState.selectedNodeKey !== this.state.selectedNodeKey ||
      prevState.showSources !== this.state.showSources ||
      prevState.showFields !== this.state.showFields ||
      prevState.includeChildNodes !== this.state.includeChildNodes ||
      prevState.selectedLayout !== this.state.selectedLayout
    ) {
      this.updateGraph()
    }
  }

  scheduleLayout = () => {
    if (this.layoutTimer) {
      window.clearTimeout(this.layoutTimer)
    }
    this.layoutTimer = window.setTimeout(() => {
      this.layoutTimer = null
      if (!this.topologyController) {
        return
      }
      action(() => {
        const graph = this.topologyController.getGraph()
        if (graph) {
          graph.setScaleExtent([0.35, 3])
          graph.layout()
          applyReadableView(graph)
        }
      })()
    }, 50)
  }

  onSurfaceReady = () => {
    this.surfaceReady = true
    this.updateGraph()
  }

  updateGraph = () => {
    if (!this.topologyController || this.state.loading || !this.surfaceReady) {
      return
    }
    const model = buildTopologyModel(this.state)
    const signature = graphSignature(model)
    const structureChanged = signature !== this.lastGraphSignature
    this.lastGraphSignature = signature
    this.topologyController.fromModel(model, true)
    if (structureChanged && model.nodes.length) {
      this.scheduleLayout()
    }
  }

  componentWillUnmount() {
    this.alive = false
    this.detailsRequestId += 1
    this.listRequestId += 1
    this.sidebarSourceRequestId += 1
    if (this.layoutTimer) {
      window.clearTimeout(this.layoutTimer)
    }
    this.props.unloadData({ key: WS_KEY })
  }

  visibleNodesForFilter = () => {
    const { nodes, selectedGatewayId } = this.state
    if (selectedGatewayId === ALL_GATEWAYS) {
      return nodes
    }
    return nodes.filter((n) => String(n.gatewayId) === String(selectedGatewayId))
  }

  onGatewaySelect = (selection) => {
    const selectedGatewayId = selection || ALL_GATEWAYS
    const goingToAll = selectedGatewayId === ALL_GATEWAYS
    this.setState(
      {
        selectedGatewayId,
        selectedNodeKey: ALL_NODES,
        includeChildNodes: false,
        showSources: goingToAll ? false : this.state.showSources,
        showFields: goingToAll ? false : this.state.showFields,
        sources: [],
        fields: [],
        ...this.closedSidebarState(),
      },
      () => {
        this.persistPrefs()
        this.refreshDetails(this.state)
      }
    )
  }

  onNodeSelect = (selection) => {
    const selectedNodeKey = selection || ALL_NODES
    this.setState(
      {
        selectedNodeKey,
        includeChildNodes: selectedNodeKey === ALL_NODES ? false : this.state.includeChildNodes,
        sources: [],
        fields: [],
        ...this.closedSidebarState(),
      },
      () => {
        this.persistPrefs()
        this.refreshDetails(this.state)
      }
    )
  }

  onToggleChildNodes = (checked, event) => {
    const includeChildNodes = typeof checked === "boolean" ? checked : event.currentTarget.checked
    this.setState({ includeChildNodes, sources: [], fields: [], ...this.closedSidebarState() }, () => {
      this.persistPrefs()
      this.refreshDetails(this.state)
    })
  }

  onToggleSources = (checked, event) => {
    const showSources = typeof checked === "boolean" ? checked : event.currentTarget.checked
    this.setState({ showSources, ...this.closedSidebarState() }, () => {
      this.persistPrefs()
      this.refreshDetails(this.state)
    })
  }

  onToggleFields = (checked, event) => {
    const showFields = typeof checked === "boolean" ? checked : event.currentTarget.checked
    this.setState({ showFields, ...this.closedSidebarState() }, () => {
      this.persistPrefs()
      this.refreshDetails(this.state)
    })
  }

  onLayoutSelect = (selection) => {
    this.setState({ selectedLayout: selection, ...this.closedSidebarState() }, this.persistPrefs)
  }

  onReset = () => {
    this.setState(
      {
        ...DEFAULT_PREFS,
        sources: [],
        fields: [],
        ...this.closedSidebarState(),
      },
      () => {
        this.persistPrefs()
        this.refreshDetails(this.state)
      }
    )
  }

  onRefresh = () => {
    this.setState({ ...this.closedSidebarState() })
    const requestId = this.listRequestId + 1
    this.listRequestId = requestId
    Promise.all([
      api.gateway.list(scopedQuery([])).catch(() => ({ data: { data: [] } })),
      api.node.list(scopedQuery([])).catch(() => ({ data: { data: [] } })),
    ])
      .then((values) => {
        if (!this.alive || requestId !== this.listRequestId) {
          return
        }
        this.setState(
          {
            gateways: listPayload(values[0]),
            nodes: listPayload(values[1]),
          },
          () => this.refreshDetails(this.state)
        )
      })
      .catch(() => {})
  }

  renderToolbar = () => {
    const { t } = this.props
    const {
      gateways,
      nodes,
      selectedGatewayId,
      selectedNodeKey,
      selectedLayout,
      includeChildNodes,
      showSources,
      showFields,
      sources,
      fields,
      loading,
    } = this.state
    const gatewayOptions = [{ value: ALL_GATEWAYS, label: t("all_gateways"), searchText: t("all_gateways") }]
    gateways.forEach((gw) => {
      const count = nodes.filter((n) => String(n.gatewayId) === String(gw.id)).length
      const name = gw.name || ""
      gatewayOptions.push({
        value: gw.id,
        label: `${gw.id} (${count})`,
        name,
        description: name && name !== gw.id ? name : gw.description || undefined,
        searchText: `${gw.id} ${name} ${gw.description || ""}`,
      })
    })

    const detailsDisabled = loading || selectedGatewayId === ALL_GATEWAYS
    const childNodesDisabled = detailsDisabled || selectedNodeKey === ALL_NODES
    const nodeOptions = [{ value: ALL_NODES, label: t("all_nodes"), searchText: t("all_nodes") }]
    this.visibleNodesForFilter().forEach((n) => {
      const name = n.name || ""
      const id = String(n.nodeId)
      const description = [name && name !== id ? name : "", selectedGatewayId === ALL_GATEWAYS ? n.gatewayId : ""]
        .filter(Boolean)
        .join(" · ")
      nodeOptions.push({
        value: nodeKey(n),
        label: id,
        name,
        description: description || undefined,
        searchText: `${id} ${name} ${n.gatewayId || ""}`,
      })
    })
    const layoutOptions = [
      { value: LAYOUT_FORCE, label: t("layout_force") },
      { value: LAYOUT_DAGRE, label: t("layout_tree") },
    ]

    return (
      <Toolbar className="topology-toolbar" id="topology-toolbar">
        <ToolbarContent>
          <ToolbarGroup variant="filter-group">
            <ToolbarItem>
              <div className="topology-filter-select">
                <Select
                  options={gatewayOptions}
                  selected={selectedGatewayId}
                  onChange={this.onGatewaySelect}
                  isDisabled={loading || gateways.length === 0}
                  isSearchable
                  maxHeight="300px"
                  label={t("select_gateway")}
                />
              </div>
            </ToolbarItem>
            <ToolbarItem>
              <div className="topology-filter-select">
                <Select
                  options={nodeOptions}
                  selected={selectedNodeKey}
                  onChange={this.onNodeSelect}
                  isDisabled={detailsDisabled || this.visibleNodesForFilter().length === 0}
                  isSearchable
                  maxHeight="300px"
                  label={t("node")}
                />
              </div>
            </ToolbarItem>
            <ToolbarItem>
              <Tooltip content={t("include_child_nodes")}>
                <Checkbox
                  id="topology-include-child-nodes"
                  label={t("child_nodes")}
                  isChecked={includeChildNodes}
                  isDisabled={childNodesDisabled}
                  onChange={this.onToggleChildNodes}
                />
              </Tooltip>
            </ToolbarItem>
            <ToolbarItem>
              <Tooltip content={t("show_sources")}>
                <Checkbox
                  id="topology-show-sources"
                  label={`${t("sources")} (${sources.length})`}
                  isChecked={showSources}
                  isDisabled={detailsDisabled}
                  onChange={this.onToggleSources}
                />
              </Tooltip>
            </ToolbarItem>
            <ToolbarItem>
              <Tooltip content={t("show_fields")}>
                <Checkbox
                  id="topology-show-fields"
                  label={`${t("fields")} (${fields.length})`}
                  isChecked={showFields}
                  isDisabled={detailsDisabled}
                  onChange={this.onToggleFields}
                />
              </Tooltip>
            </ToolbarItem>
            <ToolbarItem>
              <ResetButton onClick={this.onReset} isDisabled={loading} isSmall variant="secondary" />
            </ToolbarItem>
          </ToolbarGroup>
          <ToolbarGroup alignment={{ default: "alignRight" }}>
            <ToolbarItem>
              <RefreshButton onClick={this.onRefresh} />
            </ToolbarItem>
            <ToolbarItem>
              <Select
                options={layoutOptions}
                selected={selectedLayout}
                onChange={this.onLayoutSelect}
                isDisabled={loading}
                disableClear
                label={t("layout")}
              />
            </ToolbarItem>
          </ToolbarGroup>
        </ToolbarContent>
      </Toolbar>
    )
  }

  renderEmptyState = () => {
    const { t } = this.props
    const { selectedGatewayId } = this.state
    const titleKey =
      selectedGatewayId && selectedGatewayId !== ALL_GATEWAYS
        ? "no_topology_nodes_for_gateway"
        : "no_topology_nodes"
    return (
      <EmptyState variant={EmptyStateVariant.small} className="topology-empty-state">
        <EmptyStateIcon icon={TopologyIcon} />
        <Title headingLevel="h2" size="lg">
          {t(titleKey)}
        </Title>
        <EmptyStateBody>{t("no_data")}</EmptyStateBody>
      </EmptyState>
    )
  }

  render() {
    const {
      loading,
      gateways,
      nodes,
      sources,
      sidebarExtraSources,
      selectedGatewayId,
      selectedNodeKey,
      sidebarKind,
      sidebarResource,
      selectedGraphId,
    } = this.state
    let visibleNodes =
      selectedGatewayId === ALL_GATEWAYS
        ? nodes
        : nodes.filter((n) => String(n.gatewayId) === String(selectedGatewayId))
    if (selectedNodeKey !== ALL_NODES) {
      visibleNodes = visibleNodes.filter((n) => nodeKey(n) === selectedNodeKey)
    }
    const visibleGatewayCount =
      selectedGatewayId === ALL_GATEWAYS
        ? gateways.length
        : gateways.filter((gw) => String(gw.id) === String(selectedGatewayId)).length
    const hasGraph = !loading && (visibleGatewayCount > 0 || visibleNodes.length > 0)

    return (
      <React.Fragment>
        <PageTitle title="Topology" />
        <PageContent>
          {this.renderToolbar()}
          {loading ? (
            <Spinner key="loading" size="md" />
          ) : (
            <TopologyView
              key="tview"
              className="ws-react-c-topology"
              sideBarOpen={!!sidebarResource}
              sideBar={
                <ItemDetails
                  show={!!sidebarResource}
                  onClose={this.closeSidebar}
                  kind={sidebarKind}
                  resource={sidebarResource}
                  history={this.props.history}
                  nodes={nodes}
                  sources={sources.concat(sidebarExtraSources || [])}
                />
              }
              controlBar={
                <TopologyControlBar
                  controlButtons={createTopologyControlButtons({
                    ...defaultControlButtonsOptions,
                    zoomInCallback: action(() => {
                      this.topologyController.getGraph().scaleBy(4 / 3)
                    }),
                    zoomOutCallback: action(() => {
                      this.topologyController.getGraph().scaleBy(0.75)
                    }),
                    fitToScreenCallback: action(() => {
                      this.topologyController.getGraph().fit(80)
                    }),
                    resetViewCallback: action(() => {
                      const graph = this.topologyController.getGraph()
                      graph.layout()
                      applyReadableView(graph)
                    }),
                    legend: false,
                  })}
                />
              }
            >
              {hasGraph ? (
                <VisualizationProvider controller={this.topologyController}>
                  <SurfaceReady onReady={this.onSurfaceReady}>
                    <VisualizationSurface
                      state={{ selectedIds: selectedGraphId ? [selectedGraphId] : [] }}
                    />
                  </SurfaceReady>
                </VisualizationProvider>
              ) : (
                this.renderEmptyState()
              )}
            </TopologyView>
          )}
        </PageContent>
      </React.Fragment>
    )
  }
}

export default connect(null, { loadData, unloadData })(withTranslation()(TopologyPage))
