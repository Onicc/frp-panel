<template>
  <aside class="sidebar" :class="{ open: mobileOpen }" aria-label="Primary navigation">
    <div class="sidebar-header">
      <RouterLink to="/" class="brand" @click="emit('close')">
        <span class="brand-logo" aria-hidden="true"><img src="/frppanel-logo.svg" alt="" /></span>
        <span class="brand-name">
          <strong>{{ t('brand.name') }}</strong>
          <small>{{ t('brand.description') }}</small>
        </span>
      </RouterLink>
    </div>

    <nav class="sidebar-nav">
      <p class="sidebar-section-title">{{ t('nav.master') }}</p>
      <RouterLink to="/" class="sidebar-link" :class="{ active: isActive('/') }" @click="emit('close')">
        <Icon name="home" size="sm" />
        <span>{{ t('nav.overview') }}</span>
      </RouterLink>
      <RouterLink to="/clients" class="sidebar-link" :class="{ active: isActive('/clients') }" @click="emit('close')">
        <Icon name="users" size="sm" />
        <span>{{ t('nav.clients') }}</span>
      </RouterLink>
      <RouterLink to="/servers" class="sidebar-link" :class="{ active: isActive('/servers') }" @click="emit('close')">
        <Icon name="server" size="sm" />
        <span>{{ t('nav.servers') }}</span>
      </RouterLink>
      <RouterLink to="/tunnels" class="sidebar-link" :class="{ active: isActive('/tunnels') }" @click="emit('close')">
        <Icon name="arrowsUpDown" size="sm" />
        <span>{{ t('nav.tunnels') }}</span>
      </RouterLink>
    </nav>

    <div class="sidebar-footer">
      <div class="sidebar-status"><span class="status-pulse"></span><span>FRP {{ version }}</span></div>
      <span class="sidebar-footer-caption">{{ t('nav.masterHint') }}</span>
    </div>
  </aside>

  <div v-if="mobileOpen" class="mobile-overlay" aria-hidden="true" @click="emit('close')"></div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '../icons/Icon.vue'

defineProps<{ mobileOpen?: boolean }>()
const emit = defineEmits<{ close: [] }>()
const route = useRoute()
const { t } = useI18n()
const version = import.meta.env.VITE_APP_VERSION || '2'
const isActive = (path: string) => path === '/' ? route.path === '/' : route.path.startsWith(path)
</script>
