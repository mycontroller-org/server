import { getBase } from "../listPageBase"

const slice = getBase("resourceField")

export default slice.reducer

export const {
  updateRecords,
  loading,
  loadingFailed,
  updateFilter,
  deleteFilterValue,
  deleteFilterCategory,
  deleteAllFilter,
  onSortBy,
  updateMetricConfigGauge,
  updateMetricConfigBinary,
} = slice.actions
