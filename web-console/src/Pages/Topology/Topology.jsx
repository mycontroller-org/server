import React from "react"
import { Spinner } from "@patternfly/react-core"
import "./Topology.scss"
import {
  DefaultEdge,
  DefaultGroup,
  DefaultNode,
  EdgeStyle,
  ForceLayout,
  GraphComponent,
  ModelKind,
  NodeShape,
  NodeStatus,
  SELECTION_EVENT,
  TopologyControlBar,
  TopologyView,
  Visualization,
  VisualizationProvider,
  VisualizationSurface,
  action,
  createTopologyControlButtons,
  defaultControlButtonsOptions,
  withPanZoom,
} from "@patternfly/react-topology"

import PageContent from "../../Components/PageContent/PageContent"
import PageTitle from "../../Components/PageTitle/PageTitle"
import { PlugIcon as GatewayIcon } from "@patternfly/react-icons"
import { ContainerNodeIcon as NodeIcon } from "@patternfly/react-icons"
import { getValue } from "../../Util/Util"
import { api } from "../../Service/Api"

const baselineLayoutFactory = (type, graph) => {
  return new ForceLayout(graph, {
    layoutOnDrag: false,
  })
  // switch (type) {
  //   case "Cola":
  //     return new ColaLayout(graph)
  //   default:
  //     return new ColaGroupsLayout(graph, {
  //       layoutOnDrag: false,
  //     })
  // }
}

const BadgeColors = [
  {
    name: "A",
    badgeColor: "#ace12e",
    badgeTextColor: "#0f280d",
    badgeBorderColor: "#486b00",
  },
  {
    name: "B",
    badgeColor: "#F2F0FC",
    badgeTextColor: "#5752d1",
    badgeBorderColor: "#CBC1FF",
  },
]

const CustomNode = ({ element }) => {
  const data = getValue(element, "data", "")
  const Icon = element.type == "gateway" ? GatewayIcon : NodeIcon
  const badgeColors = BadgeColors.find((badgeColor) => badgeColor.name === data.badge)
  return (
    <DefaultNode
      element={element}
      showStatusDecorator={false}
      badge={data.badge}
      badgeColor={badgeColors?.badgeColor}
      badgeTextColor={badgeColors?.badgeTextColor}
      badgeBorderColor={badgeColors?.badgeBorderColor}
    >
      <g transform={`translate(16,16)`}>
        <Icon
          style={{
            color: "#393F44",
          }}
          width={32}
          height={32}
        />
      </g>
    </DefaultNode>
  )
}

const baselineComponentFactory = (kind, type) => {
  switch (type) {
    case "group":
      return DefaultGroup
    default:
      switch (kind) {
        case ModelKind.graph:
          return withPanZoom()(GraphComponent)
        case ModelKind.node:
          return CustomNode
        case ModelKind.edge:
          return DefaultEdge
        default:
          return undefined
      }
  }
}

const NODE_DIAMETER = 64

class TopologyPage extends React.Component {
  state = {
    loading: true,
    nData: [],
    isOpen: false,
    selected: "",
    searchValue: "",
    filteredItems: [],
    detailsShown: false,
    topologyController: {},
    gateways: [],
    nodes: [],
  }

  setSelectedIds = (_id) => {}

  componentDidMount() {
    this.topologyController = new Visualization()
    this.topologyController.registerLayoutFactory(baselineLayoutFactory)
    this.topologyController.registerComponentFactory(baselineComponentFactory)
    this.topologyController.addEventListener(SELECTION_EVENT, this.setSelectedIds)
    Promise.all([api.gateway.list(), api.node.list()]).then((_values) => {
      this.setState({
        gateways: _values[0].data.data,
        nodes: _values[1].data.data,
        loading: false,
      })
    })
  }

  render() {
    const { loading, gateways = [], nodes = [] } = this.state

    const elements = []
    if (loading) {
      elements.push(<Spinner key="loading" size="md" />)
    } else {
      const items = []
      const edges = []

      gateways.forEach((_item, _index) => {
        const itemStatus = getValue(_item, "state.status", "unknown")
        items.push({
          id: `gw-${_item.id}`,
          type: "gateway",
          label: _item.id,
          width: NODE_DIAMETER,
          height: NODE_DIAMETER,
          shape: NodeShape.hexagon,
          status:
            itemStatus == "up"
              ? NodeStatus.success
              : itemStatus == "down"
              ? NodeStatus.danger
              : NodeStatus.default,
        })
      })

      nodes.forEach((_item, _index) => {
        const itemStatus = getValue(_item, "state.status", "unknown")
        items.push({
          id: `node-${_item.gatewayId}-${_item.nodeId}`,
          type: "node",
          label: _item.nodeId,
          width: NODE_DIAMETER,
          height: NODE_DIAMETER,
          shape: NodeShape.circle,
          status:
            itemStatus == "up"
              ? NodeStatus.success
              : itemStatus == "down"
              ? NodeStatus.danger
              : NodeStatus.default,
        })
        // add gateway to node mapping
        // update parentId if available
        if (
          _item.labels &&
          _item.labels.parent_id &&
          _item.labels.parent_id != ""
        ) {
          edges.push({
            id: `edge-${_index}`,
            type: "edge",
            source: `node-${_item.gatewayId}-${_item.labels.parent_id}`,
            target: `node-${_item.gatewayId}-${_item.nodeId}`,
            edgeStyle: EdgeStyle.default,
          })
        } else {
          edges.push({
            id: `edge-${_index}`,
            type: "edge",
            source: `gw-${_item.gatewayId}`,
            target: `node-${_item.gatewayId}-${_item.nodeId}`,
            edgeStyle: EdgeStyle.default,
          })
        }
      })

      const model = {
        nodes: items,
        edges: edges,
        graph: {
          id: "g1",
          type: "graph",
          layout: "Cola",
        },
      }
      this.topologyController.fromModel(model, false)

      elements.push(
        <TopologyView
          key="tview"
          className="ws-react-c-topology"
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
                  this.topologyController.getGraph().reset()
                  this.topologyController.getGraph().layout()
                }),
                legend: false,
              })}
            />
          }
        >
          <VisualizationProvider controller={this.topologyController}>
            <VisualizationSurface state={{}} />
          </VisualizationProvider>
        </TopologyView>
      )
    }

    return (
      <React.Fragment>
        <PageTitle title="Topology" />
        <PageContent>{elements}</PageContent>
      </React.Fragment>
    )
  }
}

export default TopologyPage
