import { Button, Card, CardBody, CardTitle, Divider, Split, SplitItem } from "@patternfly/react-core"
import { Remove2Icon } from "@patternfly/react-icons"
import React from "react"
import { RefreshButton } from "../../../Components/Buttons/Buttons"
import { ClearSleepingQueueDialog } from "../../../Components/Dialog/Dialog"
import LastUpdate from "../../../Components/LastUpdate/LastUpdate"
import Loading from "../../../Components/Loading/Loading"
import { api } from "../../../Service/Api"
import SleepingMessage from "./SleepingMessage"
import PropTypes from "prop-types"

class SleepingQueue extends React.Component {
  state = {
    loading: true,
    resource: {},
    messages: [],
    cleaningQueue: false,
    showClearQueueDialog: false,
    lastUpdate: null,
    ids: {},
  }
  componentDidMount() {
    this.loadData()
  }

  loadData = () => {
    const { isGateway, id } = this.props
    const resourceApi = isGateway ? api.gateway.get : api.node.get
    resourceApi(id)
      .then((res) => {
        const resource = res.data
        const ids = {}
        if (isGateway) {
          ids["gatewayId"] = resource.id
          ids["nodeId"] = ""
        } else {
          ids["gatewayId"] = resource.gatewayId
          ids["nodeId"] = resource.nodeId
        }
        api.gateway
          .getSleepingQueue(ids.gatewayId, ids.nodeId)
          .then((sRes) => {
            const allMessages = sRes.data
            const messages = []
            if (isGateway) {
              Object.keys(allMessages).forEach((k) => {
                if (allMessages[k].length > 0) {
                  messages.push(...allMessages[k])
                }
              })
            } else {
              messages.push(...allMessages)
            }
            this.setState({
              loading: false,
              resource: resource,
              messages: messages,
              cleaningQueue: false,
              lastUpdate: new Date(),
              ids: ids,
            })
          })
          .catch((_e) => {
            this.setState({ loading: false, cleaningQueue: false, lastUpdate: new Date() })
          })
      })
      .catch((_e) => {
        this.setState({ loading: false, cleaningQueue: false, lastUpdate: new Date() })
      })
  }

  clearQueue = (gatewayId, nodeId) => {
    this.setState({ cleaningQueue: true, showClearQueueDialog: false })
    api.gateway
      .clearSleepingQueue(gatewayId, nodeId)
      .then((_res) => {
        this.loadData()
      })
      .catch((_e) => {
        this.setState({ cleaningQueue: false })
      })
  }

  showClearQueueDialog = () => {
    this.setState({ showClearQueueDialog: true })
  }

  hideClearQueueDialog = () => {
    this.setState({ showClearQueueDialog: false })
  }

  render() {
    const { loading, resource, ids, messages, cleaningQueue, showClearQueueDialog, lastUpdate } = this.state
    const { t = () => {}, isGateway } = this.props
    if (loading) {
      return <Loading key={"sleeping-queue"} />
    }
    const clearQueueButton =
      messages.length > 0 ? (
        <Button
          onClick={this.showClearQueueDialog}
          isSmall
          variant="link"
          isDanger={false}
          isLoading={cleaningQueue}
          key="clear_queue"
        >
          {cleaningQueue ? (
            `${t("in_progress")}...`
          ) : (
            <>
              <Remove2Icon /> {t("clear")}
            </>
          )}
        </Button>
      ) : null

    const sleepingMessage =
      messages.length > 0 ? (
        <SleepingMessage messages={messages} />
      ) : (
        <div style={{ paddingTop: "5px" }}>{t("no_data")}</div>
      )

    return (
      <>
        <ClearSleepingQueueDialog
          show={showClearQueueDialog}
          onCloseFn={this.hideClearQueueDialog}
          onOkFn={() => this.clearQueue(ids.gatewayId, ids.nodeId)}
        />
        <Card key={resource.id} isPlain isCompact>
          <CardTitle style={{ paddingBottom: "0px", paddingTop: "5px" }}>
            <Split>
              <SplitItem isFilled>
                <span style={{ fontWeight: 400, color: "gray" }}>
                  {isGateway ? t("gateway") : t("node")}:
                </span>{" "}
                {isGateway ? resource.id : resource.nodeId} /{" "}
                {isGateway ? resource.description : resource.name}
              </SplitItem>
              <SplitItem>
                {clearQueueButton}
                <RefreshButton key="refresh_btn" onClick={this.loadData} />
              </SplitItem>
            </Split>
            <Divider />
          </CardTitle>
          <CardBody>
            {sleepingMessage}
            <LastUpdate time={lastUpdate} />
          </CardBody>
        </Card>
      </>
    )
  }
}

SleepingQueue.propTypes = {
  isGateway: PropTypes.bool, // if true: fetch messages from gateway, if false: get from a node
  id: PropTypes.string, // isGateway: true, it is gateway id, otherwise node id (it is 'id', NOT a 'nodeId')
  t: PropTypes.func, // translate func
}

export default SleepingQueue
