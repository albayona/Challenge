<template>
  <v-dialog v-model="dialog" max-width="1200" scrollable>
    <v-card v-if="stock">
      <v-card-title class="d-flex align-center justify-space-between bg-primary text-white">
        <div>
          <div class="text-h5 font-bold">{{ stock.ticker }}</div>
          <div class="text-subtitle-1">{{ stock.company }}</div>
        </div>
        <v-btn icon="mdi-close" variant="text" @click="close"></v-btn>
      </v-card-title>

      <v-card-text class="pa-6">
        <!-- Basic Information -->
        <v-card class="mb-4" variant="outlined">
          <v-card-title class="text-h6">Basic Information</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" md="3">
                <div class="text-caption text-gray-600">Action</div>
                <v-chip size="small" color="primary" variant="flat">{{ stock.action }}</v-chip>
              </v-col>
              <v-col cols="12" md="3">
                <div class="text-caption text-gray-600">Date</div>
                <div class="text-body-1">{{ new Date(stock.date).toLocaleDateString() }}</div>
              </v-col>
              <v-col cols="12" md="3">
                <div class="text-caption text-gray-600">Cluster</div>
                <div class="text-body-1">Cluster {{ stock.cluster }}</div>
              </v-col>
              <v-col cols="12" md="3">
                <div class="text-caption text-gray-600">Final Score</div>
                <div class="text-h6 font-bold">
                  {{
                    stock.final_score?.toFixed ? stock.final_score.toFixed(2) : stock.final_score
                  }}
                </div>
              </v-col>
              <v-col
                cols="12"
                md="3"
                v-if="stock.weighted_score !== undefined && stock.weighted_score !== null"
              >
                <div class="text-caption text-gray-600">Weighted Score</div>
                <div class="text-h6 font-bold text-primary">
                  {{
                    typeof stock.weighted_score === 'number'
                      ? stock.weighted_score.toFixed(2)
                      : stock.weighted_score
                  }}
                </div>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Rating Sentiments -->
        <v-card
          v-if="stock.rating_sentiments && stock.rating_sentiments.length > 0"
          class="mb-4"
          variant="outlined"
        >
          <v-card-title class="text-h6">Rating Sentiments</v-card-title>
          <v-card-text>
            <div class="flex flex-wrap gap-2">
              <v-chip
                v-for="(sentiment, idx) in stock.rating_sentiments"
                :key="idx"
                size="small"
                color="secondary"
                variant="flat"
              >
                {{ sentiment.name }}: {{ sentiment.rating }}
              </v-chip>
            </div>
          </v-card-text>
        </v-card>

        <!-- Analyst Targets & Ratings -->
        <v-card class="mb-4" variant="outlined">
          <v-card-title class="text-h6">Analyst Targets & Ratings</v-card-title>
          <v-card-text>
            <v-row>
              <v-col
                cols="12"
                md="3"
                v-if="stock.target_from !== undefined && stock.target_from !== null"
              >
                <div class="text-caption text-gray-600">Target From</div>
                <div class="text-body-1 font-semibold">
                  {{
                    typeof stock.target_from === 'number'
                      ? stock.target_from.toFixed(2)
                      : stock.target_from
                  }}
                </div>
              </v-col>
              <v-col
                cols="12"
                md="3"
                v-if="stock.target_to !== undefined && stock.target_to !== null"
              >
                <div class="text-caption text-gray-600">Target To</div>
                <div class="text-body-1 font-semibold">
                  {{
                    typeof stock.target_to === 'number'
                      ? stock.target_to.toFixed(2)
                      : stock.target_to
                  }}
                </div>
              </v-col>
              <v-col
                cols="12"
                md="3"
                v-if="stock.target_delta !== undefined && stock.target_delta !== null"
              >
                <div class="text-caption text-gray-600">Target Delta</div>
                <div class="text-body-1 font-semibold">
                  {{
                    typeof stock.target_delta === 'number'
                      ? stock.target_delta.toFixed(2)
                      : stock.target_delta
                  }}
                </div>
              </v-col>
              <v-col cols="12" md="3" v-if="stock.rating_from">
                <div class="text-caption text-gray-600">Rating From</div>
                <div class="text-body-1">{{ stock.rating_from }}</div>
              </v-col>
              <v-col cols="12" md="3" v-if="stock.rating_to">
                <div class="text-caption text-gray-600">Rating To</div>
                <div class="text-body-1">{{ stock.rating_to }}</div>
              </v-col>
            </v-row>
            <v-row v-if="analystIndicators.length > 0">
              <v-col cols="12" v-for="indicator in analystIndicators" :key="indicator.id">
                <div class="d-flex justify-space-between align-center">
                  <div>
                    <div class="text-body-2 font-medium">{{ indicator.name }}</div>
                    <div class="text-caption text-gray-600">
                      Value: {{ indicator.value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                  <div class="text-end">
                    <div class="text-caption text-gray-600">Normalized</div>
                    <div class="text-body-1 font-semibold">
                      {{ indicator.norm_value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                </div>
                <v-divider class="mt-2"></v-divider>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Volatility & Range -->
        <v-card v-if="volatilityIndicators.length > 0" class="mb-4" variant="outlined">
          <v-card-title class="text-h6">Volatility & Range</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" v-for="indicator in volatilityIndicators" :key="indicator.id">
                <div class="d-flex justify-space-between align-center">
                  <div>
                    <div class="text-body-2 font-medium">{{ indicator.name }}</div>
                    <div class="text-caption text-gray-600">
                      Value: {{ indicator.value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                  <div class="text-end">
                    <div class="text-caption text-gray-600">Normalized</div>
                    <div class="text-body-1 font-semibold">
                      {{ indicator.norm_value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                </div>
                <v-divider class="mt-2"></v-divider>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Cumulative Volume / Flow -->
        <v-card v-if="volumeIndicators.length > 0" class="mb-4" variant="outlined">
          <v-card-title class="text-h6">Cumulative Volume / Flow</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" v-for="indicator in volumeIndicators" :key="indicator.id">
                <div class="d-flex justify-space-between align-center">
                  <div>
                    <div class="text-body-2 font-medium">{{ indicator.name }}</div>
                    <div class="text-caption text-gray-600">
                      Value: {{ indicator.value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                  <div class="text-end">
                    <div class="text-caption text-gray-600">Normalized</div>
                    <div class="text-body-1 font-semibold">
                      {{ indicator.norm_value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                </div>
                <v-divider class="mt-2"></v-divider>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Price Filters -->
        <v-card v-if="priceIndicators.length > 0" class="mb-4" variant="outlined">
          <v-card-title class="text-h6">Price Filters</v-card-title>
          <v-card-text>
            <v-row>
              <v-col cols="12" v-for="indicator in priceIndicators" :key="indicator.id">
                <div class="d-flex justify-space-between align-center">
                  <div>
                    <div class="text-body-2 font-medium">{{ indicator.name }}</div>
                    <div class="text-caption text-gray-600">
                      Value: {{ indicator.value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                  <div class="text-end">
                    <div class="text-caption text-gray-600">Normalized</div>
                    <div class="text-body-1 font-semibold">
                      {{ indicator.norm_value?.toFixed(4) || '-' }}
                    </div>
                  </div>
                </div>
                <v-divider class="mt-2"></v-divider>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" variant="flat" @click="close">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Stock, NumericalIndicator } from '@/types/stock'

const props = defineProps<{
  modelValue: boolean
  stock: Stock | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const dialog = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

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
function getIndicatorsByCategory(
  indicators: NumericalIndicator[] | undefined,
  category: string,
): NumericalIndicator[] {
  if (!indicators || !indicatorCategories[category as keyof typeof indicatorCategories]) {
    return []
  }
  const categoryIndicators =
    indicatorCategories[category as keyof typeof indicatorCategories].indicators
  return indicators.filter((ind) => categoryIndicators.includes(ind.name))
}

// Computed properties for each category
const analystIndicators = computed(() => {
  if (!props.stock) return []
  return getIndicatorsByCategory(props.stock.numerical_indicators, 'analyst_indicators')
})

const volatilityIndicators = computed(() => {
  if (!props.stock) return []
  return getIndicatorsByCategory(props.stock.numerical_indicators, 'volatility_indicators')
})

const volumeIndicators = computed(() => {
  if (!props.stock) return []
  return getIndicatorsByCategory(props.stock.numerical_indicators, 'volume_indicators')
})

const priceIndicators = computed(() => {
  if (!props.stock) return []
  return getIndicatorsByCategory(props.stock.numerical_indicators, 'price_indicators')
})

function close() {
  dialog.value = false
}
</script>

<style scoped>
.v-card {
  overflow: visible;
}
</style>
