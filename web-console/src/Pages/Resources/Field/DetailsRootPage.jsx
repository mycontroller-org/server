import React from "react"
import DetailRootPage from "../../../Components/BasePage/DetailsBase"
import TabDetails from "./Details"
import UpdatePage from "./UpdatePage"

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
    const { id } = this.props.match.params
    const { history, match } = this.props
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
    ]
    return (
      <DetailRootPage
        pageHeader="field_details"
        tabs={tabs}
        activeTabKey={activeTabKey}
        onTabClickFn={this.onTabClick}
      />
    )
  }
}

export default DetailPage
