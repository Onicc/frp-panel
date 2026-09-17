<template>
  <aside class="sidebar" :class="{open: mobileOpen}">
    <div class="brand"><div class="brand-mark">FP</div><div><strong>frp-panel</strong><small>Master console</small></div></div>
    <nav aria-label="Primary">
      <RouterLink to="/" :class="{active:isActive('/')}"><span class="nav-icon"><Icon name="home" size="sm" /></span>{{ t('nav.overview') }}</RouterLink>
      <RouterLink to="/clients" :class="{active:isActive('/clients')}"><span class="nav-icon"><Icon name="users" size="sm" /></span>{{ t('nav.clients') }}</RouterLink>
      <RouterLink to="/servers" :class="{active:isActive('/servers')}"><span class="nav-icon"><Icon name="server" size="sm" /></span>{{ t('nav.servers') }}</RouterLink>
      <RouterLink to="/tunnels" :class="{active:isActive('/tunnels')}"><span class="nav-icon"><Icon name="arrowsUpDown" size="sm" /></span>{{ t('nav.tunnels') }}</RouterLink>
    </nav>
    <div class="sidebar-footnote"><Icon name="shield" size="sm" /><small>{{ t('nav.masterHint') }}</small></div>
  </aside>
  <div v-if="mobileOpen" class="mobile-overlay" @click="emit('close')"></div>
</template>
<script setup lang="ts">
import { useRoute } from 'vue-router'; import { useI18n } from 'vue-i18n'; import Icon from '../icons/Icon.vue'
defineProps<{mobileOpen?:boolean}>(); const emit=defineEmits<{close:[]}>(); const route=useRoute(); const {t}=useI18n(); const isActive=(path:string)=>path==='/'?route.path==='/':route.path.startsWith(path)
</script>
