<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useStocksStore } from '@/stores/stocks'
import StocksTable from './StocksTable.vue'
import WeightSlider from './WeightSlider.vue'
import DashboardPanel from './DashboardPanel.vue'

const props = defineProps<{
  cluster: number
}>()

const stocksStore = useStocksStore()
const groupBy = ref<'action' | 'date' | 'company' | 'rating_to' | 'rating_from' | 'none'>('none')
const expanded = ref<string[]>([])

// Group by options from store
const groupByOptions = computed(() => {
  const options = [{ title: 'None', value: 'none' }]
  const columnLabels: Record<string, string> = {
    action: 'Action',
    rating_to: 'Rating To',
    rating_from: 'Rating From',
  }

  stocksStore.groupingColumns.forEach((column) => {
    options.push({
      title: columnLabels[column] || column,
      value: column,
    })
  })

  return options
})

// Watch for groupBy changes and fetch unique values
watch(
  groupBy,
  async (newGroupBy) => {
    expanded.value = []
    if (newGroupBy && newGroupBy !== 'none' && props.cluster !== undefined) {
      await stocksStore.fetchUniqueValues(props.cluster, newGroupBy)
    }
  },
  { immediate: true },
)

// Watch for cluster changes
watch(
  () => props.cluster,
  async (newCluster) => {
    if (newCluster !== null && newCluster !== undefined && groupBy.value !== 'none') {
      await stocksStore.fetchUniqueValues(newCluster, groupBy.value)
    }
  },
  { immediate: true },
)

// Build tree view items from unique values
const treeItems = computed(() => {
  if (!stocksStore.uniqueValues || stocksStore.uniqueValues.values.length === 0) {
    return []
  }

  return stocksStore.uniqueValues.values.map((value, index) => ({
    id: value,
    name: value,
    children: [],
    index,
  }))
})

// Toggle expand/collapse
const toggleExpand = (itemId: string) => {
  const index = expanded.value.indexOf(itemId)
  if (index > -1) {
    expanded.value.splice(index, 1)
  } else {
    expanded.value.push(itemId)
  }
}

// Loading state
const isLoading = computed(() => stocksStore.loading || stocksStore.uniqueValuesLoading)

// Indicator definitions - update these with your actual indicator lists
const numericalIndicators = ref<Record<string, number>>({
  target_from: 10,
  target_to: 10,
  target_delta: 10,
  target_growth: 10,
  relative_growth: 10,
  last_close: 10,
  atr: 10,
  std_dev: 10,
  ulcer_index: 10,
  price_distance: 10,
  obv: 10,
  ad_line: 10,
  pvt: 10,
  force_index: 10,
  hlc3: 10,
  typical_price: 10,
  vwap: 10,
})

const ratingIndicators = ref<Record<string, number>>({
  rating_from: 10,
  rating_to: 10,
  action: 10,
})

// Computed subtitle with cluster number
const weightConfigSubtitle = computed(() => `Adjust indicator weights for Cluster ${props.cluster}`)
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Weight Configuration Panel -->
    <DashboardPanel title="Indicator Weights" :subtitle="weightConfigSubtitle">
      <!-- Weight Sliders -->
      <WeightSlider
        :numerical-indicators="numericalIndicators"
        :rating-indicators="ratingIndicators"
      />
    </DashboardPanel>

    <!-- Data Display Panel -->
    <DashboardPanel
      v-if="groupBy === 'none'"
      title="Stock Data"
      subtitle="All stocks for this cluster"
    >
      <!-- Group By Selector -->
      <div class="mb-4">
        <v-select
          v-model="groupBy"
          :items="groupByOptions"
          label="Group By"
          variant="outlined"
          density="compact"
          class="w-32"
        ></v-select>
      </div>
      <StocksTable :cluster="cluster" />
    </DashboardPanel>

    <!-- Grouped Data Display Panel -->
    <DashboardPanel v-else title="Grouped Stock Data" :subtitle="`Stocks grouped by ${groupBy}`">
      <!-- Group By Selector -->
      <div class="mb-4">
        <v-select
          v-model="groupBy"
          :items="groupByOptions"
          label="Group By"
          variant="outlined"
          density="compact"
          class="w-32"
        ></v-select>
      </div>

      <div v-if="isLoading" class="flex justify-center items-center py-12">
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
      </div>

      <div v-else-if="stocksStore.uniqueValuesError" class="text-red-500 p-4">
        <p class="font-semibold">Error:</p>
        <p>{{ stocksStore.uniqueValuesError || 'Unknown error' }}</p>
      </div>

      <div
        v-else-if="!stocksStore.uniqueValues || stocksStore.uniqueValues.values.length === 0"
        class="text-gray-500 text-center py-12"
      >
        <p>No unique values found</p>
      </div>

      <div v-else class="w-full border border-gray-300 rounded-lg overflow-hidden">
        <!-- Custom styled list with table-like appearance -->
        <div
          v-for="(item, index) in treeItems"
          :key="item.id"
          :class="[
            'border-b border-gray-300 last:border-b-0',
            index % 2 === 0 ? 'bg-white' : 'bg-gray-50',
          ]"
        >
          <!-- Header row - clickable to expand/collapse -->
          <div
            @click="toggleExpand(item.id)"
            :class="[
              'px-4 py-3 cursor-pointer transition-colors flex items-center justify-between',
              expanded.includes(item.id)
                ? 'bg-blue-600 text-white hover:bg-blue-700'
                : 'bg-white text-blue-600 hover:bg-blue-50 border-l-4 border-l-blue-600',
            ]"
          >
            <span
              :class="[
                'text-lg font-bold',
                expanded.includes(item.id) ? 'text-white' : 'text-blue-600',
              ]"
            >
              {{ item.name }}
            </span>
            <v-icon
              :class="['transition-transform', expanded.includes(item.id) ? 'rotate-90' : '']"
              :color="expanded.includes(item.id) ? 'white' : 'primary'"
            >
              mdi-chevron-right
            </v-icon>
          </div>

          <!-- Collapsible content with table -->
          <div
            v-if="expanded.includes(item.id)"
            class="border-t border-gray-300 bg-white px-4 py-4"
          >
            <StocksTable :cluster="cluster" :grouping-column="groupBy" :grouping-value="item.id" />
          </div>
        </div>
      </div>
    </DashboardPanel>
  </div>
</template>
