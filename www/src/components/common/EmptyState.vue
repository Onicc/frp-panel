<template>
  <div class="empty-state">
    <!-- Broken-link / disconnected illustration -->
    <svg width="56" height="56" viewBox="0 0 56 56" fill="none" aria-hidden="true">
      <rect x="1" y="1" width="54" height="54" rx="14" stroke="currentColor" stroke-width="1.5" />
      <path d="M20 28h5m6 0h5M22 24l-4 4 4 4M34 24l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
    <strong>{{ title }}</strong>
    <p v-if="message">{{ message }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ title: string; message?: string }>(), { message: '' })
</script>
