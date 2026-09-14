import React from "react"
import DetailRootPage from "../../../Components/BasePage/DetailsBase"
import SleepingQueue from "../Node/SleepingQueue"
import TabDetails from "./Details"
import UpdatePage from "./UpdatePage"
import { withTranslation } from "react-i18next"

const DefaultTab = "details_0"
class DetailPage extends React.Component {
  state = {
    activeTabKey: "",
  }

  componentDidMount() {
    this.setState({ activeTabKey: DefaultTab })
  }

  onTabClick = (_event, tabKey) => {
    this.setState({ activeTabKey: tabKey })
  }

  updateDefaultTab = () => {
    this.setState({ activeTabKey: DefaultTab })
  }

  render() {
    const { history, match, t } = this.props
    const { id } = this.props.match.params
    const { activeTabKey } = this.state

    const tabs = [
      {
        name: "details",
        content: <TabDetails resourceId={id} history={history} />,
      },
      {
        name: "edit",
        content: <UpdatePage match={match} history={history} cancelFn={this.updateDefaultTab} />,
      },
      {
        name: "sleeping_queue",
        content: <SleepingQueue isGateway={true} id={id} t={t} />,
      },
    ]
    return (
      <DetailRootPage
        pageHeader="gateway_details"
        tabs={tabs}
        activeTabKey={activeTabKey}
        onTabClickFn={this.onTabClick}
      />
    )
  }
}

export default withTranslation()(DetailPage)
