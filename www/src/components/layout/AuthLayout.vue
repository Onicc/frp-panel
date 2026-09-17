<template>
  <div class="auth-page">
    <div class="auth-decoration auth-decoration-top" aria-hidden="true"></div>
    <div class="auth-decoration auth-decoration-bottom" aria-hidden="true"></div>
    <div class="auth-grid" aria-hidden="true"></div>

    <div class="auth-content">
      <RouterLink to="/login" class="auth-brand">
        <img class="auth-brand-logo" src="/frppanel-logo.svg" alt="" />
        <strong>{{ t('brand.name') }}</strong>
        <small>{{ t('brand.description') }}</small>
      </RouterLink>

      <section class="auth-card"><slot /></section>
      <div class="auth-footer"><slot name="footer" /></div>
      <p class="auth-copyright">© {{ year }} {{ t('brand.name') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const year = computed(() => new Date().getFullYear())
</script>
