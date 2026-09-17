import { computed, ref } from 'vue'

export type ThemePreference = 'system' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

const storageKey = 'frp-panel.theme'
const preference = ref<ThemePreference>(readPreference())
const resolved = ref<ResolvedTheme>('light')
let mediaQuery: MediaQueryList | undefined

function readPreference(): ThemePreference {
  const saved = localStorage.getItem(storageKey)
  return saved === 'light' || saved === 'dark' ? saved : 'system'
}

function resolveTheme(value: ThemePreference): ResolvedTheme {
  if (value !== 'system') return value
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(value: ThemePreference = preference.value) {
  const next = resolveTheme(value)
  preference.value = value
  resolved.value = next
  document.documentElement.classList.toggle('dark', next === 'dark')
  document.documentElement.dataset.theme = next
  document.documentElement.style.colorScheme = next
}

export function initializeTheme() {
  applyTheme(preference.value)
  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', () => {
    if (preference.value === 'system') applyTheme('system')
  })
}

export function setTheme(value: ThemePreference) {
  localStorage.setItem(storageKey, value)
  applyTheme(value)
}

export function useTheme() {
  return {
    preference,
    resolved: computed(() => resolved.value),
    setTheme,
  }
}
