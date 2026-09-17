<template>
  <div ref="containerRef" class="select-field">
    <span v-if="label" class="field-label">{{ label }}<em v-if="required"> *</em></span>
    <button
      ref="triggerRef"
      type="button"
      class="select-trigger"
      :class="{ 'select-trigger-open': isOpen, 'select-trigger-disabled': disabled }"
      :disabled="disabled"
      :id="id"
      :aria-label="ariaLabel || label || t('common.selectOption')"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      :aria-required="required"
      @click.stop="toggle"
      @keydown.down.prevent="openAndFocus"
      @keydown.up.prevent="openAndFocus"
    >
      <span class="select-value">
        <slot name="selected" :option="selectedOption">
          {{ selectedLabel }}
        </slot>
      </span>
      <span class="select-icon">
        <Icon name="chevronDown" size="sm" :class="{ 'is-open': isOpen }" />
      </span>
    </button>

    <span v-if="hint" class="field-hint">{{ hint }}</span>

    <Teleport to="body">
      <Transition name="select-dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="select-dropdown-portal"
          :class="instanceId"
          :style="dropdownStyle"
          role="listbox"
          @click.stop
          @mousedown.stop
          @keydown="onDropdownKeyDown"
        >
          <div v-if="isSearchable" class="select-search">
            <Icon name="search" size="sm" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="search"
              :placeholder="t('common.search')"
              :aria-label="t('common.search')"
              @click.stop
            />
          </div>

          <div ref="optionsListRef" class="select-options">
            <div
              v-for="(option, index) in filteredOptions"
              :key="`${String(option.value)}-${index}`"
              class="select-option"
              :class="{
                'select-option-selected': isSelected(option),
                'select-option-focused': focusedIndex === index,
                'select-option-disabled': option.disabled,
              }"
              role="option"
              :aria-selected="isSelected(option)"
              :aria-disabled="option.disabled || undefined"
              @mouseenter="focusedIndex = index"
              @click.stop="!option.disabled && selectOption(option)"
            >
              <slot name="option" :option="option" :selected="isSelected(option)">
                <span class="select-option-copy">
                  <span class="select-option-label">{{ option.label }}</span>
                  <small v-if="option.description">{{ option.description }}</small>
                </span>
                <Icon v-if="isSelected(option)" name="check" size="sm" class="select-option-check" :stroke-width="2" />
              </slot>
            </div>
            <div v-if="!filteredOptions.length" class="select-empty">{{ t('common.noData') }}</div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '../icons/Icon.vue'

export interface SelectOption {
  value: string | number | boolean | null
  label: string
  description?: string
  disabled?: boolean
  [key: string]: unknown
}

const props = withDefaults(defineProps<{
  modelValue?: string | number | boolean | null
  options: SelectOption[]
  label?: string
  hint?: string
  placeholder?: string
  required?: boolean
  disabled?: boolean
  searchable?: boolean | 'auto'
  id?: string
  ariaLabel?: string
}>(), {
  modelValue: '',
  required: false,
  disabled: false,
  searchable: 'auto',
})

const emit = defineEmits<{ 'update:modelValue': [value: string | number | boolean | null] }>()
const { t } = useI18n()
const instanceId = `select-${Math.random().toString(36).slice(2, 9)}`
const isOpen = ref(false)
const searchQuery = ref('')
const focusedIndex = ref(-1)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const optionsListRef = ref<HTMLElement | null>(null)
const triggerRect = ref<DOMRect | null>(null)
const dropdownPosition = ref<'top' | 'bottom'>('bottom')

const isSearchable = computed(() => props.searchable === true || (props.searchable === 'auto' && props.options.length > 7))
const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue) || null)
const selectedLabel = computed(() => selectedOption.value?.label || props.placeholder || t('common.selectOption'))
const filteredOptions = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return props.options
  return props.options.filter((option) => `${option.label} ${option.description || ''}`.toLowerCase().includes(query))
})
const dropdownStyle = computed<Record<string, string>>(() => {
  const rect = triggerRect.value
  if (!rect) return {}
  const edge = 8
  const right = Math.max(edge, window.innerWidth - edge)
  const left = Math.min(Math.max(edge, rect.left), right)
  const width = Math.min(Math.max(200, rect.width), Math.max(0, right - left))
  const style: Record<string, string> = {
    position: 'fixed',
    left: `${left}px`,
    minWidth: `${width}px`,
    maxWidth: `${Math.max(0, right - left)}px`,
    zIndex: '100000020',
  }
  if (dropdownPosition.value === 'top') style.bottom = `${window.innerHeight - rect.top + 5}px`
  else style.top = `${rect.bottom + 5}px`
  return style
})

const updatePosition = () => {
  if (!containerRef.value) return
  triggerRect.value = containerRef.value.getBoundingClientRect()
  nextTick(() => {
    if (!dropdownRef.value || !triggerRect.value) return
    const above = triggerRect.value.top
    const below = window.innerHeight - triggerRect.value.bottom
    dropdownPosition.value = below < Math.min(260, dropdownRef.value.offsetHeight) && above > below ? 'top' : 'bottom'
  })
}

const setInitialFocus = () => {
  const selected = filteredOptions.value.findIndex((option) => option.value === props.modelValue)
  focusedIndex.value = selected >= 0 ? selected : filteredOptions.value.findIndex((option) => !option.disabled)
}

const open = () => { if (!props.disabled) isOpen.value = true }
const toggle = () => { if (isOpen.value) isOpen.value = false; else open() }
const openAndFocus = () => { open() }

const selectOption = (option: SelectOption) => {
  emit('update:modelValue', option.value)
  isOpen.value = false
  triggerRef.value?.focus()
}

const moveFocus = (delta: number) => {
  const options = filteredOptions.value
  if (!options.length) return
  let index = focusedIndex.value
  for (let count = 0; count < options.length; count += 1) {
    index = (index + delta + options.length) % options.length
    if (!options[index].disabled) {
      focusedIndex.value = index
      const element = optionsListRef.value?.children[index] as HTMLElement | undefined
      element?.scrollIntoView({ block: 'nearest' })
      return
    }
  }
}

const onDropdownKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowDown') { event.preventDefault(); moveFocus(1) }
  if (event.key === 'ArrowUp') { event.preventDefault(); moveFocus(-1) }
  if (event.key === 'Enter' && focusedIndex.value >= 0) {
    event.preventDefault()
    const option = filteredOptions.value[focusedIndex.value]
    if (option && !option.disabled) selectOption(option)
  }
  if (event.key === 'Escape') { event.preventDefault(); isOpen.value = false; triggerRef.value?.focus() }
  if (event.key === 'Tab') isOpen.value = false
}

const isSelected = (option: SelectOption) => option.value === props.modelValue
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (isOpen.value && !containerRef.value?.contains(target) && !target.closest(`.${instanceId}`)) isOpen.value = false
}

watch(isOpen, (value) => {
  if (value) {
    setInitialFocus()
    updatePosition()
    if (isSearchable.value) nextTick(() => searchInputRef.value?.focus())
    window.addEventListener('resize', updatePosition)
    window.addEventListener('scroll', updatePosition, true)
  } else {
    searchQuery.value = ''
    focusedIndex.value = -1
    window.removeEventListener('resize', updatePosition)
    window.removeEventListener('scroll', updatePosition, true)
  }
})

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updatePosition)
  window.removeEventListener('scroll', updatePosition, true)
})
</script>
