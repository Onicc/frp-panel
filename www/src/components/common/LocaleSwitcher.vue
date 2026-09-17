<template>
  <div ref="dropdownRef" class="locale-switcher">
    <button
      type="button"
      class="locale-trigger"
      :disabled="switching"
      :aria-expanded="isOpen"
      aria-haspopup="menu"
      :title="currentLocale.name"
      @click.stop="isOpen = !isOpen"
    >
      <span class="locale-flag" aria-hidden="true">{{ currentLocale.flag }}</span>
      <span class="locale-code">{{ currentLocale.code === 'zh-CN' ? '中' : 'EN' }}</span>
      <Icon name="chevronDown" size="xs" :class="{ 'is-open': isOpen }" />
    </button>

    <Transition name="dropdown">
      <div v-if="isOpen" class="locale-menu" role="menu">
        <button
          v-for="option in availableLocales"
          :key="option.code"
          type="button"
          class="locale-option"
          :class="{ active: option.code === currentLocaleCode }"
          role="menuitem"
          :disabled="switching"
          @click="selectLocale(option.code)"
        >
          <span class="locale-flag" aria-hidden="true">{{ option.flag }}</span>
          <span>{{ option.name }}</span>
          <Icon v-if="option.code === currentLocaleCode" name="check" size="sm" class="locale-check" />
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import Icon from '../icons/Icon.vue'
import { availableLocales, locale, setLocale } from '../../i18n'

const isOpen = ref(false)
const switching = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((item) => item.code === locale.value) || availableLocales[0])

const selectLocale = async (code: string) => {
  if (switching.value || code === currentLocaleCode.value) {
    isOpen.value = false
    return
  }
  switching.value = true
  try {
    await setLocale(code)
    isOpen.value = false
  } finally {
    switching.value = false
  }
}

const onDocumentClick = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) isOpen.value = false
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocumentClick))
</script>
