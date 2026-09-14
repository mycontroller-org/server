import { Alert, Card, CardTitle, CardBody, Grid, Stack, GridItem, Divider, ExpandableSection } from "@patternfly/react-core"
import React from "react"
import PageTitle from "../PageTitle/PageTitle"
import "./ErrorBoundary.scss"

class ErrorBoundary extends React.Component {
  state = {
    error: null,
    errorInfo: null,
  }
  componentDidCatch(error, errorInfo) {
    // Catch errors in any child components and re-renders with an error message
    this.setState({
      error: error,
      errorInfo: errorInfo,
    })
  }

  render() {
    const { error, errorInfo } = this.state
    const { children, hasMargin, hasOffset } = this.props

    if (error) {
      const margin = hasMargin ? { lg: 8, md: 10, sm: 12 } : {}
      const offset = hasOffset ? { lgOffset: 2, mdOffset: 1 } : {}

      const heading = "Oops! Something went wrong :("
      const title = hasMargin ? (
        <PageTitle title={heading} />
      ) : (
        <span>
          <b>{heading}</b>
          <Divider component="li" />
        </span>
      )

      const marginClassName = hasMargin ? "has-margin" : ""

      // Fallback UI if an error occurs
      return (
        <div className="mc-error-boundary">
          <Grid {...margin} className={marginClassName}>
            <GridItem {...offset}>
              {title}
              <Stack hasGutter className="detail">
                <Alert variant="danger" isInline title={error && error.toString()} />
                <ExpandableSection toggleTextExpanded="Show Less" toggleTextCollapsed="Show More">
                  <Card isFlat>
                    <CardTitle>Component Stack Details:</CardTitle>
                    <CardBody>
                      <pre>{errorInfo && errorInfo.componentStack}</pre>
                    </CardBody>
                  </Card>
                </ExpandableSection>
              </Stack>
            </GridItem>
          </Grid>
        </div>
      )
    }
    // component normally just renders children
    return children
  }
}

export default ErrorBoundary
