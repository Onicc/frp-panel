<template>
  <AuthLayout>
    <h1>{{ registering ? t('auth.register') : t('auth.signIn') }}</h1>
    <p>{{ registering ? t('auth.registerHint') : t('auth.signInHint') }}</p>

    <form @submit.prevent="submit" autocomplete="on" novalidate>
      <Input v-model="form.username" :label="t('auth.username')" icon="user" name="username" autocomplete="username" autofocus required />
      <Input v-if="registering" v-model="form.email" :label="t('auth.email')" name="email" type="email" autocomplete="email" required />
      <Input v-model="form.password" :label="t('auth.password')" icon="lock" name="password" type="password" :autocomplete="registering ? 'new-password' : 'current-password'" required />
      <Input v-if="registering" v-model="form.confirm" :label="t('auth.confirmPassword')" name="confirm-password" type="password" autocomplete="new-password" required />
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="button primary" type="submit" :disabled="busy" style="width:100%">
        <Icon v-if="!busy" name="login" size="sm" />
        {{ busy ? t('auth.loading') : (registering ? t('auth.create') : t('auth.submit')) }}
      </button>
    </form>

    <div v-if="registering" class="setup-warning">
      {{ t('auth.registerHint') }}
    </div>

    <button v-if="canRegister || registering" class="text-button login-mode" type="button" @click="switchMode">
      {{ registering ? t('auth.switchLogin') : t('auth.switchRegister') }}
    </button>
  </AuthLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Input } from '../components/common'
import Icon from '../components/icons/Icon.vue'
import AuthLayout from '../components/layout/AuthLayout.vue'
import { useAuthStore } from '../stores/auth'
import { APIError, api } from '../api'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const registering = ref(false)
const canRegister  = ref(false)
const busy  = ref(false)
const error = ref('')
const form  = reactive({ username: '', email: '', password: '', confirm: '' })

onMounted(async () => {
  try {
    const status = await api.bootstrapStatus()
    canRegister.value = status.registrationEnabled && status.canCreateOwner
  } catch {
    canRegister.value = false
  }
})

const switchMode = () => {
  if (!canRegister.value && !registering.value) return
  registering.value = !registering.value
  error.value = ''
  form.password = ''
  form.confirm  = ''
}

const submit = async () => {
  error.value = ''
  if (!form.username || !form.password || (registering.value && (!form.email || !form.confirm))) {
    error.value = t('problems.required')
    return
  }
  if (registering.value && form.password !== form.confirm) {
    error.value = t('auth.confirmPassword')
    return
  }
  busy.value = true
  try {
    if (registering.value) await auth.register(form.username, form.email, form.password)
    else await auth.login(form.username, form.password)
    await router.replace(typeof route.query.redirect === 'string' ? route.query.redirect : '/')
  } catch (e) {
    error.value = e instanceof APIError ? e.message : t('common.error')
  } finally {
    busy.value = false
  }
}
</script>
