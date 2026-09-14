import React from "react"
import QRCode from "qrcode"

// Renders a service-account token as a QR barcode.
class TokenBarcode extends React.Component {
  canvasRef = React.createRef()

  componentDidMount() {
    this.draw()
  }

  componentDidUpdate(prevProps) {
    if (prevProps.value !== this.props.value) {
      this.draw()
    }
  }

  draw = () => {
    const { value } = this.props
    const canvas = this.canvasRef.current
    if (!canvas || !value) {
      return
    }
    QRCode.toCanvas(canvas, value, {
      width: this.props.size || 220,
      margin: 1,
      errorCorrectionLevel: "M",
    }).catch(() => {})
  }

  render() {
    if (!this.props.value) {
      return null
    }
    return (
      <div style={{ display: "flex", justifyContent: "center", margin: "16px 0" }}>
        <canvas ref={this.canvasRef} />
      </div>
    )
  }
}

export default TokenBarcode
