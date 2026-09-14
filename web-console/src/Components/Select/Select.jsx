import { Select as PfSelect, SelectOption, SelectVariant } from "@patternfly/react-core"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"

class Select extends React.Component {
  state = {
    isOpen: false,
  }

  onChange = (selection) => {
    if (this.props.onSelectionFunc) {
      this.props.onSelectionFunc(selection)
    }
  }

  componentDidMount() {
    if (this.props.defaultValue) {
      this.onChange(this.props.defaultValue)
    }
  }

  onSelect = (_event, selection) => {
    this.setState({ isOpen: false }, this.onChange(selection))
  }

  onToggle = (isOpen) => {
    this.setState({ isOpen })
  }

  render() {
    const { title, disabled, options, direction = "", defaultValue, t } = this.props
    return (
      <div>
        <span id="select-title">{title}</span>
        <PfSelect
          variant={SelectVariant.single}
          placeholderText={t("select_an_option")}
          //aria-label="Select Input with descriptions"
          onToggle={this.onToggle}
          onSelect={this.onSelect}
          selections={defaultValue}
          isOpen={this.state.isOpen}
          //aria-labelledby={titleId}
          isDisabled={disabled}
          direction={direction}
        >
          {options.map((option, index) => (
            <SelectOption isDisabled={option.disabled} key={index} value={option.value}>
              {t(option.label)}
            </SelectOption>
          ))}
        </PfSelect>
      </div>
    )
  }
}

Select.propTypes = {
  title: PropTypes.string,
  options: PropTypes.array, // [{value: "abc", display:"ABC text", disabled: false}]
  defaultValue: PropTypes.string,
  onSelectionFunc: PropTypes.func,
  disabled: PropTypes.bool,
}

export default withTranslation()(Select)
