<template>
  <div class="empty-state">
    <span class="empty-state-icon" aria-hidden="true"><Icon name="inbox" size="xl" /></span>
    <strong>{{ title }}</strong>
    <p v-if="message">{{ message }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
import Icon from '../icons/Icon.vue'

withDefaults(defineProps<{ title: string; message?: string }>(), { message: '' })
</script>
