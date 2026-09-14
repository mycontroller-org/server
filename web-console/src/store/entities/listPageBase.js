import { createSlice } from "@reduxjs/toolkit"
import { AggregationInterval, Duration, InterpolationType, MetricFunctionType } from "../../Constants/Metric"

export const getBase = (name) => {
  return createSlice({
    name: name,
    initialState: {
      loading: true,
      records: [],
      revision: 0, // this key used to call fetch records api
      lastUpdate: null,
      count: 0,
      pagination: {
        limit: 10,
        page: 0,
      },
      filters: {},
      sortBy: {
        // index: 0,
        // field: "",
        // direction: "asc",
      },
      metricConfig: {
        gauge: {
          func: MetricFunctionType.Mean,
          interpolationType: InterpolationType.Natural,
          duration: Duration.LastHour,
          interval: AggregationInterval.Minute_1,
        },
        binary: {
          duration: Duration.Last2Hours,
        },
      },
    },
    reducers: {
      updateRecords: (state, action) => {
        const { records, count, pagination } = action.payload
        state.records = records
        state.count = count
        state.loading = false
        state.lastUpdate = Date.now()

        state.pagination.limit = pagination.limit
        state.pagination.page = pagination.page

        let maxPage = count / state.pagination.limit
        if (count % state.pagination.limit > 0) {
          maxPage++
        }
        if (state.pagination.page > maxPage) {
          state.pagination.page = maxPage
        }
      },

      updateFilter: (state, action) => {
        const { category, value } = action.payload
        if (!state.filters[category]) {
          state.filters[category] = []
        }
        if (!state.filters[category].includes(value)) {
          state.filters[category].push(value)
          state.revision++
        }
      },

      deleteFilterValue: (state, action) => {
        const { category, value } = action.payload
        if (state.filters[category]) {
          const values = state.filters[category]
          state.filters[category] = values.filter((v) => v !== value)
          state.revision++
        }
      },

      deleteFilterCategory: (state, action) => {
        const { category } = action.payload
        state.filters[category] = []
        state.revision++
      },

      deleteAllFilter: (state, _action) => {
        state.filters = {}
        state.revision++
      },

      loading: (state, _action) => {
        state.loading = true
      },

      loadingFailed: (state, _action) => {
        state.loading = false
      },

      onSortBy: (state, action) => {
        const { index, field, direction } = action.payload
        state.sortBy = { index, field, direction }
        state.revision++
      },

      updateMetricConfigGauge: (state, action) => {
        const { func, interpolationType, duration, interval } = action.payload
        state.metricConfig.gauge = { func, interpolationType, duration, interval }
      },

      updateMetricConfigBinary: (state, action) => {
        const { duration } = action.payload
        state.metricConfig.binary = { duration }
      },
    },
  })
}
