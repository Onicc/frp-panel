<template>
  <header class="topbar">
    <div>
      <span class="topbar-title">{{ title }}</span>
    </div>
    <div class="topbar-actions">
      <LocaleSwitcher />
      <div v-if="auth.user" class="user-menu" ref="menuRef">
        <button type="button" class="account-link" :aria-expanded="open" aria-haspopup="menu" @click="open = !open">
          <span class="account-dot" aria-hidden="true">{{ initials }}</span>
          <span class="hide-mobile">{{ auth.user.username }}</span>
          <Icon name="chevronDown" size="xs" />
        </button>
        <div v-if="open" class="user-dropdown" role="menu">
          <div class="dropdown-user">
            <strong>{{ auth.user.username }}</strong>
            <small>{{ auth.user.email }}</small>
          </div>
          <RouterLink to="/account" role="menuitem" @click="open = false">
            <Icon name="user" size="sm" />{{ t('account.title') }}
          </RouterLink>
          <button type="button" role="menuitem" @click="handleLogout">
            <Icon name="logout" size="sm" />{{ t('auth.logout') }}
          </button>
        </div>
      </div>
      <button class="menu-button" type="button" :aria-label="t('nav.openMenu')" @click="emit('menu')">
        <Icon name="menu" />
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { LocaleSwitcher } from '../common'
import Icon from '../icons/Icon.vue'

defineProps<{ title: string }>()
const emit = defineEmits<{ menu: [] }>()

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()
const open = ref(false)
const menuRef = ref<HTMLElement>()
const initials = computed(() => auth.user?.username.slice(0, 2).toUpperCase() || 'FP')

const close = (e: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) open.value = false
}
const handleLogout = async () => {
  open.value = false
  await auth.logout()
  await router.push('/login')
}

onMounted(() => document.addEventListener('click', close))
onBeforeUnmount(() => document.removeEventListener('click', close))
</script>
