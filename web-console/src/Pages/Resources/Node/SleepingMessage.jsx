import { TableComposable, Tbody, Td, Th, Thead, Tr } from "@patternfly/react-table"
import React from "react"
import { LastSeen } from "../../../Components/Time/Time"
import { getValue } from "../../../Util/Util"
import { useTranslation } from "react-i18next"

const SleepingMessage = ({ messages = [] }) => {
  const { t } = useTranslation()
  const columns = ["s_no", "node_id", "source_id", "type", "is_sleep_node", "timestamp", "payloads"]
  const rows = messages.map((msg, index) => {
    // reformat payload
    const payloads = getValue(msg, "payloads", []).map((pl) => {
      let plValue = pl.key
      if (pl.value !== "") {
        plValue = `${plValue}=${pl.value}`
      }
      const labels = getValue(pl, "labels", {})
      const others = getValue(pl, "others", {})
      if (Object.keys(labels).length > 0) {
        plValue = `${plValue}, labels:'${JSON.stringify(labels)}'`
      }
      if (Object.keys(others).length > 0) {
        plValue = `${plValue}, others:'${JSON.stringify(others)}'`
      }
      return plValue
    })

    return {
      cells: [
        index + 1,
        getValue(msg, "nodeId", ""),
        getValue(msg, "sourceId", ""),
        getValue(msg, "type", ""),
        getValue(msg, "isSleepNode", "") ? t("true") : t("false"),
        <LastSeen date={getValue(msg, "timestamp", "")} tooltipPosition="left" />,
        payloads.join("<br>"),
      ],
    }
  })

  const endpointsTable = (
    <TableComposable aria-label="agent_table" variant="compact">
      <Thead>
        <Tr>
          {columns.map((column, columnIndex) => (
            <Th key={columnIndex}>{t(column)}</Th>
          ))}
        </Tr>
      </Thead>
      <Tbody>
        {rows.map((row, rowIndex) => (
          <Tr key={rowIndex}>
            {row.cells.map((cell, cellIndex) => (
              <Td key={`${rowIndex}_${cellIndex}`} dataLabel={columns[cellIndex]}>
                {cell}
              </Td>
            ))}
          </Tr>
        ))}
      </Tbody>
    </TableComposable>
  )

  return endpointsTable
}

export default SleepingMessage
