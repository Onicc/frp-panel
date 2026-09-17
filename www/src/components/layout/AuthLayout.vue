<template>
  <div class="auth-page">
    <div class="auth-decoration auth-decoration-top" aria-hidden="true"></div>
    <div class="auth-decoration auth-decoration-bottom" aria-hidden="true"></div>
    <div class="auth-grid" aria-hidden="true"></div>

    <div class="auth-content">
      <RouterLink to="/login" class="auth-brand">
        <span class="auth-brand-icon" aria-hidden="true"><Icon name="server" size="md" /></span>
        <span>
          <strong>frp-panel</strong>
          <small>{{ t('nav.masterHint') }}</small>
        </span>
      </RouterLink>

      <section class="auth-card"><slot /></section>
      <div class="auth-footer"><slot name="footer" /></div>
      <p class="auth-copyright">© {{ year }} frp-panel</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '../icons/Icon.vue'

const { t } = useI18n()
const year = computed(() => new Date().getFullYear())
</script>
