import {
  Divider,
  Grid,
  GridItem,
  PageSection,
  PageSectionVariants,
  Text,
  Title,
} from "@patternfly/react-core"
import React from "react"
import { useTranslation } from "react-i18next"
import "./PageTitle.scss"

const PageTitle = ({ title, description, actions, hideDivider = false }) => {
  const { t } = useTranslation()

  const divider = hideDivider ? null : (
    <GridItem span={12}>
      <Divider component="hr" />
    </GridItem>
  )

  // if non string title passed, do not add heading style
  const titleComponent =
    typeof title === "string" ? (
      <Title headingLevel="h2" size="xl">
        {t(title)}
      </Title>
    ) : (
      title
    )
  return (
    <PageSection className="mc-page-title" variant={PageSectionVariants.light}>
      <Grid hasGutter={false}>
        <GridItem span={6}>
          {titleComponent}
          {description ? (
            <Text>
              <small>{t(description)}</small>
            </Text>
          ) : null}
        </GridItem>
        <GridItem span={6}>
          <div style={{ float: "right", marginBottom: "1px" }}>{actions}</div>
        </GridItem>
        {divider}
      </Grid>
    </PageSection>
  )
}

export default PageTitle
