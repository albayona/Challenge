<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import DashboardPanel from '@/components/DashboardPanel.vue'
import type { ClusterStats } from '@/types/stock'
import PlotlyChart from '@/components/PlotlyChart.vue'

// Load cluster statistics from JSON file
const clusterStats = ref<ClusterStats[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

// Load cluster stats from JSON file
async function loadClusterStats() {
  try {
    loading.value = true
    error.value = null
    const response = await fetch('/data/cluster_analysis_percentages.json')
    if (!response.ok) {
      throw new Error(`Failed to load cluster stats: ${response.statusText}`)
    }
    const data = await response.json()
    console.log('Loaded cluster stats:', data)

    // Handle both array and object formats
    if (Array.isArray(data)) {
      clusterStats.value = data as ClusterStats[]
    } else if (data.clusters && Array.isArray(data.clusters)) {
      clusterStats.value = data.clusters as ClusterStats[]
    } else {
      // If it's an object with cluster data directly
      clusterStats.value = [data] as ClusterStats[]
    }

    console.log('Parsed cluster stats:', clusterStats.value)

    // Initialize selected cluster and feature
    if (clusterStats.value.length > 0 && clusterStats.value[0]) {
      selectedCluster.value = clusterStats.value[0].cluster
      const firstStats = clusterStats.value[0]
      if (firstStats && Object.keys(firstStats.features).length > 0) {
        selectedFeature.value = Object.keys(firstStats.features)[0] || null
      }
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load cluster statistics'
    console.error('Error loading cluster stats:', err)
  } finally {
    loading.value = false
  }
}

// Load data on component mount
onMounted(() => {
  loadClusterStats()
})

const selectedCluster = ref<number | null>(null)
const expandedFeatures = ref<string[]>([])
const selectedFeature = ref<string | null>(null)

// Get selected cluster stats
const selectedStats = computed(() => {
  console.log('selectedStats computed:', {
    selectedCluster: selectedCluster.value,
    clusterStatsLength: clusterStats.value.length,
    clusterStats: clusterStats.value,
  })
  if (selectedCluster.value === null) return null
  const found = clusterStats.value.find((stats) => stats.cluster === selectedCluster.value)
  console.log('Found stats:', found)
  return found || null
})

// Get available features (from first cluster, assuming all clusters have same features)
const availableFeatures = computed(() => {
  if (clusterStats.value.length === 0) return []
  return Object.keys(clusterStats.value[0]?.features || {})
})

// Prepare distribution data for all clusters
const histogramData = computed(() => {
  if (!selectedFeature.value || clusterStats.value.length === 0) {
    console.log('histogramData: No feature selected or no cluster stats')
    return []
  }

  const colors = [
    '#3b82f6', // blue
    '#10b981', // green
    '#f59e0b', // amber
    '#ef4444', // red
    '#8b5cf6', // purple
  ]

  const traces = clusterStats.value
    .map((clusterStat, index) => {
      const featureData = clusterStat.features[selectedFeature.value!]
      console.log(
        'Processing cluster',
        clusterStat.cluster,
        'feature',
        selectedFeature.value,
        'data:',
        featureData,
      )

      if (!featureData) {
        console.log('No featureData for cluster', clusterStat.cluster)
        return null
      }

      if (!featureData.distribution || featureData.distribution.length === 0) {
        console.log(
          'No distribution data for cluster',
          clusterStat.cluster,
          'feature',
          selectedFeature.value,
        )
        return null
      }

      // Use distribution data if available
      const distribution = featureData.distribution
      
      // Calculate x positions (midpoint of each range) and widths (range span)
      const xPositions = distribution.map((bin) => (bin.range[0] + bin.range[1]) / 2)
      const widths = distribution.map((bin) => bin.range[1] - bin.range[0])
      const percentages = distribution.map((bin) => bin.percentage)

      console.log('Creating trace for cluster', clusterStat.cluster, {
        xPositions,
        widths,
        percentages,
      })

      // Validate data before creating trace
      if (xPositions.length === 0 || percentages.length === 0 || widths.length === 0) {
        console.warn('Empty data arrays for cluster', clusterStat.cluster)
        return null
      }

      const trace = {
        x: xPositions,
        y: percentages,
        width: widths,
        type: 'bar',
        name: `Cluster ${clusterStat.cluster}`,
        marker: { color: colors[index % colors.length] },
        opacity: 0.7,
      }

      console.log('Created trace:', trace)
      return trace
    })
    .filter((item): item is NonNullable<typeof item> => item !== null)

  console.log('Final histogramData traces:', traces)
  return traces
})

// Distribution chart layout
const histogramLayout = computed(() => ({
  title: {
    text: `Distribution: ${selectedFeature.value || 'Select Feature'}`,
    font: { size: 16 },
  },
  xaxis: {
    title: { text: 'Value Range' },
    type: 'linear',
  },
  yaxis: {
    title: { text: 'Percentage' },
    tickformat: '.1%',
  },
  barmode: 'overlay',
  height: 500,
  showlegend: true,
  legend: {
    x: 1,
    y: 1,
  },
}))

// Toggle feature expansion
const toggleFeature = (featureName: string) => {
  const index = expandedFeatures.value.indexOf(featureName)
  if (index > -1) {
    expandedFeatures.value.splice(index, 1)
  } else {
    expandedFeatures.value.push(featureName)
  }
}

// Format number for display
function formatNumber(value: number): string {
  if (value === null || value === undefined) return '-'
  if (typeof value !== 'number') return String(value)
  if (Math.abs(value) >= 1000) {
    return value.toFixed(2)
  }
  return value.toFixed(4)
}
</script>

<template>
  <div class="p-6 space-y-6">
    <DashboardPanel title="Cluster Statistics" subtitle="Detailed statistics for each cluster">
      <div class="space-y-4">
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="primary"></v-progress-circular>
          <p class="mt-4 text-gray-500">Loading cluster statistics...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-8 text-red-500">
          <p class="font-semibold">Error loading cluster statistics</p>
          <p class="text-sm mt-2">{{ error }}</p>
        </div>

        <!-- No Data State -->
        <div
          v-else-if="!selectedStats || clusterStats.length === 0"
          class="text-center py-8 text-gray-500"
        >
          <p>No cluster statistics available. Please ensure data is loaded.</p>
        </div>

        <!-- Cluster Statistics Table -->
        <div v-else-if="selectedStats" class="space-y-6">
          <!-- General Cluster Info -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div class="border border-gray-300 rounded-lg p-4">
              <div class="text-sm text-gray-600 mb-1">Inertia</div>
              <div class="text-lg font-semibold">{{ formatNumber(selectedStats.inertia) }}</div>
            </div>
            <div class="border border-gray-300 rounded-lg p-4">
              <div class="text-sm text-gray-600 mb-1">Silhouette Mean</div>
              <div class="text-lg font-semibold">
                {{ formatNumber(selectedStats.silhouette.mean) }}
              </div>
            </div>
            <div class="border border-gray-300 rounded-lg p-4">
              <div class="text-sm text-gray-600 mb-1">Sample Count</div>
              <div class="text-lg font-semibold">
                {{
                  (() => {
                    const firstFeatureKey = Object.keys(selectedStats.features)[0]
                    return firstFeatureKey ? selectedStats.features[firstFeatureKey]?.count || 0 : 0
                  })()
                }}
              </div>
            </div>
          </div>

          <!-- Features Table and Histogram Side by Side -->
          <v-row no-gutters>
            <!-- Features Table -->
            <v-col>
              <DashboardPanel
                title="Feature Statistics"
                subtitle="Statistical analysis for each feature"
              >
                <!-- Cluster Selector -->
                <div class="mb-4">
                  <v-select
                    v-model="selectedCluster"
                    :items="
                      clusterStats.map((s) => ({ title: `Cluster ${s.cluster}`, value: s.cluster }))
                    "
                    label="Select Cluster"
                    variant="outlined"
                    density="compact"
                    class="max-w-xs"
                  ></v-select>
                </div>
                <div class="overflow-x-auto w-full">
                  <v-table class="w-full text-xs">
                    <thead>
                      <tr>
                        <th class="text-left text-xs px-2 py-2">Feature</th>
                        <th class="text-right text-xs px-2 py-2">Count</th>
                        <th class="text-right text-xs px-2 py-2">Mean</th>
                        <th class="text-right text-xs px-2 py-2">Std</th>
                        <th class="text-right text-xs px-2 py-2">Min</th>
                        <th class="text-right text-xs px-2 py-2">25%</th>
                        <th class="text-right text-xs px-2 py-2">50%</th>
                        <th class="text-right text-xs px-2 py-2">75%</th>
                        <th class="text-right text-xs px-2 py-2">Max</th>
                        <th class="text-right text-xs px-2 py-2">F-Value</th>
                        <th class="text-right text-xs px-2 py-2">P-Value</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template
                        v-for="(stats, featureName) in selectedStats.features"
                        :key="featureName"
                      >
                        <tr
                          class="hover:bg-gray-50 cursor-pointer"
                          @click="toggleFeature(featureName)"
                        >
                          <td class="font-semibold text-xs px-2 py-2">{{ featureName }}</td>
                          <td class="text-right text-xs px-2 py-2">{{ stats.count }}</td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.mean) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.std) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.min) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats['25%']) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats['50%']) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats['75%']) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.max) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.f_value) }}
                          </td>
                          <td class="text-right text-xs px-2 py-2">
                            {{ formatNumber(stats.p_value) }}
                          </td>
                        </tr>
                      </template>
                    </tbody>
                  </v-table>
                </div>
              </DashboardPanel>
            </v-col>

            <!-- Histogram Panel -->
            <v-col>
              <DashboardPanel
                title="Feature Distribution Histogram"
                subtitle="Compare feature distributions across all clusters"
              >
                <div class="space-y-4">
                  <!-- Feature Selector -->
                  <div class="mb-4">
                    <v-select
                      v-model="selectedFeature"
                      :items="availableFeatures.map((f) => ({ title: f, value: f }))"
                      label="Select Feature"
                      variant="outlined"
                      density="compact"
                      class="max-w-xs"
                    ></v-select>
                  </div>

                  <!-- Histogram -->
                  <div v-if="selectedFeature && histogramData.length > 0" class="w-full">
                    <PlotlyChart :data="histogramData" :layout="histogramLayout" :height="500" />
                  </div>
                  <div v-else class="text-center py-8 text-gray-500">
                    <p>Select a feature to display the histogram</p>
                    <p v-if="selectedFeature && histogramData.length === 0" class="text-xs mt-2">
                      No data available for selected feature
                    </p>
                  </div>
                </div>
              </DashboardPanel>
            </v-col>
          </v-row>
        </div>
      </div>
    </DashboardPanel>
  </div>
</template>
