<template>
  <svg :class="sizeClass" viewBox="0 0 24 24" fill="none" stroke="currentColor" :stroke-width="strokeWidth" aria-hidden="true">
    <path stroke-linecap="round" stroke-linejoin="round" :d="icons[name] || icons.info" />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ name: keyof typeof icons | string; size?: 'xs'|'sm'|'md'|'lg'|'xl'; strokeWidth?: number }>(), { size: 'md', strokeWidth: 1.7 })
const icons: Record<string, string> = {
  menu: 'M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5',
  home: 'M2.25 12l9.75-9 9.75 9M4.5 10.5v9.75h5.25v-6h4.5v6h5.25V10.5',
  users: 'M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zm8-3a3 3 0 110 6m4 7v-2a4 4 0 00-3-3.87',
  server: 'M4 5h16v5H4zM4 14h16v5H4zM7 7.5h.01M7 16.5h.01M11 7.5h6M11 16.5h6',
  arrowsUpDown: 'M8 3v18m0 0l-4-4m4 4l4-4M16 21V3m0 0l-4 4m4-4l4 4',
  arrowRight: 'M5 12h14m-6-6l6 6-6 6',
  search: 'M21 21l-4.35-4.35m2.1-5.4a7.5 7.5 0 11-15 0 7.5 7.5 0 0115 0z',
  inbox: 'M4 4h16v13H4zM4 13h4l2 3h4l2-3h4',
  chart: 'M4 19V9m5 10V5m5 14v-7m5 7V3',
  shield: 'M12 3l8 3v5c0 5-3.3 8.4-8 10-4.7-1.6-8-5-8-10V6l8-3zM9 12l2 2 4-4',
  user: 'M20 21a8 8 0 00-16 0M12 13a4 4 0 100-8 4 4 0 000 8z',
  logout: 'M10 17l5-5-5-5m5 5H3m11-9h3a2 2 0 012 2v14a2 2 0 01-2 2h-3',
  login: 'M15 3h4a2 2 0 012 2v14a2 2 0 01-2 2h-4M10 17l5-5-5-5m5 5H3',
  globe: 'M12 21a9 9 0 100-18 9 9 0 000 18zm0 0c2.5-2.4 3.8-5.4 3.8-9S14.5 5.4 12 3m0 18c-2.5-2.4-3.8-5.4-3.8-9S9.5 5.4 12 3M3 12h18',
  edit: 'M12 20h9M16.5 3.5a2.12 2.12 0 013 3L8 18l-4 1 1-4L16.5 3.5z',
  trash: 'M3 6h18M8 6V4h8v2m-9 0l1 15h8l1-15M10 11v6m4-6v6',
  plus: 'M12 5v14M5 12h14',
  copy: 'M9 9h10v10H9zM5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1',
  key: 'M21 2l-2 2m-7.6 7.6a5 5 0 117.1 0L13 17h-3v3H7v2H3v-4.6l6.4-6.4',
  x: 'M6 6l12 12M18 6L6 18',
  check: 'M5 12l4 4L19 6',
  refresh: 'M20 11a8.1 8.1 0 00-14.9-4L3 10m0 0V4m0 6h6M4 13a8.1 8.1 0 0014.9 4L21 14m0 0v6m0-6h-6',
  terminal: 'M4 17l6-5-6-5m8 10h8',
  chevronDown: 'M6 9l6 6 6-6',
  exclamationTriangle: 'M12 9v4m0 4h.01M10.3 3.5L2.8 17a2 2 0 001.7 3h15a2 2 0 001.7-3L13.7 3.5a2 2 0 00-3.4 0z',
  info: 'M12 8h.01M11 12h1v4h1m-1 5a9 9 0 110-18 9 9 0 010 18z',
  lock: 'M6 10V7a6 6 0 0112 0v3m-13 0h14v10H5V10z'
} as const
const sizeClass = computed(() => ({ xs: 'h-3 w-3', sm: 'h-4 w-4', md: 'h-5 w-5', lg: 'h-6 w-6', xl: 'h-8 w-8' }[props.size]))
</script>
