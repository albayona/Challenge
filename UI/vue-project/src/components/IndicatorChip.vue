<template>
  <v-chip
    color="primary"
    variant="flat"
    size="x-small"
    class="text-[9px] px-1 py-0"
    density="compact"
  >
    <div class="flex items-center gap-1">
      <!-- Color circle indicator -->
      <div class="w-2 h-2 rounded-full flex-shrink-0" :style="circleStyle"></div>
      <!-- Indicator name only -->
      <span>{{ indicatorName }}</span>
    </div>
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  indicatorName: string
  indicatorNormValue: number
  weight: number
}>()

// Calculate color based on norm_value (0 = worst/red, 1 = best/green)
const colorValue = computed(() => {
  const norm = Math.max(0, Math.min(1, props.indicatorNormValue))

  // Interpolate from red (worst) to green (best)
  if (norm <= 0.5) {
    // Red to Yellow (0 to 0.5)
    const ratio = norm * 2
    const r = 255
    const g = Math.round(255 * ratio)
    const b = 0
    return `rgb(${r}, ${g}, ${b})`
  } else {
    // Yellow to Green (0.5 to 1)
    const ratio = (norm - 0.5) * 2
    const r = Math.round(255 * (1 - ratio))
    const g = 255
    const b = 0
    return `rgb(${r}, ${g}, ${b})`
  }
})

// Calculate opacity based on weight (0 = transparent, 10 = opaque)
const opacity = computed(() => {
  return Math.max(0, Math.min(1, props.weight / 10))
})

// Circle style with color and opacity
const circleStyle = computed(() => {
  return {
    backgroundColor: colorValue.value,
    opacity: opacity.value,
  }
})
</script>

<style scoped></style>
