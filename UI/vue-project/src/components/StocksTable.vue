<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import debounce from 'lodash-es/debounce'
import type { Stock, IndicatorWeight } from '@/types/stock'
import { apiService } from '@/services/api'
import { useStocksStore } from '@/stores/stocks'
import IndicatorChip from './IndicatorChip.vue'
import StockDetail from './StockDetail.vue'

const stocksStore = useStocksStore()

// Convert weights to IndicatorWeight[] format
const numericalWeights = computed<IndicatorWeight[]>(() => {
  return Object.entries(stocksStore.weights.numerical).map(([indicator_name, weight]) => ({
    indicator_name,
    weight,
  }))
})

const ratingWeights = computed<IndicatorWeight[]>(() => {
  return Object.entries(stocksStore.weights.rating).map(([indicator_name, weight]) => ({
    indicator_name,
    weight,
  }))
})

const props = defineProps<{
  cluster?: number
  groupingColumn?: string
  groupingValue?: string
  action?: string
  company?: string
  ticker?: string
}>()

const serverItems = ref<Stock[]>([])
const loading = ref(false)
const totalItems = ref(0)
const itemsPerPage = ref(10)
const search = ref('')
const currentSortBy = ref<Array<{ key: string; order: 'asc' | 'desc' }>>([
  { key: 'weighted_score', order: 'desc' },
])
const weightsChanged = ref(false)
const selectedStock = ref<Stock | null>(null)
const showStockDetail = ref(false)

// simple column filters similar to the example (sent to server)
const tickerFilter = ref('')
const companyFilter = ref('')
const actionFilter = ref('')

// Table headers
const headers = [
  {
    title: 'Ticker',
    key: 'ticker',
    align: 'start',
    sortable: true,
  },
  {
    title: 'Company',
    key: 'company',
    align: 'start',
    sortable: true,
  },
  {
    title: 'Action',
    key: 'action',
    align: 'start',
    sortable: true,
  },
  {
    title: 'Date',
    key: 'date',
    align: 'start',
    sortable: true,
  },
  {
    title: 'Final Score',
    key: 'final_score',
    align: 'end',
    sortable: true,
  },
  {
    title: 'Weighted Score',
    key: 'weighted_score',
    align: 'end',
    sortable: true,
  },
  {
    title: 'Target From',
    key: 'target_from',
    align: 'end',
    sortable: true,
  },
  {
    title: 'Target Delta',
    key: 'target_delta',
    align: 'end',
    sortable: true,
  },
  {
    title: 'Rating Sentiments',
    key: 'rating_sentiments',
    align: 'start',
    sortable: false,
  },
  {
    title: 'Analyst Targets & Ratings',
    key: 'analyst_indicators',
    align: 'start',
    sortable: false,
  },
  {
    title: 'Volatility & Range',
    key: 'volatility_indicators',
    align: 'start',
    sortable: false,
  },
  {
    title: 'Cumulative Volume / Flow',
    key: 'volume_indicators',
    align: 'start',
    sortable: false,
  },
  {
    title: 'Price Filters',
    key: 'price_indicators',
    align: 'start',
    sortable: false,
  },
] as const

// Category mapping for numerical indicators
const indicatorCategories = {
  analyst_indicators: {
    title: 'Analyst Targets & Ratings',
    indicators: ['target_from', 'target_delta', 'target_growth', 'relative_growth'],
  },
  volatility_indicators: {
    title: 'Volatility & Range',
    indicators: ['atr', 'std_dev', 'ulcer_index', 'price_distance'],
  },
  volume_indicators: {
    title: 'Cumulative Volume / Flow',
    indicators: ['obv', 'ad_line', 'pvt', 'force_index'],
  },
  price_indicators: {
    title: 'Price Filters',
    indicators: ['hlc3', 'typical_price', 'vwap', 'last_close'],
  },
}

// Helper function to filter indicators by category
function getIndicatorsByCategory(indicators: Stock['numerical_indicators'], category: string) {
  if (!indicators || !indicatorCategories[category as keyof typeof indicatorCategories]) {
    return []
  }
  const categoryIndicators =
    indicatorCategories[category as keyof typeof indicatorCategories].indicators
  return indicators.filter((ind) => categoryIndicators.includes(ind.name))
}

