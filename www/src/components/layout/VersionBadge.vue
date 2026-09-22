<template>
  <div ref="root" class="sidebar-version">
    <button class="version-pill" type="button" :aria-expanded="open" :title="t('updates.masterVersion')" @click="open = !open">
      <span>{{ displayedVersion }}</span>
      <span v-if="release?.available" class="version-indicator" aria-hidden="true"></span>
      <Icon name="chevronDown" size="xs" />
    </button>
    <div v-if="open" class="version-popover">
      <div class="version-popover-title">
        <strong>{{ t('updates.masterVersion') }}</strong>
        <button class="icon-button" type="button" :title="t('common.refresh')" @click="refresh(true)"><Icon name="refresh" size="xs" /></button>
      </div>
      <p class="version-line">{{ t('updates.current') }}: <span class="mono">{{ release?.currentVersion || '—' }}</span></p>
      <p class="version-line">{{ t('updates.latest') }}: <span class="mono">{{ release?.latestVersion || '—' }}</span></p>
      <p v-if="release?.error" class="field-hint">{{ t('updates.checkFailed') }}: {{ release.error }}</p>
      <p v-else-if="release?.available === false" class="field-hint">{{ t('updates.upToDate') }}</p>
      <p v-if="operation" class="field-hint" role="status">{{ t(`updates.states.${operation.state}`) }}{{ operation.error ? `: ${operation.error}` : '' }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button v-if="canUpdate && release?.available && supported" class="button primary version-update-button" type="button" :disabled="busy || inProgress" @click="confirm = true">
        <Icon name="download" size="xs" />{{ t('updates.updateNow') }}
      </button>
      <p v-else-if="release?.available && !supported" class="field-hint">{{ t('updates.imageBootstrap') }}</p>
      <a v-if="release?.releaseUrl" class="text-button" :href="release.releaseUrl" rel="noopener noreferrer" target="_blank">{{ t('updates.releaseNotes') }}</a>
    </div>
    <ConfirmDialog :show="confirm" :title="t('updates.updateMasterTitle')" :message="t('updates.masterDowntime')" :busy="busy" @cancel="confirm = false" @confirm="start" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { api, type ReleaseSnapshot, type UpdateOperation } from '../../api'
import { useAuthStore } from '../../stores/auth'
import ConfirmDialog from '../common/ConfirmDialog.vue'
import Icon from '../icons/Icon.vue'

const { t } = useI18n()
const auth = useAuthStore()
const root = ref<HTMLElement | null>(null)
const open = ref(false)
const confirm = ref(false)
const busy = ref(false)
const supported = ref(false)
const error = ref('')
const release = ref<ReleaseSnapshot>()
const operation = ref<UpdateOperation>()
const canUpdate = computed(() => auth.user?.role === 'owner')
const inProgress = computed(() => operation.value && !['succeeded', 'failed', 'rolled_back'].includes(operation.value.state))
const displayedVersion = computed(() => {
  const current = release.value
  if (!current?.currentVersion) return '—'
  return current.currentVersion === 'main' ? `main · ${current.currentCommit.slice(0, 7)}` : current.currentVersion
})

let poll: ReturnType<typeof setInterval> | undefined
let refreshTimer: ReturnType<typeof setInterval> | undefined
const watchOperation = () => {
  if (poll || !inProgress.value) return
  poll = setInterval(async () => {
    if (!operation.value) return
    try {
      const result = await api.updateOperation(operation.value.id)
      operation.value = result.operation
      if (['succeeded', 'failed', 'rolled_back'].includes(result.operation.state)) {
        clearInterval(poll)
        poll = undefined
        await refresh(true)
      }
    } catch { /* expected while Master restarts; keep polling */ }
  }, 2500)
}
const refresh = async (force = false) => {
  if (!auth.user) return
  try {
    const result = await api.release(undefined, force)
    release.value = result.release
    supported.value = result.supported
    if (result.operation) { operation.value = result.operation; watchOperation() }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : t('common.error')
  }
}
const start = async () => {
  confirm.value = false
  busy.value = true
  error.value = ''
  try {
    const result = await api.updateMaster()
    operation.value = { id: result.operationId, kind: 'master', targetId: 'master', targetCommit: release.value?.latestCommit || '', version: release.value?.latestVersion || '', state: 'queued', startedAt: new Date().toISOString(), updatedAt: new Date().toISOString() }
    open.value = true
    watchOperation()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : t('common.error')
  } finally { busy.value = false }
}
const closeOutside = (event: MouseEvent) => { if (root.value && !root.value.contains(event.target as Node)) open.value = false }
onMounted(() => {
  refreshTimer = setInterval(() => { if (auth.user) void refresh() }, 10 * 60_000)
  document.addEventListener('click', closeOutside)
})
watch(() => auth.user, (user) => { if (user) void refresh() }, { immediate: true })
onBeforeUnmount(() => {
  if (poll) clearInterval(poll)
  if (refreshTimer) clearInterval(refreshTimer)
  document.removeEventListener('click', closeOutside)
})
</script>
