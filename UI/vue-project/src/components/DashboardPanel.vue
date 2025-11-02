<template>
  <v-card :color="color" :variant="variant" class="mx-auto">
    <v-card-item>
      <div>
        <div v-if="overline" class="text-overline mb-1">{{ overline }}</div>
        <div v-if="title" class="text-h6 mb-1">{{ title }}</div>
        <div v-if="subtitle" class="text-caption">{{ subtitle }}</div>
      </div>
    </v-card-item>

    <v-card-text v-if="$slots.default">
      <slot></slot>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    overline?: string
    title?: string
    subtitle?: string
    color?: string
    variant?: 'outlined' | 'text' | 'flat' | 'elevated' | 'plain' | 'tonal'
  }>(),
  {
    color: 'primary',
    variant: 'outlined',
  },
)
</script>