async function loadItems({
  page,
  itemsPerPage,
  sortBy,
}: {
  page: number
  itemsPerPage: number
  sortBy: Array<{ key: string; order: 'asc' | 'desc' }>
}) {
  loading.value = true

  // If weights have changed, force sort by weighted_score
  let sortColumn: string | undefined
  let sortOrder: 'asc' | 'desc' = 'desc'

  if (weightsChanged.value) {
    sortColumn = 'weighted_score'
    sortOrder = 'desc'
    weightsChanged.value = false
    currentSortBy.value = [{ key: 'weighted_score', order: 'desc' }]
  } else {
    sortColumn = sortBy && sortBy.length > 0 && sortBy[0] ? sortBy[0].key : undefined
    sortOrder = sortBy && sortBy.length > 0 && sortBy[0] ? sortBy[0].order : 'asc'
    currentSortBy.value = sortBy
  }

  try {
    if (props.cluster !== undefined) {
      // Always use filtered endpoint with weights
      const res = await apiService.getFilteredStocksByCluster({
        cluster: props.cluster,
        grouping_column: props.groupingColumn || '',
        grouping_value: props.groupingValue || '',
        sort_by: sortColumn,
        order: sortOrder,
        page,
        per_page: itemsPerPage,
        numerical_weights: numericalWeights.value.length > 0 ? numericalWeights.value : undefined,
        rating_weights: ratingWeights.value.length > 0 ? ratingWeights.value : undefined,
      })
      serverItems.value = res.data.map((s) => ({
        ...s,
        date: new Date(s.date).toLocaleDateString(),
      }))
      totalItems.value = res.total_count
    } else {
      // Fallback to paginated endpoint if no cluster
      const res = await apiService.getStocksPaginated({
        page,
        perPage: itemsPerPage,
        sortBy: sortColumn,
        order: sortOrder,
        cluster: props.cluster,
        action: props.action || actionFilter.value || undefined,
        company: props.company || companyFilter.value || undefined,
        ticker: props.ticker || tickerFilter.value || undefined,
      })
      serverItems.value = res.data.map((s) => ({
        ...s,
        date: new Date(s.date).toLocaleDateString(),
      }))
      totalItems.value = res.count
    }
  } catch (err) {
    console.error('Failed to load items:', err)
    serverItems.value = []
  } finally {
    loading.value = false
  }
}

// Reset search when filters change to trigger reload
watch([tickerFilter, companyFilter, actionFilter], () => {
  search.value = String(Date.now())
})

// Watch for prop changes and reload
watch(
  () => [
    props.cluster,
    props.groupingColumn,
    props.groupingValue,
    props.action,
    props.company,
    props.ticker,
  ],
  () => {
    search.value = String(Date.now())
  },
  { immediate: false },
)

// Debounced function to reload data when weights change
const debouncedReloadOnWeightChange = debounce(() => {
  weightsChanged.value = true
  search.value = String(Date.now())
}, 500)

// Watch for weight changes and reload data with weighted_score sort
watch(
  () => [stocksStore.weights.numerical, stocksStore.weights.rating],
  () => {
    debouncedReloadOnWeightChange()
  },
  { deep: true },
)

// Open stock detail modal
function openStockDetail(event: unknown, row: { item: Stock }) {
  selectedStock.value = row.item
  showStockDetail.value = true
}
</script>

