import {
  Select,
  SelectOption,
  Split,
  SplitItem,
  TextInput,
  ToolbarFilter,
  ToolbarGroup,
} from "@patternfly/react-core"
import { FilterIcon } from "@patternfly/react-icons"
import React from "react"

const Filters = ({
  filters = [],
  selectedCategory = "",
  chips = {},
  deleteChipFunc,
  deleteChipGroupFunc,
  typeToggleFunc,
  onTypeChangeFunc,
  isTypeOpen = false,
  onFilterUpdate,
  t = () => {},
}) => {
  let selectedFilter = null
  for (let index = 0; index < filters.length; index++) {
    const f = filters[index]
    if (f.category === selectedCategory) {
      selectedFilter = f
      break
    }
  }
  const filterList = getFilterList(filters, typeToggleFunc, isTypeOpen, onTypeChangeFunc, selectedCategory, t)
  const inputField = getFilterInputField(selectedFilter, onFilterUpdate)
  const inputComponent = (
    <Split>
      <SplitItem>{filterList}</SplitItem>
      <SplitItem>{inputField}</SplitItem>
    </Split>
  )

  const components = filters.map((f, index) => {
    const { category, categoryName } = f
    return {
      component: ToolbarFilter,
      props: {
        key: index,
        categoryName: categoryName,
        chips: chips[category],
        deleteChip: (_, value) => deleteChipFunc(category, value),
        deleteChipGroup: () => deleteChipGroupFunc(category),
      },
    }
  })

  let rootComponent = null
  const size = components.length - 1
  for (let index = size; index >= 0; index--) {
    const filter = components[index]
    if (index === size) {
      rootComponent = React.createElement(filter.component, { ...filter.props, children: inputComponent })
    } else {
      rootComponent = React.createElement(filter.component, { ...filter.props, children: rootComponent })
    }
  }

  return <ToolbarGroup variant="filter-group">{rootComponent}</ToolbarGroup>
}

const getFilterList = (filters, typeToggleFunc, isTypeOpen, onTypeChangeFunc, selectedCategory, t) => {
  let selectedCategoryName = selectedCategory
  const options = filters.map((f) => {
    if (f.category === selectedCategory) {
      selectedCategoryName = f.categoryName
    }
    return (
      <SelectOption key={f.category} onSelect={() => {}} value={f.category}>
        {f.categoryName}
      </SelectOption>
    )
  })
  return (
    <Select
      onToggle={typeToggleFunc}
      isOpen={isTypeOpen}
      onSelect={onTypeChangeFunc}
      placeholderText={
        <span>
          <FilterIcon /> {selectedCategoryName ? selectedCategoryName : t("filter")}
        </span>
      }
    >
      {options}
    </Select>
  )
}

const getFilterInputField = (f, onFilterUpdate) => {
  if (!f) {
    return null
  }
  switch (f.fieldType) {
    case "input":
    case "label":
      return (
        <TextFilterComponent
          fieldType={f.fieldType}
          category={f.category}
          onFilterUpdate={onFilterUpdate}
        ></TextFilterComponent>
      )
    case "enabled":
      return (
        <SelectFilterComponent
          options={[
            { key: "true", text: "True" },
            { key: "false", text: "False" },
          ]}
          onUpdate={(value) => onFilterUpdate(f.category, value)}
        ></SelectFilterComponent>
      )
    case "state":
      // return options
      break
    default:
    // what should return?
  }
}

class SelectFilterComponent extends React.Component {
  state = {
    show: false,
  }

  hide = (e, value) => {
    this.setState({ show: false }, this.props.onUpdate(value))
  }

  onToggle = () => {
    this.setState((prevState) => {
      return { show: !prevState.show }
    })
  }

  render() {
    const options = this.props.options.map((o, index) => {
      return (
        <SelectOption key={index} value={o.key}>
          {o.text}
        </SelectOption>
      )
    })
    return (
      <Select
        key="random"
        onSelect={this.hide}
        isOpen={this.state.show}
        onToggle={this.onToggle}
        placeholderText="Select an option"
      >
        {options}
      </Select>
    )
  }
}

class TextFilterComponent extends React.Component {
  state = { value: "" }

  submitValue = (e) => {
    if (e.charCode === 13) {
      this.setState({ value: "" }, this.props.onFilterUpdate(this.props.category, e.target.value))
    }
  }

  onChange = (v) => {
    this.setState({ value: v })
  }

  render() {
    const text = this.props.fieldType === "label" ? "key=value" : "Enter a value"
    return (
      <TextInput
        type="text"
        aria-label="text-input-component"
        placeholder={text}
        onKeyPress={this.submitValue}
        onChange={this.onChange}
        value={this.state.value}
      ></TextInput>
    )
  }
}

export default Filters
