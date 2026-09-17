<template>
  <label>
    <span v-if="label">{{ label }}<em v-if="required"> *</em></span>
    <input
      v-bind="$attrs"
      :required="required"
      :value="modelValue"
      :class="{'error-field': !!error}"
      @input="onInput"
    />
    <span v-if="hint && !error" class="field-hint">{{ hint }}</span>
    <span v-if="error" class="error-text" role="alert">
      <Icon name="exclamationTriangle" size="xs" />{{ error }}
    </span>
  </label>
</template>

<script setup lang="ts">
import Icon from '../icons/Icon.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  modelValue?: string | number
  label?: string
  required?: boolean
  hint?: string
  error?: string
  modelModifiers?: { number?: boolean; trim?: boolean }
}>(), { modelValue: '', required: false })

const emit = defineEmits<{ 'update:modelValue': [value: string | number] }>()

const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  let value: string | number = target.value
  if (props.modelModifiers?.trim) value = value.trim()
  if (props.modelModifiers?.number && value !== '') value = Number(value)
  emit('update:modelValue', value)
}
</script>
