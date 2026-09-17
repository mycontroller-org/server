import "./Topology.scss"

import {
  Button,
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
import { DownloadIcon } from "@patternfly/react-icons"
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
  selectedGatewayIds: [],
  selectedNodeKeys: [],
  showSources: false,
  showFields: false,
  includeChildNodes: false,
  selectedLayout: LAYOUT_FORCE,
}

const normalizeNodeKeys = (value) => {
  if (Array.isArray(value)) {
    return value.filter((key) => key && key !== ALL_NODES)
  }
  if (value && value !== ALL_NODES) {
    return [value]
  }
  return []
}

const normalizeGatewayIds = (value) => {
  if (Array.isArray(value)) {
    return value.filter((id) => id && id !== ALL_GATEWAYS).map(String)
  }
  if (value && value !== ALL_GATEWAYS) {
    return [String(value)]
  }
  return []
}

const loadPrefs = () => {
  try {
    const raw = window.localStorage.getItem(PREFS_KEY)
    if (!raw) {
      return { ...DEFAULT_PREFS }
    }
    const parsed = JSON.parse(raw)
    const selectedLayout = parsed.selectedLayout === LAYOUT_DAGRE ? LAYOUT_DAGRE : LAYOUT_FORCE
    const selectedGatewayIds = normalizeGatewayIds(parsed.selectedGatewayIds || parsed.selectedGatewayId)
    const allGateways = selectedGatewayIds.length === 0
    return {
      ...DEFAULT_PREFS,
      ...parsed,
      selectedLayout,
      selectedGatewayIds,
      selectedNodeKeys: allGateways ? [] : normalizeNodeKeys(parsed.selectedNodeKeys || parsed.selectedNodeKey),
      showSources: allGateways ? false : !!parsed.showSources,
      showFields: allGateways ? false : !!parsed.showFields,
      includeChildNodes:
        allGateways || !normalizeNodeKeys(parsed.selectedNodeKeys || parsed.selectedNodeKey).length
          ? false
          : !!parsed.includeChildNodes,
    }
  } catch (_err) {
    return { ...DEFAULT_PREFS }
  }
}
const SVG_STYLE_PROPS = [
  "fill",
  "fill-opacity",
  "stroke",
  "stroke-width",
  "stroke-dasharray",
  "stroke-opacity",
  "opacity",
  "font-family",
  "font-size",
  "font-weight",
  "font-style",
  "text-anchor",
  "dominant-baseline",
  "letter-spacing",
  "color",
  "display",
  "visibility",
]

const copySvgComputedStyles = (source, target) => {
  if (!source || !target || source.nodeType !== 1) {
    return
  }
  const computed = window.getComputedStyle(source)
  const css = SVG_STYLE_PROPS.map((prop) => {
    const value = computed.getPropertyValue(prop)
    return value ? `${prop}:${value}` : ""
  })
    .filter(Boolean)
    .join(";")
  if (css) {
    target.setAttribute("style", css)
  }
  const sourceKids = source.children || []
  const targetKids = target.children || []
  for (let i = 0; i < sourceKids.length && i < targetKids.length; i++) {
    copySvgComputedStyles(sourceKids[i], targetKids[i])
  }
}

const downloadSvg = (svg, background, fileName) => {
  if (!svg) {
    return
  }
  const rect = svg.getBoundingClientRect()
  const width = Math.max(1, Math.round(rect.width))
  const height = Math.max(1, Math.round(rect.height))
  const clone = svg.cloneNode(true)
  clone.setAttribute("xmlns", "http://www.w3.org/2000/svg")
  clone.setAttribute("xmlns:xlink", "http://www.w3.org/1999/xlink")
  clone.setAttribute("width", String(width))
  clone.setAttribute("height", String(height))
  copySvgComputedStyles(svg, clone)
  const backdrop = document.createElementNS("http://www.w3.org/2000/svg", "rect")
  backdrop.setAttribute("x", "0")
  backdrop.setAttribute("y", "0")
  backdrop.setAttribute("width", "100%")
  backdrop.setAttribute("height", "100%")
  backdrop.setAttribute("fill", background || "#f0f0f0")
  clone.insertBefore(backdrop, clone.firstChild)
  const source = new XMLSerializer().serializeToString(clone)
  const blob = new Blob([source], { type: "image/svg+xml;charset=utf-8" })
  const link = document.createElement("a")
  link.href = URL.createObjectURL(blob)
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => URL.revokeObjectURL(link.href), 1000)
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

// ForceLayout hardcodes layoutOnDrag: true. Turn that off so dragging one
// node does not restart the simulation and bounce the rest.
class TopologyForceLayout extends ForceLayout {
  constructor(graph, options = {}) {
    super(graph, options)
    this.options.layoutOnDrag = false
  }

