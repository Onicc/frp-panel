<template>
  <div ref="dropdownRef" class="theme-switcher">
    <button
      type="button"
      class="theme-trigger"
      :aria-expanded="isOpen"
      aria-haspopup="menu"
      :title="t('theme.title')"
      @click.stop="isOpen = !isOpen"
    >
      <Icon :name="resolved === 'dark' ? 'moon' : 'sun'" size="sm" />
      <span class="theme-trigger-label">{{ currentOption.label }}</span>
      <Icon name="chevronDown" size="xs" :class="{ 'is-open': isOpen }" />
    </button>

    <Transition name="dropdown">
      <div v-if="isOpen" class="theme-menu" role="menu">
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="theme-option"
          :class="{ active: option.value === preference }"
          role="menuitem"
          @click="selectTheme(option.value)"
        >
          <Icon :name="option.icon" size="sm" />
          <span class="theme-option-copy">
            <strong>{{ option.label }}</strong>
            <small>{{ option.hint }}</small>
          </span>
          <Icon v-if="option.value === preference" name="check" size="sm" class="theme-check" />
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '../icons/Icon.vue'
import { useTheme, type ThemePreference } from '../../theme'

const { t } = useI18n()
const { preference, resolved, setTheme } = useTheme()
const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const options = computed(() => [
  { value: 'system' as const, icon: 'cog', label: t('theme.system'), hint: t('theme.systemHint') },
  { value: 'light' as const, icon: 'sun', label: t('theme.light'), hint: t('theme.lightHint') },
  { value: 'dark' as const, icon: 'moon', label: t('theme.dark'), hint: t('theme.darkHint') },
])
const currentOption = computed(() => options.value.find((item) => item.value === preference.value) || options.value[0])

const selectTheme = (value: ThemePreference) => {
  setTheme(value)
  isOpen.value = false
}

const onDocumentClick = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) isOpen.value = false
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocumentClick))
</script>
