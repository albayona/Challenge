import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { apiService } from '@/services/api'
import type { Stock, IndicatorWeight, UniqueValuesResponse } from '@/types/stock'

export const useStocksStore = defineStore('stocks', () => {
  // State
  const stocks = ref<Stock[]>([])
  const selectedStock = ref<Stock | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const count = ref(0)

  // Filter options
  const actions = ref<string[]>([])
  const clusters = ref<number[]>([])
  const companies = ref<string[]>([])
  const filtersLoading = ref(false)

  // Grouping columns
  const groupingColumns = ['action', 'rating_to', 'rating_from'] as const

  // Unique values
  const uniqueValues = ref<UniqueValuesResponse | null>(null)
  const uniqueValuesLoading = ref(false)
  const uniqueValuesError = ref<string | null>(null)

  // Indicator weights state
  const weights = ref<{
    numerical: Record<string, number>
    rating: Record<string, number>
  }>({
    numerical: {},
    rating: {},
  })

  // Computed
  const hasStocks = computed(() => stocks.value.length > 0)
  const hasSelectedStock = computed(() => selectedStock.value !== null)

  // Actions
  async function fetchAllStocks() {
    loading.value = true
    error.value = null
    try {
      const response = await apiService.getAllStocks()
      stocks.value = response.data
      count.value = response.count
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stocks'
      stocks.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchStocksByAction(action: string) {
    loading.value = true
    error.value = null
    try {
      const response = await apiService.getStocksByAction(action)
      stocks.value = response.data
      count.value = response.count
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stocks by action'
      stocks.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchStocksByCluster(cluster: number) {
    loading.value = true
    error.value = null
    try {
      const response = await apiService.getStocksByCluster(cluster)
      stocks.value = response.data
      count.value = response.count
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stocks by cluster'
      stocks.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchFilteredStocksByCluster(options: {
    cluster: number
    grouping_column?: string
    grouping_value?: string
    sort_by?: string
    order?: 'asc' | 'desc'
    page?: number
    per_page?: number
    numerical_weights?: IndicatorWeight[]
    rating_weights?: IndicatorWeight[]
  }) {
    loading.value = true
    error.value = null
    try {
      const response = await apiService.getFilteredStocksByCluster(options)
      stocks.value = response.data
      count.value = response.total_count
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : 'Failed to fetch filtered stocks by cluster'
      stocks.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchStocksByCompany(company: string) {
    loading.value = true
    error.value = null
    try {
      const response = await apiService.getStocksByCompany(company)
      stocks.value = response.data
      count.value = response.count
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stocks by company'
      stocks.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  async function fetchStockById(id: number) {
    loading.value = true
    error.value = null
    try {
      selectedStock.value = await apiService.getStockById(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stock'
      selectedStock.value = null
    } finally {
      loading.value = false
    }
  }

  async function fetchStockByTicker(ticker: string) {
    loading.value = true
    error.value = null
    try {
      selectedStock.value = await apiService.getStockByTicker(ticker)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch stock'
      selectedStock.value = null
    } finally {
      loading.value = false
    }
  }

  // Fetch filter options
  async function fetchActions() {
    if (actions.value.length > 0) return // Already loaded
    filtersLoading.value = true
    try {
      const response = await apiService.getActions()
      actions.value = (response.data as string[]) || []
    } catch (err) {
      console.error('Failed to fetch actions:', err)
    } finally {
      filtersLoading.value = false
    }
  }

  async function fetchClusters() {
    if (clusters.value.length > 0) return // Already loaded
    filtersLoading.value = true
    try {
      const response = await apiService.getClusters()
      clusters.value = (response.data as number[]) || []
    } catch (err) {
      console.error('Failed to fetch clusters:', err)
    } finally {
      filtersLoading.value = false
    }
  }

  async function fetchCompanies() {
    if (companies.value.length > 0) return // Already loaded
    filtersLoading.value = true
    try {
      const response = await apiService.getCompanies()
      companies.value = (response.data as string[]) || []
    } catch (err) {
      console.error('Failed to fetch companies:', err)
    } finally {
      filtersLoading.value = false
    }
  }

  function setSelectedStock(stock: Stock | null) {
    selectedStock.value = stock
  }

  function clearFilters() {
    stocks.value = []
    count.value = 0
    error.value = null
  }

  // Update indicator weight
  function updateWeight(type: 'numerical' | 'rating', indicatorName: string, weight: number) {
    weights.value[type][indicatorName] = weight
  }

  // Initialize weights from default values
  function initializeWeights(
    numericalIndicators: Record<string, number>,
    ratingIndicators: Record<string, number>,
  ) {
    Object.keys(numericalIndicators).forEach((indicatorName) => {
      if (weights.value.numerical[indicatorName] === undefined) {
        weights.value.numerical[indicatorName] = numericalIndicators[indicatorName] ?? 10
      }
    })
    Object.keys(ratingIndicators).forEach((indicatorName) => {
      if (weights.value.rating[indicatorName] === undefined) {
        weights.value.rating[indicatorName] = ratingIndicators[indicatorName] ?? 10
      }
    })
  }

  // Fetch unique values for a grouping column within a cluster
  async function fetchUniqueValues(cluster: number, groupingColumn: string) {
    uniqueValuesLoading.value = true
    uniqueValuesError.value = null
    try {
      uniqueValues.value = await apiService.getUniqueValues(cluster, groupingColumn)
    } catch (err) {
      uniqueValuesError.value = err instanceof Error ? err.message : 'Failed to fetch unique values'
      uniqueValues.value = null
    } finally {
      uniqueValuesLoading.value = false
    }
  }

  return {
    // State
    stocks,
    selectedStock,
    loading,
    error,
    count,
    actions,
    clusters,
    companies,
    filtersLoading,
    groupingColumns,
    uniqueValues,
    uniqueValuesLoading,
    uniqueValuesError,
    weights,
    // Computed
    hasStocks,
    hasSelectedStock,
    // Actions
    fetchAllStocks,
    fetchStocksByAction,
    fetchStocksByCluster,
    fetchFilteredStocksByCluster,
    fetchStocksByCompany,
    fetchStockById,
    fetchStockByTicker,
    fetchActions,
    fetchClusters,
    fetchCompanies,
    fetchUniqueValues,
    setSelectedStock,
    clearFilters,
    updateWeight,
    initializeWeights,
  }
})
