import {
  Bullseye,
  Button,
  Grid,
  GridItem,
  Split,
  SplitItem,
  Text,
  TextInput,
  TextVariants,
} from "@patternfly/react-core"
import { AddCircleOIcon, MinusCircleIcon } from "@patternfly/react-icons"
import _ from "lodash"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"
import "./Form.scss"

class KeyValueMapForm extends React.Component {
  state = {
    items: [],
  }

  updateItems = () => {
    const { keyValueMap, forceSync = false } = this.props

    if (!keyValueMap) {
      return
    }

    if (forceSync) {
      const keys = Object.keys(keyValueMap) // reload all the keys
      const newItems = keys.map((key) => {
        return { key, value: keyValueMap[key] }
      })
      this.setState({ items: newItems })
      return
    }

    this.setState((prevState) => {
      const { items } = prevState
      let keys = Object.keys(keyValueMap)

      let isEqual = true

      for (let index = 0; index < items.length; index++) {
        const item = items[index]
        if (keys.includes(item.key)) {
          items[index].value = keyValueMap[item.key]
          const keyIndex = keys.indexOf(item.key)
          keys.splice(keyIndex, 1)
        }
      }

      if (keys.length !== 0) {
        isEqual = false
      }

      if (!isEqual) {
        keys = Object.keys(keyValueMap) // reload all the keys
        const newItems = keys.map((key) => {
          return { key, value: keyValueMap[key] }
        })
        return { items: newItems }
      } else {
        return { items: items }
      }
    })
  }

  componentDidMount() {
    this.updateItems()
  }

  componentDidUpdate(prevProps) {
    if (!_.isEqual(this.props.keyValueMap, prevProps.keyValueMap)) {
      this.updateItems()
    }
  }

  callParentOnChange = (items) => {
    const { onChange } = this.props
    if (onChange) {
      // convert to key values
      const keyValueMap = {}
      const keys = []
      for (let index = 0; index < items.length; index++) {
        const item = items[index]
        const key = item.key
        if (!keys.includes(key)) {
          keyValueMap[item.key] = item.value
        }
        keys.push(key)
      }
      onChange(keyValueMap)
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
      items.push({ key: "", value: "" })
      return { items: items }
    })
  }

  render() {
    const { items } = this.state
    const {
      validateKeyFunc,
      validateValueFunc,
      keyLabel,
      valueLabel,
      actionSpan = 1,
      isActionDisabled = false,
      showUpdateButton = false,
      updateButtonCallback = () => {},
      keyField = getKeyField,
      valueField = getValueField,
      isKeyDisabled = false,
      isValueDisabled = false,
      t,
    } = this.props
    const keys = []

    const formItems = items.map((item, index) => {
      let isValidKey = true
      if (validateKeyFunc) {
        isValidKey = validateKeyFunc(item.key)
      }

      let isValidValue = true
      if (validateValueFunc) {
        isValidValue = validateValueFunc(item.value)
      }
      const validatedValue = isValidValue ? "default" : "error"

      // verify duplicate key
      const isDuplicateKey = keys.includes(item.key)
      if (isDuplicateKey) {
        isValidKey = false
      }
      const validatedKey = isValidKey ? "default" : "error"

      // add the key into local map for next iteration
      keys.push(item.key)

      const addButton =
        items.length === index + 1 ? (
          <AddCircleOIcon key={"add-btn" + index} onClick={this.onAdd} className="btn-add icon-btn" />
        ) : null

      const updateButton = showUpdateButton
        ? updateButtonCallback(index, item, (newValue) => this.onChange(index, "value", newValue))
        : null
      return (
        <>
          <GridItem span={4}>
            {keyField(
              index,
              item.key,
              (newKey) => this.onChange(index, "key", newKey),
              validatedKey,
              isKeyDisabled
            )}
          </GridItem>

          <GridItem span={8 - actionSpan}>
            <Split>
              <SplitItem isFilled>
                {valueField(
                  index,
                  item.value,
                  (newValue) => this.onChange(index, "value", newValue),
                  validatedValue,
                  isValueDisabled
                )}
              </SplitItem>
              <SplitItem>{updateButton}</SplitItem>
            </Split>
          </GridItem>
          <GridItem span={actionSpan}>
            {isActionDisabled ? null : (
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
            )}
          </GridItem>
        </>
      )
    })

    if (!items || items.length === 0) {
      formItems.push(
        <Button key="btn-add-an-item" variant="secondary" onClick={this.onAdd} isDisabled={isActionDisabled}>
          {t("add_an_item")}
        </Button>
      )
    }

    return (
      <Grid className="mc-key-value-map-items">
        <GridItem span={4}>
          <Text className="field-title" component={TextVariants.h4}>
            {t(keyLabel ? keyLabel : "key")}
          </Text>
        </GridItem>

        <GridItem span={8 - actionSpan}>
          <Text className="field-title" component={TextVariants.h4}>
            {t(valueLabel ? valueLabel : "value")}
          </Text>
        </GridItem>
        <GridItem span={actionSpan}></GridItem>
        {formItems}
      </Grid>
    )
  }
}

KeyValueMapForm.propTypes = {
  keyLabel: PropTypes.string,
  valueLabel: PropTypes.string,
  isKeyDisabled: PropTypes.bool,
  isValueDisabled: PropTypes.bool,
  actionSpan: PropTypes.number,
  isActionDisabled: PropTypes.bool,
  keyValueMap: PropTypes.object,
  validateKeyFunc: PropTypes.func,
  validateValueFunc: PropTypes.func,
  showUpdateButton: PropTypes.bool,
  updateButtonCallback: PropTypes.func,
  keyField: PropTypes.func,
  valueField: PropTypes.func,
  forceSync: PropTypes.bool,
}

export default withTranslation()(KeyValueMapForm)

// helper functions

const getKeyField = (index, key, onChange, validated, isDisabled = false) => {
  return (
    <TextInput
      id={"key_id_" + index}
      key={"key_" + index}
      value={key}
      validated={validated}
      isDisabled={isDisabled}
      onChange={onChange}
    />
  )
}

const getValueField = (index, value, onChange, validated, isDisabled = false) => {
  return (
    <TextInput
      id={"value_id_" + index}
      key={"value_" + index}
      value={value}
      validated={validated}
      isDisabled={isDisabled}
      onChange={onChange}
    />
  )
}
