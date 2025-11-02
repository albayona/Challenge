import type {
  Stock,
  StocksResponse,
  StocksListResponse,
  IndicatorWeight,
  FilteredStocksResponse,
  UniqueValuesResponse,
  ClusterStatsResponse,
} from '@/types/stock'

// Base API URL - configure this in your .env file or update here
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8887'

class ApiService {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  // Build query string from params
  private buildQuery(params: Record<string, string | number | undefined | null>): string {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        search.set(key, String(value))
      }
    })
    const qs = search.toString()
    return qs ? `?${qs}` : ''
  }

  // Build query string with JSON array encoding for weights
  private buildFilterQuery(params: {
    grouping_column?: string
    grouping_value?: string
    sort_by?: string
    order?: 'asc' | 'desc'
    page?: number
    per_page?: number
    numerical_weights?: IndicatorWeight[]
    rating_weights?: IndicatorWeight[]
  }): string {
    const search = new URLSearchParams()

    if (params.grouping_column) {
      search.set('grouping_column', params.grouping_column)
    }

    if (params.grouping_value) {
      search.set('grouping_value', params.grouping_value)
    }

    if (params.sort_by) {
      search.set('sort_by', params.sort_by)
    }

    if (params.order) {
      search.set('order', params.order)
    }

    if (params.page !== undefined) {
      search.set('page', String(params.page))
    }

    if (params.per_page !== undefined) {
      search.set('per_page', String(params.per_page))
    }

    if (params.numerical_weights && params.numerical_weights.length > 0) {
      search.set('numerical_weights', JSON.stringify(params.numerical_weights))
    }

    if (params.rating_weights && params.rating_weights.length > 0) {
      search.set('rating_weights', JSON.stringify(params.rating_weights))
    }

    const qs = search.toString()
    return qs ? `?${qs}` : ''
  }

  // Paginated stocks with optional filters and sorting
  async getStocksPaginated(options: {
    page: number
    perPage: number
    sortBy?: string
    order?: 'asc' | 'desc'
    action?: string
    company?: string
    cluster?: number
    ticker?: string
  }): Promise<StocksResponse> {
    const { page, perPage, sortBy, order, action, company, cluster, ticker } = options
    const query = this.buildQuery({
      page,
      per_page: perPage,
      sort_by: sortBy,
      order,
      action,
      company,
      cluster,
      ticker,
    })
    // Use base endpoint with query params; backend may ignore unknown params gracefully
    return this.request<StocksResponse>(`/api/v1/stocks${query}`)
  }

  private async request<T>(endpoint: string): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    const response = await fetch(url)

    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`)
    }

    return response.json()
  }

  // Get all stocks
  async getAllStocks(): Promise<StocksResponse> {
    return this.request<StocksResponse>('/api/v1/stocks')
  }

  // Get stocks by action
  async getStocksByAction(action: string): Promise<StocksResponse> {
    const encodedAction = encodeURIComponent(action)
    return this.request<StocksResponse>(`/api/v1/stocks/action/${encodedAction}`)
  }

  // Get unique actions
  async getActions(): Promise<StocksListResponse> {
    return this.request<StocksListResponse>('/api/v1/stocks/actions')
  }

  // Get stocks by cluster
  async getStocksByCluster(cluster: number): Promise<StocksResponse> {
    return this.request<StocksResponse>(`/api/v1/stocks/cluster/${cluster}`)
  }

  // Get unique clusters
  async getClusters(): Promise<StocksListResponse> {
    return this.request<StocksListResponse>('/api/v1/stocks/clusters')
  }

  // Get unique companies
  async getCompanies(): Promise<StocksListResponse> {
    return this.request<StocksListResponse>('/api/v1/stocks/companies')
  }

  // Get stocks by company
  async getStocksByCompany(company: string): Promise<StocksResponse> {
    const encodedCompany = encodeURIComponent(company)
    return this.request<StocksResponse>(`/api/v1/stocks/company/${encodedCompany}`)
  }

  // Get stock by ticker
  async getStockByTicker(ticker: string): Promise<Stock> {
    return this.request<Stock>(`/api/v1/stocks/ticker/${ticker}`)
  }

  // Get stock by ID
  async getStockById(id: number): Promise<Stock> {
    return this.request<Stock>(`/api/v1/stocks/${id}`)
  }

  // Get filtered stocks by cluster with grouping, sorting, pagination, and weights
  async getFilteredStocksByCluster(options: {
    cluster: number
    grouping_column?: string
    grouping_value?: string
    sort_by?: string
    order?: 'asc' | 'desc'
    page?: number
    per_page?: number
    numerical_weights?: IndicatorWeight[]
    rating_weights?: IndicatorWeight[]
  }): Promise<FilteredStocksResponse> {
    const { cluster, ...queryParams } = options
    const query = this.buildFilterQuery(queryParams)
    return this.request<FilteredStocksResponse>(`/api/v1/stocks/cluster/${cluster}/filter${query}`)
  }

  // Get unique values for a grouping column within a cluster
  async getUniqueValues(cluster: number, groupingColumn: string): Promise<UniqueValuesResponse> {
    const encodedColumn = encodeURIComponent(groupingColumn)
    return this.request<UniqueValuesResponse>(
      `/api/v1/stocks/cluster/${cluster}/unique/${encodedColumn}`,
    )
  }

  // Get cluster statistics
  async getClusterStats(): Promise<ClusterStatsResponse> {
    return this.request<ClusterStatsResponse>('/api/v1/clusters/stats')
  }
}

export const apiService = new ApiService(API_BASE_URL)
