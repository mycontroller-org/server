import { ActionList, ActionListGroup, ActionListItem, Button, Level, LevelItem } from "@patternfly/react-core"
import React from "react"

const actionListItem = ({ variant, text, onClickFunc, isDisabled, icon }) => {
  let btnIcon = null
  if (icon) {
    const Icon = icon
    btnIcon = <Icon style={{ marginRight: "5px" }} />
  }
  return (
    <ActionListItem>
      <Button variant={variant} id={text} onClick={onClickFunc} isDisabled={isDisabled}>
        <div>
          {btnIcon}
          {text}
        </div>
      </Button>
    </ActionListItem>
  )
}

const ActionBar = ({ leftBar = [], rightBar = [] }) => {
  const leftActions = leftBar.map((action) => {
    return actionListItem(action)
  })
  const rightActions = rightBar.map((action) => {
    return actionListItem(action)
  })
  return (
    <Level>
      <LevelItem>
        <ActionList>
          <ActionListGroup>{leftActions}</ActionListGroup>
        </ActionList>
      </LevelItem>
      <LevelItem>
        <ActionList>
          <ActionListGroup>{rightActions}</ActionListGroup>
        </ActionList>
      </LevelItem>
    </Level>
  )
}

export default ActionBar
