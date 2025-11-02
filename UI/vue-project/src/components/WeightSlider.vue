<template>
  <div class="flex flex-wrap gap-4">
    <!-- Category Panels -->
    <DashboardPanel
      v-for="category in categoryPanels"
      :key="category.key"
      :title="category.title"
      :subtitle="category.subtitle"
      class="flex-1 min-w-[300px]"
    >
      <div class="flex flex-wrap gap-3">
        <v-menu
          v-for="indicatorName in category.indicators"
          :key="`${category.key}-${indicatorName}`"
          location="bottom"
        >
          <template v-slot:activator="{ props: menuProps }">
            <v-card
              v-bind="menuProps"
              variant="outlined"
              class="cursor-pointer hover:bg-gray-50 transition-colors w-fit"
            >
              <v-card-item class="px-2 py-1">
                <div class="flex items-center justify-between gap-1">
                  <span
                    class="text-[10px] font-medium text-gray-700 uppercase tracking-wide flex-1"
                  >
                    {{ indicatorName }}
                  </span>
                  <v-chip size="x-small" color="secondary" variant="flat" class="text-[10px]">
                    {{ getWeight(category.type, indicatorName) }}
                  </v-chip>
                </div>
              </v-card-item>
            </v-card>
          </template>

          <v-card min-width="300" class="pa-4">
            <div class="text-sm font-semibold mb-3 text-gray-700">{{ indicatorName }}</div>
            <div class="max-w-[280px]">
              <v-slider
                :model-value="getWeight(category.type, indicatorName)"
                color="secondary"
                :step="1"
                :ticks="{ 0: '0', 5: '5', 10: '10' }"
                :tick-size="4"
                max="10"
                min="0"
                track-color="grey-lighten-2"
                density="compact"
                hide-details
                @update:model-value="
                  (value) => {
                    const actualType =
                      category.type === 'mixed' &&
                      ['rating_from', 'rating_to', 'action'].includes(indicatorName)
                        ? 'rating'
                        : 'numerical'
                    stocksStore.updateWeight(actualType, indicatorName, value)
                  }
                "
              >
                <template v-slot:prepend>
                  <v-btn
                    color="secondary"
                    icon="mdi-minus"
                    size="x-small"
                    variant="text"
                    density="compact"
                    @click="decrement(category.type, indicatorName)"
                  ></v-btn>
                </template>
                <template v-slot:append>
                  <v-btn
                    color="secondary"
                    icon="mdi-plus"
                    size="x-small"
                    variant="text"
                    density="compact"
                    @click="increment(category.type, indicatorName)"
                  ></v-btn>
                </template>
              </v-slider>
            </div>
          </v-card>
        </v-menu>
      </div>

      <!-- Reset Button for Category -->
      <div class="flex items-center justify-end mt-2 pt-2 border-t border-gray-200">
        <v-btn
          size="small"
          color="secondary"
          variant="outlined"
          density="compact"
          @click="toggleResetCategory(category)"
        >
          Reset
        </v-btn>
      </div>
    </DashboardPanel>

    <div v-if="categoryPanels.length === 0" class="text-center py-4 text-gray-500 text-sm">
      <p>No indicators configured. Please provide the indicator list.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStocksStore } from '@/stores/stocks'
import DashboardPanel from './DashboardPanel.vue'

const props = defineProps<{
  numericalIndicators?: Record<string, number>
  ratingIndicators?: Record<string, number>
}>()

const stocksStore = useStocksStore()

// Get weights from store
const weights = computed(() => stocksStore.weights)

// Category mapping based on the documentation
const indicatorCategories = {
  'analyst-targets-ratings': {
    title: 'Analyst Targets & Ratings',
    subtitle: 'Analyst expectations, ratings, and target price movements',
    indicators: ['target_from', 'target_to', 'target_delta', 'target_growth', 'relative_growth'],
    ratingIndicators: ['rating_from', 'rating_to', 'action'],
    type: 'mixed' as const,
  },
  'volatility-range': {
    title: 'Volatility & Range',
    subtitle: 'Price volatility, range indicators, and market stability measures',
    indicators: ['atr', 'std_dev', 'ulcer_index', 'price_distance'],
    type: 'numerical' as const,
  },
  'cumulative-volume-flow': {
    title: 'Cumulative Volume / Flow',
    subtitle: 'Volume accumulation, distribution, and flow indicators',
    indicators: ['obv', 'ad_line', 'pvt', 'force_index'],
    type: 'numerical' as const,
  },
  'price-filters': {
    title: 'Price Filters (Level Indicators)',
    subtitle: 'Price level indicators and volume-weighted averages',
    indicators: ['hlc3', 'typical_price', 'vwap', 'last_close'],
    type: 'numerical' as const,
  },
}

