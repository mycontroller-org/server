import React from "react"
import { ResetButton } from "../../../Components/Buttons/Buttons"
import { CommonDialog } from "../../../Components/Dialog/Dialog"
import { api } from "../../../Service/Api"

class ResetJwtSecret extends React.Component {
  state = {
    resetting: false,
    showDialog: false,
  }

  performReset = () => {
    this.setState({ showDialog: false }, () => {
      api.settings
        .resetJwtSecret()
        .then((_res) => {
          this.setState({
            resetting: false,
          })
        })
        .catch((_e) => {
          this.setState({ resetting: false })
        })
    })
  }

  showDialog = () => {
    this.setState({ showDialog: true })
  }

  hideDialog = () => {
    this.setState({ showDialog: false })
  }

  render() {
    const { showDialog, resetting } = this.state
    return (
      <>
        <CommonDialog
          dialogTitle={"dialog.reset_title_jwt_secret"}
          dialogMsg="dialog.reset_msg_jwt_secret"
          confirmBtnText="reset"
          show={showDialog}
          onCloseFn={this.hideDialog}
          onOkFn={this.performReset}
        />
        <ResetButton
          text={resetting ? "resetting" : "reset"}
          isDisabled={resetting}
          onClick={this.showDialog}
        />
      </>
    )
  }
}

export default ResetJwtSecret
