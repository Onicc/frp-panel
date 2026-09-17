<template>
  <section>
    <div class="page-header">
      <div>
        <h1>{{ t('account.title') }}</h1>
        <p>{{ t('account.passwordHint') }}</p>
      </div>
    </div>

    <div class="settings-grid">
      <!-- Profile card -->
      <article class="panel profile-card">
        <div class="profile-avatar">{{ initials }}</div>
        <h2>{{ auth.user?.username }}</h2>
        <p class="email">{{ auth.user?.email }}</p>
        <dl>
          <div>
            <dt>{{ t('account.role') }}</dt>
            <dd><span class="role-badge">{{ auth.user?.role }}</span></dd>
          </div>
          <div>
            <dt>{{ t('auth.username') }}</dt>
            <dd>{{ auth.user?.username }}</dd>
          </div>
          <div>
            <dt>{{ t('auth.email') }}</dt>
            <dd>{{ auth.user?.email }}</dd>
          </div>
        </dl>
      </article>

      <!-- Change password card -->
      <article class="panel password-card">
        <h2>{{ t('account.password') }}</h2>
        <p>{{ t('account.passwordHint') }}</p>
        <form @submit.prevent="save">
          <Input v-model="form.current" :label="t('auth.currentPassword')" type="password" autocomplete="current-password" required />
          <Input v-model="form.next" :label="t('auth.newPassword')" type="password" autocomplete="new-password" required />
          <Input v-model="form.confirm" :label="t('auth.confirmPassword')" type="password" autocomplete="new-password" required />
          <p v-if="error" class="error" role="alert">{{ error }}</p>
          <p v-if="saved" class="success-alert" role="status">{{ t('common.success') }}</p>
          <button class="button primary" type="submit" :disabled="busy">
            {{ busy ? t('auth.loading') : t('account.save') }}
          </button>
        </form>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Input } from '../components/common'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { api, APIError } from '../api'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const toast = useToastStore()

const form  = reactive({ current: '', next: '', confirm: '' })
const error = ref('')
const saved = ref(false)
const busy  = ref(false)

const initials = computed(() => auth.user?.username.slice(0, 2).toUpperCase() || 'FP')

const save = async () => {
  error.value = ''
  saved.value = false
  if (form.next.length < 12 || form.next !== form.confirm) {
    error.value = t('account.passwordHint')
    return
  }
  busy.value = true
  try {
    await api.changePassword(form.current, form.next)
    saved.value = true
    toast.show(t('common.success'))
    form.current = ''
    form.next    = ''
    form.confirm = ''
    await auth.logout()
    await router.replace('/login')
  } catch (e) {
    error.value = e instanceof APIError ? e.message : t('common.error')
  } finally {
    busy.value = false
  }
}
</script>
