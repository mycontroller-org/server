import { PageSection, PageSectionVariants } from "@patternfly/react-core"
import React from "react"
import "./PageContent.scss"

const PageContent = ({ children, hasNoPaddingTop = false }) => {
  const className = hasNoPaddingTop ? "has-no-padding-top" : ""
  return (
    <div className="mc-page-content">
      <PageSection className={className} variant={PageSectionVariants.light}>
        {children}
      </PageSection>
    </div>
  )
}

export default PageContent
