<template>
  <div ref="plotContainer" :style="{ width: '100%', height: height + 'px' }"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
// @ts-expect-error - plotly.js-dist-min types may not be perfect
import * as Plotly from 'plotly.js-dist-min'

const props = defineProps<{
  data: unknown[]
  layout?: Record<string, unknown>
  config?: Record<string, unknown>
  height?: number
}>()

const plotContainer = ref<HTMLElement | null>(null)
const plotInstance = ref<HTMLElement | null>(null)

const defaultHeight = 500

const renderPlot = async () => {
  if (!plotContainer.value || !props.data || props.data.length === 0) return

  // Validate and clean data
  interface PlotlyTrace {
    x: unknown[]
    type?: string
    [key: string]: unknown
  }

  const cleanedData = props.data
    .map((trace: unknown) => {
      const t = trace as PlotlyTrace
      if (!t || !Array.isArray(t.x)) return null

      // For categorical/bar charts, x can be strings
      // For numeric charts (histogram), x must be numbers
      const isCategorical = t.type === 'bar' || t.type === 'scatter'
      let cleanedX: unknown[]

      if (isCategorical) {
        // For categorical charts, allow strings and numbers, just filter out null/undefined/NaN
        cleanedX = t.x.filter((val: unknown) => {
          if (val === null || val === undefined) return false
          if (typeof val === 'string') return val.length > 0
          if (typeof val === 'number') return !isNaN(val) && isFinite(val)
          return true
        })
      } else {
        // For numeric charts (like histogram), only allow valid numbers
        cleanedX = t.x.filter((val: unknown) => {
          return typeof val === 'number' && !isNaN(val) && isFinite(val)
        }) as number[]
      }

      if (cleanedX.length === 0) return null

      return {
        ...t,
        x: cleanedX,
      }
    })
    .filter((trace: unknown) => trace !== null)

  if (cleanedData.length === 0) {
    console.warn('No valid data to plot')
    return
  }

  // Validate and clean layout
  const baseLayout = {
    height: props.height || defaultHeight,
    autosize: true,
    margin: { l: 60, r: 50, t: 60, b: 50 },
  }

  const layoutXaxis = (props.layout as Record<string, unknown>)?.xaxis as
    | Record<string, unknown>
    | undefined
  const layoutYaxis = (props.layout as Record<string, unknown>)?.yaxis as
    | Record<string, unknown>
    | undefined

  const plotLayout = {
    ...baseLayout,
    ...(props.layout || {}),
    // Ensure axis configurations don't have NaN values
    xaxis: {
      automargin: true,
      ...(layoutXaxis || {}),
    },
    yaxis: {
      automargin: true,
      ...(layoutYaxis || {}),
    },
  }

  const plotConfig = {
    displayModeBar: false,
    responsive: true,
    ...props.config,
  }

  try {
    // Purge existing plot if any
    if (plotInstance.value) {
      Plotly.purge(plotInstance.value)
    }

    await Plotly.newPlot(plotContainer.value, cleanedData, plotLayout, plotConfig)
    plotInstance.value = plotContainer.value
  } catch (error) {
    console.error('Error rendering plot:', error)
  }
}

const resizePlot = async () => {
  if (plotInstance.value) {
    try {
      await Plotly.Plots.resize(plotInstance.value)
    } catch (error) {
      console.error('Error resizing plot:', error)
    }
  }
}

watch(
  () => [props.data, props.layout],
  () => {
    renderPlot()
  },
  { deep: true },
)

onMounted(() => {
  renderPlot()
  window.addEventListener('resize', resizePlot)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizePlot)
  if (plotInstance.value) {
    Plotly.purge(plotInstance.value)
  }
})
</script>
