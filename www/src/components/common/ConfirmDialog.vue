<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="emit('cancel')">
    <p class="dialog-copy">{{ message }}</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <template #footer><button class="button secondary" type="button" @click="emit('cancel')">{{ cancelLabel }}</button><button class="button danger" type="button" :disabled="busy" @click="emit('confirm')">{{ busy ? busyLabel : confirmLabel }}</button></template>
  </BaseDialog>
</template>
<script setup lang="ts">
import BaseDialog from './BaseDialog.vue'
withDefaults(defineProps<{ show: boolean; title: string; message: string; error?: string; confirmLabel?: string; cancelLabel?: string; busy?: boolean; busyLabel?: string }>(), { error: '', confirmLabel: 'Confirm', cancelLabel: 'Cancel', busy: false, busyLabel: 'Working…' })
const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>
