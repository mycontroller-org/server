import { Pagination, PaginationVariant, Title } from "@patternfly/react-core"
import { sortable, Table as PfTable, TableBody, TableHeader, TableVariant } from "@patternfly/react-table"
import PropTypes from "prop-types"
import React from "react"
import { withTranslation } from "react-i18next"

const ITEMS_LIMIT = 10

class Table extends React.Component {
  state = {
    loading: true,
    limit: ITEMS_LIMIT,
    sortBy: { index: -1 },
    page: 0,
    rows: [],
    records: {},
  }

  fetchRecords = (pagination = {}) => {
    // update filters
    const filters = Object.keys(this.props.filters).map((key) => {
      const value = this.props.filters[key]
      return { k: key, o: "eq", v: value }
    })

    const limit = pagination.limit ? pagination.limit : ITEMS_LIMIT
    const pageNo = pagination.page ? pagination.page : 0
    const _page = {
      limit: limit,
      offset: pageNo * limit,
      filter: filters,
    }
    // update sortBy
    let sortBy = {}
    if (pagination.sortBy) {
      sortBy = { ...this.state.sortBy, ...pagination.sortBy }
    } else if (this.state.sortBy && this.state.sortBy.field) {
      sortBy = this.state.sortBy
    }
    if (sortBy) {
      _page.sortBy = [{ f: sortBy.field, o: sortBy.direction }]
    }

    this.props
      .apiGetRecords(_page)
      .then((res) => {
        const rows = res.data.data.map((d, index) => {
          return this.props.toRowFunc(d, index, this.props.history)
        })
        this.setState({
          loading: false,
          rows: rows,
          records: res.data,
          limit: _page.limit,
          page: pageNo,
          sortBy: sortBy,
        })
      })
      .catch((_e) => {
        this.setState({ loading: false })
      })
  }

  componentDidMount() {
    this.fetchRecords()
  }

  onPerPage = (_e, perPage) => {
    this.fetchRecords({ limit: perPage, page: 0 })
  }

  onPageSet = (_e, page) => {
    const pagination = { limit: this.state.limit, page: this.state.page }
    pagination.page = page - 1
    this.fetchRecords(pagination)
  }

  onSortBy = (_event, index, direction) => {
    // console.log(index, direction, this.props.tableColumns[index])
    const header = this.props.tableColumns[index]
    if (header) {
      this.fetchRecords({ sortBy: { field: header.fieldKey, direction: direction, index: index } })
    }
  }

  render() {
    const { t, tableColumns, title } = this.props
    const { loading, rows, sortBy, records, limit, page } = this.state

    // update table headers
    const tableColumnsUpdated = tableColumns.map((tc) => {
      const header = { title: t(tc.title) }
      if (tc.sortable) {
        header.transforms = [sortable]
      }
      return header
    })

    const titleUpdated = title ? <Title headingLevel="h6">{t(title)}</Title> : null

    const loadingElement = <div style={{ marginBottom: "15px" }}>{t("loading")}...</div>

    let table = null
    if (!loading) {
      table =
        rows.length > 0 ? (
          <>
            <PfTable
              aria-label="Compact Table"
              variant={TableVariant.compact}
              cells={tableColumnsUpdated}
              rows={rows}
              rowLabeledBy="rowIndex"
              canSelectAll={false}
              className="mc-table"
              borders={true}
              sortBy={sortBy}
              onSort={this.onSortBy}
            >
              <TableHeader />
              <TableBody rowKey="rid" />
            </PfTable>
            <Pagination
              itemCount={records.count}
              widgetId="pagination-options-menu-bottom"
              perPage={limit}
              page={page + 1}
              variant={PaginationVariant.bottom}
              onSetPage={this.onPageSet}
              onPerPageSelect={this.onPerPage}
            />
          </>
        ) : (
          <div style={{ marginBottom: "15px" }}>{t("no_data")}</div>
        )
    }

    return (
      <>
        {titleUpdated}
        {loading ? loadingElement : table}
      </>
    )
  }
}

Table.propTypes = {
  title: PropTypes.string,
  apiGetRecords: PropTypes.func,
  filters: PropTypes.object,
  tableColumns: PropTypes.array,
  toRowFunc: PropTypes.func,
  history: PropTypes.object,
}

export default withTranslation()(Table)
