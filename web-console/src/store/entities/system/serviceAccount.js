import { getBase } from "../listPageBase"

const slice = getBase("serviceAccount")

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
} = slice.actions
