<template><label><span v-if="label">{{ label }}<em v-if="required"> *</em></span><input v-bind="$attrs" :required="required" :value="modelValue" @input="onInput"></label></template>
<script setup lang="ts">
defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ modelValue?: string | number; label?: string; required?: boolean; modelModifiers?: { number?: boolean; trim?: boolean } }>(), { modelValue: '', required: false })
const emit = defineEmits<{ 'update:modelValue': [value: string | number] }>()
const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  let value: string | number = target.value
  if (props.modelModifiers?.trim) value = value.trim()
  if (props.modelModifiers?.number && value !== '') value = Number(value)
  emit('update:modelValue', value)
}
</script>
