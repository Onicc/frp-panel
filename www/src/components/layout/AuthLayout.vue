<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-toolbar">
        <RouterLink to="/" class="brand login-brand">
          <div class="brand-logo" aria-hidden="true">
            <Icon name="topology" size="sm" />
          </div>
          <div class="brand-name">
            <strong>frp-panel</strong>
          </div>
        </RouterLink>
        <LocaleSwitcher />
      </div>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { LocaleSwitcher } from '../common'
import Icon from '../icons/Icon.vue'
</script>
