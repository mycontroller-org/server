import { ToggleGroup, ToggleGroupItem } from "@patternfly/react-core"
import React from "react"

const ToggleButtonGroup = ({
  selected = "",
  options = [],
  onSelectionFunc = () => {},
  isDisabled = false,
}) => {
  const toggleItems = options.map((item) => {
    // item = [{value: "abc", label:"ABC text", disabled: false}]
    return (
      <ToggleGroupItem
        text={item.label}
        buttonId={item.value}
        isSelected={selected === item.value}
        isDisabled={item.disabled}
        onChange={(_isSelected, event) => onSelectionFunc(event.currentTarget.id)}
      />
    )
  })
  return <ToggleGroup variant="default">{toggleItems}</ToggleGroup>
}

export default ToggleButtonGroup
