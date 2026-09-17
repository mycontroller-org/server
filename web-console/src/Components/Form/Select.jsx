import React from "react"
import PropTypes from "prop-types"
import { Select as PfSelect, SelectOption, SelectVariant } from "@patternfly/react-core"

class Select extends React.Component {
  state = {
    isOpen: false,
  }

  onToggle = (isOpen) => {
    this.setState({ isOpen })
  }

  onSelect = (_event, selectionLabel, _isPlaceholder) => {
    const { onChange, isMulti, isArrayData, options, selected, variant } = this.props
    const checkboxMulti = isCheckboxMulti(isMulti, variant)
    const isOpen = isMulti ? true : false
    this.setState({ isOpen: isOpen }, () => {
      if (onChange) {
        const selectionValue = checkboxMulti ? selectionLabel : getValueByLabel(options, selectionLabel)
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

  onFilter = (_event, text) => {
    const { options, hideDescription, isMulti, variant } = this.props
    const checkboxMulti = isCheckboxMulti(isMulti, variant)
    const query = String(text || "").toLowerCase()
    return options
      .filter((option) => {
        if (!query) {
          return true
        }
        const haystack = [option.label, option.value, option.name, option.description, option.searchText]
          .filter(Boolean)
          .join(" ")
          .toLowerCase()
        return haystack.includes(query)
      })
      .map((option, index) => (
        <SelectOption
          isDisabled={option.disabled}
          key={option.value != null ? String(option.value) : index}
          value={checkboxMulti ? option.value : option.label}
          {...(!hideDescription && option.description && { description: option.description })}
        >
          {checkboxMulti ? option.label : undefined}
        </SelectOption>
      ))
  }

  onCreateOption = (_newValue) => {}

  clearSelection = () => {
    const { onChange, isArrayData } = this.props
    this.setState({ isOpen: false }, () => {
      if (onChange) {
        onChange(isArrayData ? [] : "")
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
      isSearchable,
      disableClear,
      hideDescription,
      isArrayData,
      isMulti,
      direction = "",
      maxHeight = "300px",
    } = this.props
    const { isOpen } = this.state
    const checkboxMulti = isCheckboxMulti(isMulti, variant)

    // get label with value
    let selections = []

    if (checkboxMulti) {
      selections = isArrayData
        ? Array.isArray(selected)
          ? selected
          : []
        : String(selected || "")
            .split(",")
            .filter((v) => v !== "")
    } else if (isArrayData) {
      const selectedArr = Array.isArray(selected) ? selected : []
      selections = selectedArr.map((s) => getLabelByValue(options, s))
    } else if (selected !== undefined && selected !== "") {
      selections = selected.split(",").map((s) => getLabelByValue(options, s))
    }

    const selectOptions = options.map((option, index) => (
      <SelectOption
        isDisabled={option.disabled}
        key={option.value != null ? String(option.value) : index}
        value={checkboxMulti ? option.value : option.label}
        {...(!hideDescription && option.description && { description: option.description })}
      >
        {checkboxMulti ? option.label : undefined}
      </SelectOption>
    ))

    let placeholder = label
    if (checkboxMulti && selections.length) {
      const selectedLabels = selections.map((value) => {
        const option = options.find((item) => String(item.value) === String(value))
        if (!option) {
          return String(value)
        }
        return option.toggleLabel || option.label || String(value)
      })
      placeholder = selectedLabels.join(", ")
    }

    return (
      <PfSelect
        variant={
          variant ||
          (checkboxMulti
            ? SelectVariant.checkbox
            : isSearchable
              ? SelectVariant.typeahead
              : SelectVariant.single)
        }
        onFilter={isSearchable && !checkboxMulti ? this.onFilter : undefined}
        hasInlineFilter={!!(checkboxMulti && isSearchable)}
        isCheckboxSelectionBadgeHidden={!!checkboxMulti}
        onToggle={this.onToggle}
        onSelect={this.onSelect}
        onClear={disableClear ? undefined : this.clearSelection}
        selections={selections}
        isOpen={isOpen}
        maxHeight={maxHeight}
        placeholderText={placeholder}
        isDisabled={isDisabled}
        isCreatable={isCreatable}
        direction={direction}
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

const isCheckboxMulti = (isMulti, variant) =>
  !!isMulti && (!variant || variant === SelectVariant.checkbox)

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
