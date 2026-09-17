<template>
  <div class="pagination">
    <span class="page-info">{{ from }}–{{ to }} / {{ total }}</span>
    <div style="display:flex;gap:6px;">
      <button class="button secondary" type="button" :disabled="page <= 1" @click="emit('update:page', page - 1)">
        {{ t('common.previous') }}
      </button>
      <button class="button secondary" type="button" :disabled="page >= pages" @click="emit('update:page', page + 1)">
        {{ t('common.next') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ page: number; pageSize: number; total: number }>()
const emit = defineEmits<{ 'update:page': [value: number] }>()
const { t } = useI18n()

const pages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const from  = computed(() => props.total ? (props.page - 1) * props.pageSize + 1 : 0)
const to    = computed(() => Math.min(props.total, props.page * props.pageSize))
</script>
