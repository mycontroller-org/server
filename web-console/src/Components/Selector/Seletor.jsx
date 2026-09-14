import { Divider, Dropdown, DropdownItem, DropdownToggle, Split, SplitItem } from "@patternfly/react-core"
import { CaretDownIcon, MinusCircleIcon, PlusCircleIcon } from "@patternfly/react-icons"
import PropTypes from "prop-types"
import React from "react"
import "./Selector.scss"

// Sample data
// {
//   id: "abc",
//   text: "xyz",
//   description: "xyz details",
//   favorite: true,
//   disabled: false,
// }
class Selector extends React.Component {
  state = {
    isOpen: false,
  }

  onToggle = (isOpen) => {
    this.setState({ isOpen })
  }
  onFocus = () => {
    const element = document.getElementById("selector-toggle-id")
    element.focus()
  }

  onSelect = (item) => {
    // do local update
    this.setState({
      isOpen: !this.state.isOpen,
      selected: item,
    })
    // notify to parent
    if (this.props.onChange) {
      this.props.onChange(item)
    }
  }

  componentDidMount() {
    // fetch from the given api
  }

  render() {
    const { items, selection, disabled, prefix, onBookmarkAction } = this.props

    const elements = []
    // update favorite and other items
    const itemsFavorite = []
    const itemsOther = []
    if (items) {
      items.forEach((item) => {
        if (item.favorite) {
          itemsFavorite.push(getItem(item, this.onSelect, onBookmarkAction))
        } else {
          itemsOther.push(getItem(item, this.onSelect, onBookmarkAction))
        }
      })
    }

    // update all
    if (itemsFavorite.length > 0) {
      elements.push(...itemsFavorite)
    }
    if (itemsFavorite.length > 0 && itemsOther.length > 0) {
      elements.push(<Divider key="divider" component="li" />)
    }
    elements.push(...itemsOther)

    const title = prefix ? <span className="rb-selector-title">{prefix}:&nbsp;</span> : null

    const className = disabled ? "mc-selector disabled" : "mc-selector"
    return (
      <Dropdown
        className={className}
        isPlain
        disabled={disabled}
        toggle={
          <DropdownToggle
            id="selector-toggle-id"
            isDisabled={disabled}
            onToggle={this.onToggle}
            toggleIndicator={CaretDownIcon}
          >
            <span className="mc-mobile-view">{title}</span>
            <span>{selection.text}</span>
          </DropdownToggle>
        }
        isOpen={this.state.isOpen}
        dropdownItems={elements}
      />
    )
  }
}

Selector.propTypes = {
  items: PropTypes.array,
  selected: PropTypes.object,
  onChange: PropTypes.func,
  onBookmarkAction: PropTypes.func,
  isDisabled: PropTypes.bool,
}

export default Selector

// helper functions

const getItem = (item, onSelectFn, onBookmarkChangeFn) => {
  const Icon = item.favorite ? MinusCircleIcon : PlusCircleIcon
  return (
    <DropdownItem key={item.id} className="mc-selector-dropdown" isDisabled={item.disabled}>
      <Split style={{ width: "100%" }}>
        <SplitItem>
          <div className="mc-selector-favorite" onClick={() => onBookmarkChangeFn(item)}>
            <Icon />
          </div>
        </SplitItem>
        <SplitItem isFilled>
          <div className="mc-selector-item" onClick={() => onSelectFn(item)}>
            {item.text}
          </div>
        </SplitItem>
      </Split>
    </DropdownItem>
  )
}