// Build category panels based on available indicators
const categoryPanels = computed(() => {
  const panels: Array<{
    key: string
    title: string
    subtitle: string
    indicators: string[]
    type: 'numerical' | 'rating' | 'mixed'
  }> = []

  Object.entries(indicatorCategories).forEach(([key, category]) => {
    const availableIndicators: string[] = []

    // Check numerical indicators
    if (props.numericalIndicators) {
      const numerical = props.numericalIndicators
      category.indicators.forEach((indicator) => {
        if (indicator in numerical) {
          availableIndicators.push(indicator)
        }
      })
    }

    // Check rating indicators for analyst category
    if (category.type === 'mixed' && props.ratingIndicators) {
      const rating = props.ratingIndicators
      category.ratingIndicators?.forEach((indicator) => {
        if (indicator in rating) {
          availableIndicators.push(indicator)
        }
      })
    }

    if (availableIndicators.length > 0) {
      panels.push({
        key,
        title: category.title,
        subtitle: category.subtitle,
        indicators: availableIndicators,
        type: category.type,
      })
    }
  })

  return panels
})

// Get weight value for an indicator
function getWeight(type: 'numerical' | 'rating' | 'mixed', indicatorName: string): number {
  // Check if it's a rating indicator
  const ratingIndicators = ['rating_from', 'rating_to', 'action']
  if (type === 'mixed' && ratingIndicators.includes(indicatorName)) {
    return weights.value.rating[indicatorName] ?? 10
  }
  // Default to numerical
  return weights.value.numerical[indicatorName] ?? 10
}

// Initialize weights if props are provided
if (props.numericalIndicators || props.ratingIndicators) {
  stocksStore.initializeWeights(props.numericalIndicators || {}, props.ratingIndicators || {})
}

function decrement(type: 'numerical' | 'rating' | 'mixed', indicatorName: string) {
  const actualType =
    type === 'mixed' && ['rating_from', 'rating_to', 'action'].includes(indicatorName)
      ? 'rating'
      : 'numerical'
  const currentValue = getWeight(actualType, indicatorName)
  if (currentValue > 0) {
    stocksStore.updateWeight(actualType, indicatorName, currentValue - 1)
  }
}

function increment(type: 'numerical' | 'rating' | 'mixed', indicatorName: string) {
  const actualType =
    type === 'mixed' && ['rating_from', 'rating_to', 'action'].includes(indicatorName)
      ? 'rating'
      : 'numerical'
  const currentValue = getWeight(actualType, indicatorName)
  if (currentValue < 10) {
    stocksStore.updateWeight(actualType, indicatorName, currentValue + 1)
  }
}

// Get average weight for a category
function getCategoryAverageWeight(category: {
  indicators: string[]
  type: 'numerical' | 'rating' | 'mixed'
}): number {
  const ratingIndicators = ['rating_from', 'rating_to', 'action']
  const weights: number[] = []

  category.indicators.forEach((indicatorName) => {
    const actualType =
      category.type === 'mixed' && ratingIndicators.includes(indicatorName)
        ? 'rating'
        : 'numerical'
    weights.push(getWeight(actualType, indicatorName))
  })

  if (weights.length === 0) return 10
  const sum = weights.reduce((acc, val) => acc + val, 0)
  return Math.round(sum / weights.length)
}

// Update all weights in a category to the same value
function updateCategoryWeights(
  category: { indicators: string[]; type: 'numerical' | 'rating' | 'mixed' },
  value: number,
) {
  const ratingIndicators = ['rating_from', 'rating_to', 'action']
  category.indicators.forEach((indicatorName) => {
    const actualType =
      category.type === 'mixed' && ratingIndicators.includes(indicatorName)
        ? 'rating'
        : 'numerical'
    stocksStore.updateWeight(actualType, indicatorName, value)
  })
}

// Decrement all weights in a category
function decrementCategory(category: {
  indicators: string[]
  type: 'numerical' | 'rating' | 'mixed'
}) {
  const currentAvg = getCategoryAverageWeight(category)
  if (currentAvg > 0) {
    updateCategoryWeights(category, currentAvg - 1)
  }
}

// Increment all weights in a category
function incrementCategory(category: {
  indicators: string[]
  type: 'numerical' | 'rating' | 'mixed'
}) {
  const currentAvg = getCategoryAverageWeight(category)
  if (currentAvg < 10) {
    updateCategoryWeights(category, currentAvg + 1)
  }
}

// Toggle reset: if average is 10, set to 0; otherwise set to 10
function toggleResetCategory(category: {
  indicators: string[]
  type: 'numerical' | 'rating' | 'mixed'
}) {
  const currentAvg = getCategoryAverageWeight(category)
  const newValue = currentAvg === 10 ? 0 : 10
  updateCategoryWeights(category, newValue)
}
</script>

<style scoped>
.v-slider {
  margin: 0;
}
</style>
