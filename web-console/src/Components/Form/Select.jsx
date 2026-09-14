import React from "react"
import PropTypes from "prop-types"
import { Select as PfSelect, SelectOption } from "@patternfly/react-core"

class Select extends React.Component {
  state = {
    isOpen: false,
  }

  onToggle = (isOpen) => {
    this.setState({
      isOpen,
    })
  }

  onSelect = (_event, selectionLabel, _isPlaceholder) => {
    const { onChange, isMulti, isArrayData, options, selected } = this.props
    const isOpen = isMulti ? true : false
    this.setState({ isOpen: isOpen }, () => {
      if (onChange) {
        // get value with label
        const selectionValue = getValueByLabel(options, selectionLabel)
        let finalValue = selectionValue
        if (isMulti) {
          let itemsSelected = isArrayData
            ? [...(Array.isArray(selected) ? selected : [])]
            : String(selected || "").split(",")
          if (itemsSelected.includes(selectionValue)) {
            itemsSelected = itemsSelected.filter((v) => v !== selectionValue)
          } else {
            itemsSelected.push(selectionValue)
          }
          const uniqueValues = [...new Set(itemsSelected)].filter((v) => v !== "")
          finalValue = isArrayData ? uniqueValues : uniqueValues.join(",")
        }
        onChange(finalValue)
      }
    })
  }

  onCreateOption = (_newValue) => {}

  clearSelection = () => {
    const { onChange, isArrayData } = this.props
    this.setState({ isOpen: false }, () => {
      if (onChange) {
        if (isArrayData) {
          onChange([])
        } else {
          onChange("")
        }
      }
    })
  }

  render() {
    const {
      label,
      options,
      selected,
      isDisabled,
      isCreatable,
      variant,
      disableClear,
      hideDescription,
      isArrayData,
      direction = "",
    } = this.props
    const { isOpen } = this.state

    // get label with value
    let selections = []

    if (isArrayData) {
      const selectedArr = Array.isArray(selected) ? selected : []
      selections = selectedArr.map((s) => {
        return getLabelByValue(options, s)
      })
    } else {
      if (selected !== undefined && selected !== "") {
        selections = selected.split(",").map((s) => {
          return getLabelByValue(options, s)
        })
      }
    }

    const selectOptions = options.map((option, index) => (
      <SelectOption
        isDisabled={option.disabled}
        key={index}
        value={option.label}
        {...(!hideDescription && option.description && { description: option.description })}
      />
    ))

    // console.log("selections:", selections)

    return (
      <PfSelect
        variant={variant}
        //typeAheadAriaLabel="Select a state"
        onToggle={this.onToggle}
        onSelect={this.onSelect}
        onClear={disableClear ? undefined : this.clearSelection}
        selections={selections}
        isOpen={isOpen}
        //aria-labelledby={titleId}
        placeholderText={label}
        isDisabled={isDisabled}
        isCreatable={isCreatable}
        direction={direction}
        //onCreateOption={(hasOnCreateOption && this.onCreateOption) || undefined}
      >
        {selectOptions}
      </PfSelect>
    )
  }
}

Select.propTypes = {
  title: PropTypes.string,
  options: PropTypes.array, // [{value: "abc", label:"ABC text", disabled: false}]
  defaultValue: PropTypes.string,
  onSelectionFunc: PropTypes.func,
  disabled: PropTypes.bool,
  isMulti: PropTypes.bool,
  isArrayData: PropTypes.bool,
  direction: PropTypes.string,
}

export default Select

// helper functions

const getValueByLabel = (items, label) => {
  for (let index = 0; index < items.length; index++) {
    const item = items[index]
    if (label === item.label) {
      return item.value
    }
  }
  return ""
}

const getLabelByValue = (items, value) => {
  for (let index = 0; index < items.length; index++) {
    const item = items[index]
    if (value === item.value) {
      return item.label
    }
  }
  return ""
}
