import PropTypes from "prop-types"
import React from "react"
import Table from "../Table/Table"

export default class DetailTableBase extends React.Component {
  state = {
    loading: true,
    data: {},
  }

  loadDetail() {
    const { id } = this.props.match.params
    if (id) {
      this.props
        .apiGetRecord(id)
        .then((res) => {
          this.setState({ data: res.data, loading: false })
        })
        .catch((_e) => {
          this.setState({ loading: false })
        })
    }
  }

  render() {
    const { loading, data } = this.state

    if (loading) {
      return <div>Loading</div>
    }

    const filters = this.props.getTableFilterFunc(data)
    return (
      <Table
        apiGetRecords={this.props.apiListTablesRecord}
        tableColumns={this.props.tableColumns}
        toRowFunc={this.props.tableRowsFunc}
        history={this.props.history}
        filters={filters}
      />
    )
  }
}

DetailTableBase.propTypes = {
  apiGetRecord: PropTypes.func,
  apiListTablesRecord: PropTypes.func,
  getTableFilterFunc: PropTypes.func,
  tableRowsFunc: PropTypes.func,
  tableColumns: PropTypes.array,
  match: PropTypes.object,
  history: PropTypes.object,
}
