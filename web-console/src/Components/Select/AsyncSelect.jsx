import { Grid, GridItem, Select, SelectOption, SelectVariant, Spinner } from "@patternfly/react-core"
import PropTypes from "prop-types"
import React from "react"
import { getDynamicFilter } from "../../Util/Filter"

const defaultItemsLimit = 10

class AsyncSelect extends React.Component {
  state = {
    isOpen: false,
    loading: false,
    options: [],
  }

  onFilter = (filterValue = "") => {
    if (this.state.loading) {
      return
    }
    const { apiOptions, optionValueFunc } = this.props
    const valueFunc = optionValueFunc ? optionValueFunc : this.getOptionValueFunc
    if (apiOptions) {
      this.setState({ loading: true }, () => {
        const filters = this.getFilters(filterValue)
        const limit = this.props.limit || defaultItemsLimit
        apiOptions({ filter: filters, limit: limit })
          .then((res) => {
            const items = res.data.data
            const options = items.map((item) => {
              const value = valueFunc(item)
              return { value: value, label: value, description: this.getDescription(item) }
            })
            this.setState({ options: options, loading: false })
          })
          .catch((_e) => {
            this.setState({ loading: false })
          })
      })
    }
  }

  componentDidMount() {
    this.onFilter()
  }

  componentDidUpdate(prevProps) {
    const { isDisabled, apiOptions } = this.props
    if (!isDisabled && isDisabled !== prevProps.isDisabled) {
      this.onFilter()
    }
    if (apiOptions !== prevProps.apiOptions) {
      this.onFilter()
    }
  }

  getOptionValueFunc = (item) => {
    const { optionValueKey } = this.props
    const valueKey = optionValueKey ? optionValueKey : "id"
    return item[valueKey]
  }

  getFilters = (value) => {
    if (!value) {
      return []
    }
    if (this.props.getFiltersFunc) {
      return this.props.getFiltersFunc(value)
    }

    return getDynamicFilter("name", value, [])
    //  return [{ k: "name", o: "regex", v: value }]
  }

  getDescription = (item) => {
    if (this.props.getOptionsDescriptionFunc) {
      return this.props.getOptionsDescriptionFunc(item)
    }
    return item.name
  }

  onSelection = (_event, selection) => {
    const { isMulti, selected, onSelectionFunc } = this.props
    if (isMulti) {
      const current = Array.isArray(selected) ? [...selected] : []
      const next = current.includes(selection)
        ? current.filter((v) => v !== selection)
        : [...current, selection]
      if (onSelectionFunc) {
        onSelectionFunc(next)
      }
      return
    }
    this.setState({ isOpen: false }, () => {
      if (onSelectionFunc) {
        onSelectionFunc(selection)
      }
    })
  }

  onClear = () => {
    const { isMulti, onSelectionFunc } = this.props
    this.onFilter()
    if (onSelectionFunc) {
      onSelectionFunc(isMulti ? [] : "")
    }
  }

  onToggle = (isOpen) => {
    this.setState({ isOpen })
  }

  render() {
    const { isOpen, options, loading } = this.state
    const {
      isDisabled,
      selected,
      showSpinner = false,
      direction = "down",
      isCreatable,
      createText,
      isMulti = false,
    } = this.props
    const selectOptions = options.map((option) => {
      return (
        <SelectOption key={option.value} value={option.value} description={option.description}>
          {option.label}
        </SelectOption>
      )
    })
    const selections = isMulti ? (Array.isArray(selected) ? selected : []) : selected

    const spinnerSpan = showSpinner ? 1 : 0
    const spinner =
      loading && showSpinner ? (
        <GridItem span={spinnerSpan}>
          <Spinner size="lg" />{" "}
        </GridItem>
      ) : null
    return (
      <Grid hasGutter>
        <GridItem span={12 - spinnerSpan}>
          <Select
            variant={isMulti ? SelectVariant.typeaheadMulti : SelectVariant.typeahead}
            onToggle={this.onToggle}
            onFilter={(event) => {
              if (event) {
                this.onFilter(event.target.value)
              }
            }}
            onClear={this.onClear}
            isOpen={isOpen}
            onSelect={this.onSelection}
            selections={selections}
            isDisabled={isDisabled}
            direction={direction}
            isCreatable={isCreatable}
            createText={createText}
            onCreateOption={() => {}}
          >
            {selectOptions}
          </Select>
        </GridItem>
        {spinner}
      </Grid>
    )
  }
}
AsyncSelect.propTypes = {
  apiOptions: PropTypes.func,
  getFiltersFunc: PropTypes.func,
  optionValueKey: PropTypes.string,
  optionValueFunc: PropTypes.func,
  getOptionsDescriptionFunc: PropTypes.func,
  onSelectionFunc: PropTypes.func,
  isDisabled: PropTypes.bool,
  isMulti: PropTypes.bool,
  selected: PropTypes.oneOfType([PropTypes.string, PropTypes.array]),
  showSpinner: PropTypes.bool,
  direction: PropTypes.string,
  isCreatable: PropTypes.bool,
  createText: PropTypes.string,
  limit: PropTypes.number,
}

export default AsyncSelect
