<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useStocksStore } from '@/stores/stocks'
import GroupComponent from './GroupComponent.vue'
import ClusterStatsView from '@/views/ClusterStatsView.vue'

const stocksStore = useStocksStore()
const model = ref<number | string | null>(null)

// Computed property to get sorted clusters
const sortedClusters = computed(() => {
  return [...stocksStore.clusters].sort((a, b) => a - b)
})

// Watch for tab changes and fetch stocks for the selected cluster
watch(model, async (newCluster) => {
  if (newCluster !== null && typeof newCluster === 'number') {
    await stocksStore.fetchStocksByCluster(newCluster)
  }
})

// Fetch clusters on mount
onMounted(async () => {
  await stocksStore.fetchClusters()
  // Set default tab to Cluster Stats
  model.value = 'stats'
})
</script>

<template>
  <v-card class="w-full h-full">
    <v-toolbar color="primary">
      <v-app-bar-nav-icon></v-app-bar-nav-icon>
      <v-toolbar-title class="text-lg font-semibold">Stock Clusters</v-toolbar-title>
      <v-btn icon="mdi-magnify"></v-btn>
      <v-btn icon="mdi-dots-vertical"></v-btn>

      <template v-slot:extension>
        <v-tabs v-model="model" align-tabs="center" class="w-full">
          <v-tab value="stats" text="Cluster Stats"></v-tab>
          <v-tab
            v-for="cluster in sortedClusters"
            :key="cluster"
            :value="cluster"
            :text="`Cluster ${cluster}`"
          ></v-tab>
        </v-tabs>
      </template>
    </v-toolbar>

    <v-tabs-window v-model="model">
      <v-tabs-window-item value="stats">
        <ClusterStatsView />
      </v-tabs-window-item>
      <v-tabs-window-item v-for="cluster in sortedClusters" :key="cluster" :value="cluster">
        <GroupComponent :cluster="cluster" />
      </v-tabs-window-item>
    </v-tabs-window>
  </v-card>
</template>
