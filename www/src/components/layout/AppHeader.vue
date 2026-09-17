<template>
  <header class="topbar">
    <div class="header-left">
      <button class="menu-button" type="button" :aria-label="t('nav.openMenu')" @click="emit('menu')">
        <Icon name="menu" size="md" />
      </button>
      <div class="header-title-block">
        <h1>{{ title }}</h1>
        <span class="header-kicker">frp-panel</span>
      </div>
    </div>

    <div class="topbar-actions">
      <ThemeSwitcher />
      <LocaleSwitcher />
      <div v-if="auth.user" ref="menuRef" class="user-menu">
        <button
          type="button"
          class="account-link"
          :aria-expanded="open"
          aria-haspopup="menu"
          @click="open = !open"
        >
          <span class="account-avatar" aria-hidden="true">{{ initials }}</span>
          <span class="account-copy">
            <strong>{{ auth.user.username }}</strong>
            <small>{{ auth.user.role }}</small>
          </span>
          <Icon name="chevronDown" size="xs" :class="{ 'is-open': open }" />
        </button>

        <Transition name="dropdown">
          <div v-if="open" class="user-dropdown dropdown" role="menu">
            <div class="dropdown-user">
              <strong>{{ auth.user.username }}</strong>
              <small>{{ auth.user.email }}</small>
            </div>
            <RouterLink to="/account" role="menuitem" class="dropdown-item" @click="open = false">
              <Icon name="user" size="sm" />{{ t('account.title') }}
            </RouterLink>
            <button type="button" role="menuitem" class="dropdown-item danger" @click="handleLogout">
              <Icon name="login" size="sm" class="logout-icon" />{{ t('auth.logout') }}
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { LocaleSwitcher, ThemeSwitcher } from '../common'
import Icon from '../icons/Icon.vue'

defineProps<{ title: string }>()
const emit = defineEmits<{ menu: [] }>()
const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()
const open = ref(false)
const menuRef = ref<HTMLElement | null>(null)
const initials = computed(() => auth.user?.username.slice(0, 2).toUpperCase() || 'FP')

const close = (event: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) open.value = false
}
const handleLogout = async () => {
  open.value = false
  await auth.logout()
  await router.push('/login')
}

onMounted(() => document.addEventListener('click', close))
onBeforeUnmount(() => document.removeEventListener('click', close))
</script>
