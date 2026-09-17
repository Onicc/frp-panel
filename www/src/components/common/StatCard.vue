<template>
  <article class="stat-card">
    <div class="stat-heading">
      <span>{{ label }}</span>
      <Icon :name="icon" size="sm" class="stat-icon" />
    </div>
    <strong class="stat-number">{{ value }}</strong>
    <div v-if="hint || trend" style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;">
      <small v-if="hint" class="stat-hint">{{ hint }}</small>
      <span v-if="trend" class="stat-trend" :class="trendClass">{{ trend }}</span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '../icons/Icon.vue'

const props = withDefaults(defineProps<{
  label: string
  value: string | number
  hint?: string
  icon?: string
  trend?: string
}>(), { icon: 'chart' })

const trendClass = computed(() => {
  if (!props.trend) return ''
  return props.trend.startsWith('-') ? 'down' : 'up'
})
</script>