  // PatternFly calls getParent() on every link. During fromModel / halt,
  // a node can exist without a parent yet and the whole layout throws,
  // leaving every resource stacked. We do not use groups, so skip that.
  getLinkDistance = (link) => {
    const source = link && link.source
    const target = link && link.target
    const sourceRadius = source && source.radius ? source.radius : 0
    const targetRadius = target && target.radius ? target.radius : 0
    return this.options.linkDistance + sourceRadius + targetRadius
  }

  stopSimulation() {
    try {
      super.stopSimulation()
    } catch (_err) {
      // stale links after a model merge
    }
  }
}

const MIN_GRAPH_SIZE = 40

const graphHasSize = (graph) => {
  if (!graph || typeof graph.getDimensions !== "function") {
    return false
  }
  const { width, height } = graph.getDimensions()
  return width >= MIN_GRAPH_SIZE && height >= MIN_GRAPH_SIZE
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
  return new TopologyForceLayout(graph, {
    nodeDistance: 50,
    linkDistance: 80,
    collideDistance: 16,
    chargeStrength: -80,
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

// Live Alt state: left or right, press/release anytime during the drag.
let altHeld = false

const syncAltFromEvent = (event) => {
  if (!event) {
    return
  }
  if (typeof event.altKey === "boolean") {
    altHeld = event.altKey
    return
  }
  const key = event.key || event.code
  if (key === "Alt" || key === "AltLeft" || key === "AltRight") {
    altHeld = event.type !== "keyup"
  }
}

const onAltKeyChange = (event) => {
  syncAltFromEvent(event)
}

const onAltLost = () => {
  altHeld = false
}

const onAltVisibilityChange = () => {
  if (document.hidden) {
    altHeld = false
  }
}

const collectDescendantNodes = (node) => {
  if (!node || typeof node.getGraph !== "function") {
    return []
  }
  let graph
  try {
    graph = node.getGraph()
  } catch (_err) {
    return []
  }
  const descendants = []
  const seen = new Set([node.getId()])
  const walk = (id) => {
    const edges = (graph.getEdges && graph.getEdges()) || []
    edges.forEach((edge) => {
      const source = edge.getSource && edge.getSource()
      const target = edge.getTarget && edge.getTarget()
      if (!source || !target || source.getId() !== id || seen.has(target.getId())) {
        return
      }
      seen.add(target.getId())
      descendants.push(target)
      walk(target.getId())
    })
  }
  walk(node.getId())
  return descendants
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
    syncAltFromEvent(event)
  }
  const onPointerMove = (event) => {
    syncAltFromEvent(event)
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

const DraggableNode = withSelection()(
  withDragNode({
    drag: (event, monitor) => {
      const sourceEvent = event && (event.sourceEvent || event.event)
      syncAltFromEvent(sourceEvent)
      if (!altHeld) {
        return
      }
      const dx = event.dx
      const dy = event.dy
      if (!dx && !dy) {
        return
      }
      const dragged = monitor.getItem()
      collectDescendantNodes(dragged).forEach((child) => {
        const pos = child.getPosition && child.getPosition()
        if (pos && pos.clone) {
          child.setPosition(pos.clone().translate(dx, dy))
        }
      })
    },
  })(CustomNode)
)
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

const collectChildNodesForKeys = (nodes, rootKeys) => {
  const extra = []
  const seen = new Set(rootKeys)
  rootKeys.forEach((key) => {
    collectChildNodes(nodes, key).forEach((n) => {
      const keyId = nodeKey(n)
      if (!seen.has(keyId)) {
        seen.add(keyId)
        extra.push(n)
      }
    })
  })
  return extra
}

const listPayload = (response) => {
  const data = response && response.data ? response.data.data : null
  return Array.isArray(data) ? data : []
}

const compareAlphanumeric = (a, b) =>
  String(a ?? "").localeCompare(String(b ?? ""), undefined, { numeric: true, sensitivity: "base" })

const sortGateways = (items) => [...items].sort((a, b) => compareAlphanumeric(a.id, b.id))

const sortNodes = (items) =>
  [...items].sort((a, b) => {
    const byGateway = compareAlphanumeric(a.gatewayId, b.gatewayId)
    if (byGateway !== 0) {
      return byGateway
    }
    return compareAlphanumeric(a.nodeId, b.nodeId)
  })

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
  selectedGatewayIds,
  selectedNodeKeys,
  showSources,
  showFields,
  includeChildNodes,
  selectedLayout,
}) => {
  const selectedGwIds = normalizeGatewayIds(selectedGatewayIds)
  const visibleGateways = sortGateways(
    selectedGwIds.length === 0
      ? gateways
      : gateways.filter((gw) => selectedGwIds.indexOf(String(gw.id)) !== -1)
  )
  let visibleNodes = sortNodes(
    selectedGwIds.length === 0
      ? nodes
      : nodes.filter((n) => selectedGwIds.indexOf(String(n.gatewayId)) !== -1)
  )
  const selectedKeys = normalizeNodeKeys(selectedNodeKeys)
  if (selectedKeys.length) {
    const selected = visibleNodes.filter((n) => selectedKeys.indexOf(nodeKey(n)) !== -1)
    const extra = includeChildNodes ? collectChildNodesForKeys(visibleNodes, selectedKeys) : []
    visibleNodes = selected.concat(extra)
  }
  const visibleNodeKeys = new Set(visibleNodes.map(nodeKey))
  const gatewayAllowed = (gatewayId) =>
    selectedGwIds.length === 0 || selectedGwIds.indexOf(String(gatewayId)) !== -1
  const nodeAllowed = (item) => !selectedKeys.length || visibleNodeKeys.has(nodeKey(item))

  // Keep orphan sources/fields. They attach to the nearest parent that is on the graph.
  const visibleSources = showSources
    ? sources.filter((s) => gatewayAllowed(s.gatewayId) && nodeAllowed(s))
    : []

  const visibleFields = showFields
    ? fields.filter((f) => gatewayAllowed(f.gatewayId) && nodeAllowed(f))
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
    const nodeId = `node-${String(s.gatewayId)}-${String(s.nodeId)}`
    const gatewayId = `gw-${String(s.gatewayId)}`
    const parent = itemIds.has(nodeId) ? nodeId : gatewayId
    addEdge(`edge-src-${index}`, parent, id, EdgeStyle.default)
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
    const sourceId = `src-${String(f.gatewayId)}-${String(f.nodeId)}-${String(f.sourceId)}`
    const nodeId = `node-${String(f.gatewayId)}-${String(f.nodeId)}`
    const gatewayId = `gw-${String(f.gatewayId)}`
    const parent = itemIds.has(sourceId) ? sourceId : itemIds.has(nodeId) ? nodeId : gatewayId
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
      selectedGatewayIds: prefs.selectedGatewayIds,
      selectedNodeKeys: prefs.selectedNodeKeys,
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
    this.needsLayout = true
    this.layoutAttempts = 0
    this.hadValidSize = false
    this.fitTimer = null
    this.resizeObserver = null
    this.surfaceHost = null
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
          selectedGatewayIds: this.state.selectedGatewayIds,
          selectedNodeKeys: this.state.selectedNodeKeys,
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
    window.addEventListener("keydown", onAltKeyChange, true)
    window.addEventListener("keyup", onAltKeyChange, true)
    window.addEventListener("blur", onAltLost)
    document.addEventListener("visibilitychange", onAltVisibilityChange)
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
        const gateways = sortGateways(listPayload(values[0]))
        const nodes = sortNodes(listPayload(values[1]))
        const availableGw = new Set(gateways.map((gw) => String(gw.id)))
        const selectedGatewayIds = prefs.selectedGatewayIds.filter((id) => availableGw.has(String(id)))
        const allGateways = selectedGatewayIds.length === 0
        const availableKeys = new Set(nodes.map(nodeKey))
        const selectedNodeKeys = allGateways
          ? []
          : prefs.selectedNodeKeys.filter((key) => availableKeys.has(key))
        this.setState(
          {
            gateways,
            nodes,
            selectedGatewayIds,
            selectedNodeKeys,
            showSources: allGateways ? false : prefs.showSources,
            showFields: allGateways ? false : prefs.showFields,
            includeChildNodes:
              allGateways || selectedNodeKeys.length === 0 ? false : prefs.includeChildNodes,
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

  filtersForSelection = (gatewayIds, nodeKeys, includeChildNodes, nodes) => {
    const filters = []
    const gwIds = normalizeGatewayIds(gatewayIds)
    if (gwIds.length === 1) {
      filters.push(eqFilter("gatewayId", gwIds[0]))
    } else if (gwIds.length > 1) {
      filters.push({ k: "gatewayId", o: "in", v: gwIds })
    }
    const keys = normalizeNodeKeys(nodeKeys)
    if (keys.length) {
      const nodeIds = keys.map((key) => String(key.split("::")[1]))
      if (includeChildNodes) {
        collectChildNodesForKeys(nodes || [], keys).forEach((n) => nodeIds.push(String(n.nodeId)))
      }
      const uniqueIds = [...new Set(nodeIds)]
      if (uniqueIds.length === 1) {
        filters.push(eqFilter("nodeId", uniqueIds[0]))
      } else {
        filters.push({ k: "nodeId", o: "in", v: uniqueIds })
      }
    }
    return filters
  }

  refreshDetails = (nextState) => {
    const requestId = this.detailsRequestId + 1
    this.detailsRequestId = requestId
    const gatewayIds = nextState.selectedGatewayIds
    const nodeKeys = nextState.selectedNodeKeys
    const showSources = nextState.showSources
    const showFields = nextState.showFields
    const includeChildNodes = nextState.includeChildNodes
    const apply = (patch) => {
      if (!this.alive || requestId !== this.detailsRequestId) {
        return
      }
      this.setState(patch)
    }
    if (!normalizeGatewayIds(gatewayIds).length) {
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
    const filters = this.filtersForSelection(gatewayIds, nodeKeys, includeChildNodes, nextState.nodes)
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
      prevState.selectedGatewayIds !== this.state.selectedGatewayIds ||
      prevState.selectedNodeKeys !== this.state.selectedNodeKeys ||
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
    if (this.fitTimer) {
      window.clearTimeout(this.fitTimer)
      this.fitTimer = null
    }
    this.layoutTimer = window.setTimeout(() => {
      this.layoutTimer = null
      if (!this.topologyController) {
        return
      }
      const graph = this.topologyController.getGraph()
      if (!graph) {
        return
      }
      if (!graphHasSize(graph)) {
        this.hadValidSize = false
        this.layoutAttempts += 1
        if (this.layoutAttempts < 25) {
          this.scheduleLayout()
        }
        return
      }
      this.hadValidSize = true
      this.layoutAttempts = 0
      action(() => {
        graph.setScaleExtent([0.35, 3])
        try {
          graph.layout()
        } catch (_err) {
          // Force can throw if a node is mid-attach; retry once after a paint.
          window.requestAnimationFrame(() => {
            if (!this.topologyController) {
              return
            }
            try {
              this.topologyController.getGraph().layout()
            } catch (_retryErr) {
              // leave whatever positions we have
            }
          })
        }
        if (this.state.selectedLayout === LAYOUT_DAGRE) {
          applyReadableView(graph)
          this.needsLayout = false
        } else {
          this.fitTimer = window.setTimeout(() => {
            this.fitTimer = null
            if (!this.topologyController) {
              return
            }
            action(() => {
              const next = this.topologyController.getGraph()
              if (next && graphHasSize(next)) {
                applyReadableView(next)
                this.needsLayout = false
              }
            })()
          }, 600)
        }
      })()
    }, this.layoutAttempts === 0 ? 50 : 100)
  }

  setSurfaceHost = (el) => {
    if (this.surfaceHost === el) {
      return
    }
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
      this.resizeObserver = null
    }
    this.surfaceHost = el
    if (!el || typeof ResizeObserver === "undefined") {
      return
    }
    this.resizeObserver = new ResizeObserver(() => {
      if (!this.needsLayout || !this.topologyController) {
        return
      }
      const graph = this.topologyController.getGraph()
      const ready = graphHasSize(graph)
      if (ready && !this.hadValidSize) {
        this.scheduleLayout()
      }
      if (!ready) {
        this.hadValidSize = false
      }
    })
    this.resizeObserver.observe(el)
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
    try {
      const graph = this.topologyController.hasGraph && this.topologyController.hasGraph()
        ? this.topologyController.getGraph()
        : null
      const layout = graph && graph.getLayout && graph.getLayout()
      if (layout && typeof layout.stopSimulation === "function") {
        layout.stopSimulation()
      }
    } catch (_err) {
      // previous force run may already be torn down
    }
    this.topologyController.fromModel(model, true)
    if (structureChanged && model.nodes.length) {
      this.needsLayout = true
      this.layoutAttempts = 0
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
    if (this.fitTimer) {
      window.clearTimeout(this.fitTimer)
    }
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
      this.resizeObserver = null
    }
    window.removeEventListener("keydown", onAltKeyChange, true)
    window.removeEventListener("keyup", onAltKeyChange, true)
    window.removeEventListener("blur", onAltLost)
    document.removeEventListener("visibilitychange", onAltVisibilityChange)
    altHeld = false
    this.props.unloadData({ key: WS_KEY })
  }

  visibleNodesForFilter = () => {
    const { nodes, selectedGatewayIds } = this.state
    const gwIds = normalizeGatewayIds(selectedGatewayIds)
    if (!gwIds.length) {
      return nodes
    }
    const allowed = new Set(gwIds)
    return nodes.filter((n) => allowed.has(String(n.gatewayId)))
  }

  onGatewaySelect = (selection) => {
    const selectedGatewayIds = normalizeGatewayIds(selection)
    const goingToAll = selectedGatewayIds.length === 0
    this.setState(
      {
        selectedGatewayIds,
        selectedNodeKeys: [],
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
    const selectedNodeKeys = normalizeNodeKeys(selection)
    this.setState(
      {
        selectedNodeKeys,
        includeChildNodes: selectedNodeKeys.length === 0 ? false : this.state.includeChildNodes,
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

  onDownloadImage = () => {
    const host = this.surfaceHost
    const svg = host && host.querySelector("svg")
    if (!svg) {
      return
    }
    const background =
      (host && window.getComputedStyle(host).backgroundColor) ||
      window.getComputedStyle(document.body).backgroundColor ||
      "#f0f0f0"
    const stamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")
    downloadSvg(svg, background, `topology-${stamp}.svg`)
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
            gateways: sortGateways(listPayload(values[0])),
            nodes: sortNodes(listPayload(values[1])),
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
      selectedGatewayIds,
      selectedNodeKeys,
      selectedLayout,
      includeChildNodes,
      showSources,
      showFields,
      sources,
      fields,
      loading,
    } = this.state
    const gatewayOptions = []
    gateways.forEach((gw) => {
      const count = nodes.filter((n) => String(n.gatewayId) === String(gw.id)).length
      const name = gw.name || ""
      gatewayOptions.push({
        value: String(gw.id),
        label: `${gw.id} (${count})`,
        toggleLabel: String(gw.id),
        name,
        description: name && name !== gw.id ? name : gw.description || undefined,
        searchText: `${gw.id} ${name} ${gw.description || ""}`,
      })
    })

    const detailsDisabled = loading || selectedGatewayIds.length === 0
    const childNodesDisabled = detailsDisabled || selectedNodeKeys.length === 0
    const nodeOptions = []
    this.visibleNodesForFilter().forEach((n) => {
      const name = n.name || ""
      const id = String(n.nodeId)
      const description = [name && name !== id ? name : "", selectedGatewayIds.length !== 1 ? n.gatewayId : ""]
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
                  selected={selectedGatewayIds}
                  onChange={this.onGatewaySelect}
                  isDisabled={loading || gateways.length === 0}
                  isSearchable
                  isMulti
                  isArrayData
                  maxHeight="300px"
                  label={t("all_gateways")}
                />
              </div>
            </ToolbarItem>
            <ToolbarItem>
              <div className="topology-filter-select">
                <Select
                  options={nodeOptions}
                  selected={selectedNodeKeys}
                  onChange={this.onNodeSelect}
                  isDisabled={detailsDisabled || this.visibleNodesForFilter().length === 0}
                  isSearchable
                  isMulti
                  isArrayData
                  maxHeight="300px"
                  label={t("all_nodes")}
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
            <ToolbarItem>
              <Tooltip content={t("download")}>
                <Button
                  variant="secondary"
                  isSmall
                  onClick={this.onDownloadImage}
                  isDisabled={loading}
                  aria-label={t("download")}
                >
                  <DownloadIcon />
                </Button>
              </Tooltip>
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
    const { selectedGatewayIds } = this.state
    const titleKey =
      selectedGatewayIds.length > 0 ? "no_topology_nodes_for_gateway" : "no_topology_nodes"
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
      selectedGatewayIds,
      selectedNodeKeys,
      sidebarKind,
      sidebarResource,
      selectedGraphId,
    } = this.state
    const selectedGwIds = normalizeGatewayIds(selectedGatewayIds)
    let visibleNodes =
      selectedGwIds.length === 0
        ? nodes
        : nodes.filter((n) => selectedGwIds.indexOf(String(n.gatewayId)) !== -1)
    if (selectedNodeKeys.length) {
      const selected = new Set(selectedNodeKeys)
      visibleNodes = visibleNodes.filter((n) => selected.has(nodeKey(n)))
    }
    const visibleGatewayCount =
      selectedGwIds.length === 0
        ? gateways.length
        : gateways.filter((gw) => selectedGwIds.indexOf(String(gw.id)) !== -1).length
    const hasGraph = !loading && (visibleGatewayCount > 0 || visibleNodes.length > 0)

    return (
      <React.Fragment>
        <PageTitle title="topology" />
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
                    <div className="topology-surface-host" ref={this.setSurfaceHost}>
                      <VisualizationSurface
                        state={{ selectedIds: selectedGraphId ? [selectedGraphId] : [] }}
                      />
                    </div>
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
