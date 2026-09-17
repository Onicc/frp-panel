<template>
  <label>
    <span v-if="label">{{ label }}</span>
    <select v-bind="$attrs" :value="modelValue" @change="onChange">
      <slot />
    </select>
    <span v-if="hint" class="field-hint">{{ hint }}</span>
  </label>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{
  modelValue?: string | number
  label?: string
  hint?: string
}>(), { modelValue: '' })

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const onChange = (e: Event) => {
  emit('update:modelValue', (e.target as HTMLSelectElement).value)
}
</script>
