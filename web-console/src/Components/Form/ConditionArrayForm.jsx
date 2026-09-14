import {
  Bullseye,
  Button,
  Grid,
  GridItem,
  SelectVariant,
  Split,
  Text,
  TextInput,
  TextVariants,
} from "@patternfly/react-core"
import { AddCircleOIcon, MinusCircleIcon } from "@patternfly/react-icons"
import _ from "lodash"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import { Operator, OperatorOptions } from "../../Constants/Filter"
import "./Form.scss"
import Select from "./Select"

class ConditionArrayMapForm extends React.Component {
  state = {
    items: [],
  }

  updateItems = () => {
    const { conditionsArrayMap } = this.props

    if (!conditionsArrayMap) {
      return
    }

    this.setState((prevState) => {
      const { items } = prevState
      let variables = conditionsArrayMap.map((d) => {
        return d.variable
      })

      for (let index = 0; index < items.length; index++) {
        const item = items[index]
        if (variables.includes(item.variable)) {
          const keyIndex = variables.indexOf(item.variable)
          variables.splice(keyIndex, 1)
        }
      }

      if (variables !== 0) {
        return { items: conditionsArrayMap }
      } else {
        return { items: items }
      }
    })
  }

  componentDidMount() {
    this.updateItems()
  }

  componentDidUpdate(prevProps) {
    if (!_.isEqual(this.props.variableValueMap, prevProps.variableValueMap)) {
      this.updateItems()
    }
  }

  callParentOnChange = (items) => {
    const { onChange } = this.props
    if (onChange) {
      for (let index = 0; index < items.length; index++) {
        const item = items[index]
        switch (item.operator) {
          case Operator.In:
          case Operator.NotIn:
          case Operator.RangeIn:
          case Operator.RangeNotIn:
            if (!Array.isArray(item.value)) {
              items[index].value = item.value.split(",")
            }
            break

          default:
        }
      }
      onChange(items)
    }
  }

  onChange = (index, type, newValue) => {
    this.setState((prevState) => {
      const { items } = prevState
      items[index][type] = newValue
      this.callParentOnChange(items)
      return { items: items }
    })
  }

  onDelete = (index) => {
    this.setState((prevState) => {
      const { items } = prevState
      items.splice(index, 1)
      this.callParentOnChange(items)
      return { items: items }
    })
  }

  onAdd = () => {
    this.setState((prevState) => {
      const { items } = prevState
      items.push({ variable: "", operator: "", value: "" })
      return { items: items }
    })
  }

  render() {
    const { items } = this.state
    const {
      validateKeyFunc,
      validateValueFunc,
      variableLabel,
      operatorLabel,
      valueLabel,
      direction = "",
      t,
    } = this.props
    const variables = []

    const formItems = items.map((item, index) => {
      let isValidKey = true
      if (validateKeyFunc) {
        isValidKey = validateKeyFunc(item.variable)
      }

      let isValidValue = true
      if (validateValueFunc) {
        isValidValue = validateValueFunc(item.variable)
      }
      const validatedValue = isValidValue ? "default" : "error"
      const validatedKey = isValidKey ? "default" : "error"

      // add the key into local map for next iteration
      variables.push(item.variable)

      const addButton =
        items.length === index + 1 ? (
          <AddCircleOIcon key={"add-btn" + index} onClick={this.onAdd} className="btn-add icon-btn" />
        ) : null
      return (
        <>
          <GridItem span={4}>
            <TextInput
              id={"variable_" + index}
              value={item.variable}
              validated={validatedKey}
              onChange={(newValue) => {
                this.onChange(index, "variable", newValue)
              }}
            />
          </GridItem>

          <GridItem span={2}>
            <Select
              key={"operator_" + index}
              disableClear={true}
              hideDescription={true}
              variant={SelectVariant.single}
              options={OperatorOptions}
              selected={item.operator}
              direction={direction}
              onChange={(newValue) => {
                this.onChange(index, "operator", newValue)
              }}
            />
          </GridItem>

          <GridItem span={5}>
            <TextInput
              id={"value_" + index}
              key={"value_" + index}
              value={item.value}
              validated={validatedValue}
              onChange={(newValue) => {
                this.onChange(index, "value", newValue)
              }}
            />
          </GridItem>
          <GridItem span={1}>
            <Bullseye className="btn-layout">
              <Split hasGutter className="btn-split">
                <MinusCircleIcon
                  key={"btn-remove-" + index}
                  onClick={() => this.onDelete(index)}
                  className="btn-remove icon-btn"
                />
                {addButton}
              </Split>
            </Bullseye>
          </GridItem>
        </>
      )
    })

    if (!items || items.length === 0) {
      formItems.push(
        <Button key="btn-add-an-item" variant="secondary" onClick={this.onAdd}>
          {t("add_an_item")}
        </Button>
      )
    }

    return (
      <Grid className="mc-key-value-map-items">
        <GridItem span={4}>
          <Text className="field-title" component={TextVariants.h4}>
            {t(variableLabel ? variableLabel : "variable")}
          </Text>
        </GridItem>
        <GridItem span={2}>
          <Text className="field-title" component={TextVariants.h4}>
            {t(operatorLabel ? operatorLabel : "operator")}
          </Text>
        </GridItem>

        <GridItem span={5}>
          <Text className="field-title" component={TextVariants.h4}>
            {t(valueLabel ? valueLabel : "value")}
          </Text>
        </GridItem>
        <GridItem span={1}></GridItem>

        {formItems}
      </Grid>
    )
  }
}

ConditionArrayMapForm.propTypes = {
  variableLabel: PropTypes.string,
  operatorLabel: PropTypes.string,
  valueLabel: PropTypes.string,
  conditionsArrayMap: PropTypes.array,
  validateKeyFunc: PropTypes.func,
  validateValueFunc: PropTypes.func,
  direction: PropTypes.string,
}

export default withTranslation()(ConditionArrayMapForm)
