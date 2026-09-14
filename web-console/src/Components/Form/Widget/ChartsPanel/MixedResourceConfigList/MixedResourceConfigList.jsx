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
import "../../../Form.scss"
import MixedResourceConfigPicker from "./MixedResourceConfigPicker"

class MixedResourceConfigList extends React.Component {
  state = {
    items: [],
  }

  updateItems = () => {
    const { valuesList } = this.props
    if (!valuesList) {
      return
    }

    let valuesReceived = [...valuesList]

    this.setState((prevState) => {
      const { items } = prevState

      for (let index = 0; index < items.length; index++) {
        const item = items[index]
        if (valuesReceived.includes(item)) {
          const valueIndex = valuesReceived.indexOf(item)
          valuesReceived.splice(valueIndex, 1)
        }
      }

      if (valuesReceived.length !== 0) {
        return { items: valuesList }
      } else {
        return { items: items }
      }
    })
  }

  componentDidMount() {
    this.updateItems()
  }

  componentDidUpdate(prevProps) {
    if (!_.isEqual(this.props.valuesList, prevProps.valuesList)) {
      this.updateItems()
    }
  }

  callParentOnChange = (items) => {
    const { onChange } = this.props
    if (onChange) {
      onChange(items)
    }
  }

  onChange = (index, newValue) => {
    this.setState((prevState) => {
      const { items } = prevState
      items[index] = newValue
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
      items.push("")
      return { items: items }
    })
  }

  render() {
    const { items } = this.state
    const { validateValueFunc, valueLabel, yAxisConfig, t } = this.props
    const values = []

    const formItems = items.map((item, index) => {
      let isValidValue = true
      if (validateValueFunc) {
        isValidValue = validateValueFunc(item)
      }

      // verify duplicate value
      if (values.includes(item)) {
        isValidValue = false
      }

      const validatedValue = isValidValue ? "default" : "error"

      // add the item into local map for next iteration
      values.push(item)

      const addButton =
        items.length === index + 1 ? (
          <AddCircleOIcon key={"add-btn" + index} onClick={this.onAdd} className="btn-add icon-btn" />
        ) : null

      return (
        <>
          <GridItem span={11}>
            <Split>
              <SplitItem isFilled>
                <TextInput
                  id={"value_" + index}
                  key={"value_" + index}
                  value={getItemDetails(item, t)}
                  validated={validatedValue}
                  onChange={() => {}}
                  isDisabled={true}
                />
              </SplitItem>
              <SplitItem>
                <MixedResourceConfigPicker
                  key={"picker_" + index}
                  value={item}
                  yAxisConfig={yAxisConfig}
                  id={"mixed_ctrl_model_" + index}
                  onChange={(newValue) => {
                    this.onChange(index, newValue)
                  }}
                />
              </SplitItem>
            </Split>
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
        <GridItem span={11}>
          <Text className="field-title" component={TextVariants.h4}>
            {valueLabel ? valueLabel : ""}
          </Text>
        </GridItem>
        <GridItem span={1}></GridItem>
        {formItems}
      </Grid>
    )
  }
}

MixedResourceConfigList.propTypes = {
  valueLabel: PropTypes.string,
  valuesList: PropTypes.array,
  validateValueFunc: PropTypes.func,
  yAxisConfig: PropTypes.object,
}

export default withTranslation()(MixedResourceConfigList)

// helper functions
const getItemDetails = (item, t) => {
  if (item) {
    const { resourceType: rt, quickId: qid, chartType } = item
    return `resource=${rt}:${qid}, chartType=${chartType}`
  }
  return t("do_update")
}
