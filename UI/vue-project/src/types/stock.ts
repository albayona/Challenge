export interface RatingSentiment {
  id: number
  stock_data_point_id: number
  name: string
  rating: string
  rating_score: number
  norm_rating_score: number
  created_at: string
  updated_at: string
}

export interface NumericalIndicator {
  id: number
  stock_data_point_id: number
  name: string
  value: number
  norm_value: number
  created_at: string
  updated_at: string
}

export interface Stock {
  id: number
  ticker: string
  action: string
  date: string
  company: string
  cluster: number
  final_score: number
  target_to?: number
  target_from?: number
  target_delta?: number
  rating_to?: string
  rating_from?: string
  weighted_score?: number
  created_at: string
  updated_at: string
  rating_sentiments: RatingSentiment[]
  numerical_indicators: NumericalIndicator[]
}

export interface StocksResponse {
  count: number
  data: Stock[]
}

export interface FilteredStocksResponse {
  data: Stock[]
  grouping_column?: string
  grouping_value?: string
  order?: string
  page?: number
  per_page?: number
  sort_by?: string
  total_count: number
}

export interface StocksListResponse {
  data: string[] | number[]
}

export interface IndicatorWeight {
  indicator_name: string
  weight: number
}

export interface UniqueValuesResponse {
  cluster: number
  column_name: string
  count: number
  values: string[]
}

export interface SilhouetteStats {
  mean: number
  min: number
  max: number
  std: number
}

export interface DistributionBin {
  range: [number, number]
  percentage: number
}

export interface FeatureStats {
  count: number
  mean: number
  std: number
  min: number
  '25%': number
  '50%': number
  '75%': number
  max: number
  f_value: number
  p_value: number
  distribution?: DistributionBin[]
}

export interface ClusterStats {
  cluster: number
  inertia: number
  silhouette: SilhouetteStats
  features: Record<string, FeatureStats>
}

export type ClusterStatsResponse = ClusterStats[]
