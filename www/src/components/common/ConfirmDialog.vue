<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="emit('cancel')">
    <div v-if="warning" class="confirm-warning">
      <Icon name="exclamationTriangle" size="md" />
      <p>{{ warning }}</p>
    </div>
    <p v-if="message" class="dialog-copy confirm-message">{{ message }}</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <template #footer><button class="button secondary" type="button" @click="emit('cancel')">{{ cancelLabel }}</button><button class="button danger" type="button" :disabled="busy" @click="emit('confirm')">{{ busy ? busyLabel : confirmLabel }}</button></template>
  </BaseDialog>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'
import Icon from '../icons/Icon.vue'

const props = withDefaults(defineProps<{
  show: boolean
  title: string
  message: string
  warning?: string
  error?: string
  confirmLabel?: string
  cancelLabel?: string
  busy?: boolean
  busyLabel?: string
}>(), { error: '', warning: '', busy: false })

const { t } = useI18n()
const confirmLabel = computed(() => props.confirmLabel || t('common.confirm'))
const cancelLabel = computed(() => props.cancelLabel || t('common.cancel'))
const busyLabel = computed(() => props.busyLabel || t('common.loading'))
const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>
