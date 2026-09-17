<template>
  <aside class="sidebar" :class="{open: mobileOpen}" aria-label="Primary navigation">
    <RouterLink to="/" class="brand" tabindex="0">
      <div class="brand-logo" aria-hidden="true">
        <Icon name="topology" size="sm" />
      </div>
      <div class="brand-name">
        <strong>frp-panel</strong>
        <small>{{ t('nav.masterHint') }}</small>
      </div>
    </RouterLink>

    <nav>
      <RouterLink to="/" :class="{active: isActive('/')}">
        <span class="nav-icon" aria-hidden="true"><Icon name="home" size="sm" /></span>
        {{ t('nav.overview') }}
      </RouterLink>
      <RouterLink to="/clients" :class="{active: isActive('/clients')}">
        <span class="nav-icon" aria-hidden="true"><Icon name="users" size="sm" /></span>
        {{ t('nav.clients') }}
      </RouterLink>
      <RouterLink to="/servers" :class="{active: isActive('/servers')}">
        <span class="nav-icon" aria-hidden="true"><Icon name="server" size="sm" /></span>
        {{ t('nav.servers') }}
      </RouterLink>
      <RouterLink to="/tunnels" :class="{active: isActive('/tunnels')}">
        <span class="nav-icon" aria-hidden="true"><Icon name="arrowsUpDown" size="sm" /></span>
        {{ t('nav.tunnels') }}
      </RouterLink>
    </nav>

    <div class="sidebar-footnote">
      <Icon name="shield" size="xs" aria-hidden="true" />
      <small>v{{ version }}</small>
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

const isActive = (path: string) =>
  path === '/' ? route.path === '/' : route.path.startsWith(path)
</script>
