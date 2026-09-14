import { Split, SplitItem, Stack } from "@patternfly/react-core"
import React from "react"

// itemsPerRow == 0, all items in a single row
const ArrangeItems = ({ items = [], itemsPerRow = 0 }) => {
  if (itemsPerRow === 0 || items.length <= itemsPerRow) {
    return items
  }

  const rows = []
  const rowsCount = Math.round(items.length / itemsPerRow)
  for (let row = 0; row < rowsCount; row++) {
    const start = itemsPerRow * row
    let end = start + itemsPerRow
    if (items.length < end) {
      end = items.length
    }
    rows.push(getRow(items.slice(start, end)))
  }

  // add remaining items
  const remainingCount = items.length - rowsCount * itemsPerRow
  if (remainingCount > 0) {
    rows.push(getRow(items.slice(items.length - remainingCount)))
  }

  return <Stack>{rows}</Stack>
}

export default ArrangeItems

// helper functions
const getRow = (items = []) => {
  const elements = []
  items.forEach((item) => {
    elements.push(<SplitItem>{item}</SplitItem>)
  })
  return <Split>{elements}</Split>
}
