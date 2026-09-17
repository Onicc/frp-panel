<template>
  <div class="app-shell">
    <div class="mesh-background" aria-hidden="true"></div>
    <AppSidebar :mobile-open="mobileOpen" @close="mobileOpen = false" />
    <div class="app-main">
      <AppHeader
        :title="String(route.meta.title ? t(String(route.meta.title)) : t('nav.overview'))"
        @menu="mobileOpen = true"
      />
      <main class="content"><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const mobileOpen = ref(false)
const route = useRoute()
const { t } = useI18n()
</script>
