<template>
  <span class="status-badge" :class="statusClass">
    <span class="dot" aria-hidden="true"></span>
    {{ displayLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ status: string; label?: string }>()
const { t } = useI18n()

const statusClass = computed(() => props.status.toLowerCase())
const displayLabel = computed(() =>
  props.label || t(`common.${statusClass.value}`, statusClass.value)
)
</script>