<template>
  <v-data-table-server
    v-model:items-per-page="itemsPerPage"
    :headers="headers"
    :items="serverItems"
    :items-length="totalItems"
    :loading="loading"
    :search="search"
    :sort-by="currentSortBy"
    item-value="id"
    @update:options="loadItems"
    @click:row="openStockDetail"
    class="custom-blue-scrollbar"
  >
    <template v-slot:[`item.ticker`]="{ item }">
      <span class="font-semibold text-blue-600 border-b-2 border-primary pb-1 mb-1 inline-block">{{
        item.ticker
      }}</span>
    </template>

    <template v-slot:[`item.company`]="{ item }">
      <span class="text-gray-700">{{ item.company }}</span>
    </template>

    <template v-slot:[`item.action`]="{ item }">
      <v-chip size="small" color="primary" variant="flat">{{ item.action }}</v-chip>
    </template>

    <template v-slot:[`item.date`]="{ item }">
      <span class="text-gray-600">{{ item.date }}</span>
    </template>

    <template v-slot:[`item.final_score`]="{ item }">
      <span class="font-semibold">{{
        item.final_score?.toFixed ? item.final_score.toFixed(2) : item.final_score
      }}</span>
    </template>

    <template v-slot:[`item.weighted_score`]="{ item }">
      <span class="font-semibold text-end">{{
        item.weighted_score !== null && item.weighted_score !== undefined
          ? typeof item.weighted_score === 'number'
            ? item.weighted_score.toFixed(2)
            : item.weighted_score
          : '0.00'
      }}</span>
    </template>

    <template v-slot:[`item.target_from`]="{ item }">
      <span class="text-end">{{
        item.target_from !== null && item.target_from !== undefined
          ? typeof item.target_from === 'number'
            ? item.target_from.toFixed(2)
            : item.target_from
          : '-'
      }}</span>
    </template>

    <template v-slot:[`item.target_delta`]="{ item }">
      <span class="text-end">{{
        item.target_delta !== null && item.target_delta !== undefined
          ? typeof item.target_delta === 'number'
            ? item.target_delta.toFixed(2)
            : item.target_delta
          : '-'
      }}</span>
    </template>

    <template v-slot:[`item.rating_sentiments`]="{ item }">
      <div class="flex flex-wrap gap-1">
        <v-chip
          v-for="(sentiment, idx) in (item as Stock).rating_sentiments || []"
          :key="idx"
          size="x-small"
          color="secondary"
          variant="flat"
          class="text-[9px] px-1 py-0"
          density="compact"
        >
          {{ sentiment.name }}: {{ sentiment.rating }}
        </v-chip>
      </div>
    </template>

    <template v-slot:[`item.analyst_indicators`]="{ item }">
      <div class="flex flex-wrap gap-1 mb-2 mt-2">
        <IndicatorChip
          v-for="(indicator, idx) in getIndicatorsByCategory(
            (item as Stock).numerical_indicators,
            'analyst_indicators',
          )"
          :key="idx"
          :indicator-name="indicator.name"
          :indicator-norm-value="indicator.norm_value"
          :weight="stocksStore.weights.numerical[indicator.name] ?? 10"
        />
      </div>
    </template>

    <template v-slot:[`item.volatility_indicators`]="{ item }">
      <div class="flex flex-wrap gap-1 mb-2 mt-2">
        <IndicatorChip
          v-for="(indicator, idx) in getIndicatorsByCategory(
            (item as Stock).numerical_indicators,
            'volatility_indicators',
          )"
          :key="idx"
          :indicator-name="indicator.name"
          :indicator-norm-value="indicator.norm_value"
          :weight="stocksStore.weights.numerical[indicator.name] ?? 10"
        />
      </div>
    </template>

    <template v-slot:[`item.volume_indicators`]="{ item }">
      <div class="flex flex-wrap gap-1 mb-2 mt-2">
        <IndicatorChip
          v-for="(indicator, idx) in getIndicatorsByCategory(
            (item as Stock).numerical_indicators,
            'volume_indicators',
          )"
          :key="idx"
          :indicator-name="indicator.name"
          :indicator-norm-value="indicator.norm_value"
          :weight="stocksStore.weights.numerical[indicator.name] ?? 10"
        />
      </div>
    </template>

    <template v-slot:[`item.price_indicators`]="{ item }">
      <div class="flex flex-wrap gap-1 mb-2 mt-2">
        <IndicatorChip
          v-for="(indicator, idx) in getIndicatorsByCategory(
            (item as Stock).numerical_indicators,
            'price_indicators',
          )"
          :key="idx"
          :indicator-name="indicator.name"
          :indicator-norm-value="indicator.norm_value"
          :weight="stocksStore.weights.numerical[indicator.name] ?? 10"
        />
      </div>
    </template>

    <template v-slot:tfoot>
      <tr>
        <td>
          <v-text-field
            v-model="tickerFilter"
            class="ma-2"
            density="compact"
            placeholder="Search ticker..."
            hide-details
          ></v-text-field>
        </td>
        <td>
          <v-text-field
            v-model="companyFilter"
            class="ma-2"
            density="compact"
            placeholder="Search company..."
            hide-details
          ></v-text-field>
        </td>
        <td>
          <v-text-field
            v-model="actionFilter"
            class="ma-2"
            density="compact"
            placeholder="Search action..."
            hide-details
          ></v-text-field>
        </td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
      </tr>
    </template>
  </v-data-table-server>

  <!-- Stock Detail Modal -->
  <StockDetail v-model="showStockDetail" :stock="selectedStock" />
</template>

<style scoped>
.custom-blue-scrollbar::-webkit-scrollbar {
  width: 12px;
}

.custom-blue-scrollbar::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 10px;
}

.custom-blue-scrollbar::-webkit-scrollbar-thumb {
  background: #3b82f6;
  border-radius: 10px;
}

.custom-blue-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #2563eb;
}

/* Firefox */
.custom-blue-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: #3b82f6 #f1f1f1;
}

/* Sticky footer */
.sticky-footer {
  position: sticky;
  bottom: 0;
  z-index: 10;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}

/* Thin divider lines between rows */
:deep(table tbody tr) {
  border-bottom: 1px solid #e5e7eb;
}

:deep(table tbody tr:last-child) {
  border-bottom: none;
}

/* Header styling - primary color and bold */
:deep(.v-data-table__thead th) {
  color: rgb(var(--v-theme-primary)) !important;
  font-weight: bold !important;
}

:deep(table thead th) {
  color: rgb(var(--v-theme-primary)) !important;
  font-weight: bold !important;
}
</style>
