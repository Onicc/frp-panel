<template>
  <div ref="root" class="sidebar-version">
    <button v-if="canUpdate" class="version-pill" :class="{ 'has-update': release?.available }" type="button" :aria-expanded="open" aria-controls="master-version-popover" :title="displayVersionDetails(release?.currentVersion, release?.currentCommit, versionLabels)" @click="open = !open">
      <span>{{ currentVersion }}</span>
      <span v-if="release?.available" class="version-indicator" aria-hidden="true"><span /></span>
    </button>
    <span v-else class="version-static" :title="displayVersionDetails(release?.currentVersion, release?.currentCommit, versionLabels)">{{ currentVersion }}</span>

    <Transition name="version-dropdown">
      <div v-if="canUpdate && open" id="master-version-popover" class="version-popover" @click.stop>
        <div class="version-popover-header">
          <span>{{ t('updates.masterVersion') }}</span>
          <button class="version-refresh" type="button" :title="t('common.refresh')" :disabled="loading" @click="refresh(true)"><Icon name="refresh" size="sm" :class="{ 'version-spin': loading }" /></button>
        </div>
        <div class="version-popover-body">
          <div v-if="loading && !release" class="version-loading" role="status"><Icon name="refresh" size="lg" class="version-spin" /></div>
          <template v-else>
            <div class="version-current">
              <div class="version-current-value">
                <strong>{{ currentVersion }}</strong>
                <span v-if="release?.available === false && !release?.error" class="version-current-check" aria-hidden="true"><Icon name="check" size="xs" /></span>
              </div>
              <p v-if="release?.available" :title="displayVersionDetails(release.latestVersion, release.latestCommit)">{{ t('updates.latest') }}: {{ latestVersion }}</p>
              <p v-else-if="release?.available === false && !release?.error">{{ t('updates.upToDate') }}</p>
              <p v-else-if="release?.channel === 'legacy'">{{ t('updates.migration') }}</p>
              <p v-else-if="release?.error">{{ t('updates.checkFailed') }}</p>
              <p v-else>{{ t('updates.checking') }}</p>
            </div>

            <div v-if="release?.channel === 'legacy'" class="version-notice version-notice-info" role="status">
              <span class="version-notice-icon"><Icon name="clock" size="sm" /></span>
              <span class="version-notice-copy"><strong>{{ t('updates.migration') }}</strong><small>{{ t('updates.masterMigration') }}</small></span>
            </div>
            <div v-else-if="operation && inProgress" class="version-notice version-notice-info" role="status">
              <span class="version-notice-icon"><Icon name="refresh" size="sm" class="version-spin" /></span>
              <span class="version-notice-copy"><strong>{{ t('updates.updating') }}</strong><small>{{ t(`updates.states.${operation.state}`) }}</small></span>
            </div>
            <div v-else-if="operation?.state === 'succeeded'" class="version-notice version-notice-success" role="status">
              <span class="version-notice-icon"><Icon name="check" size="sm" /></span>
              <span class="version-notice-copy"><strong>{{ t('updates.states.succeeded') }}</strong><small>{{ currentVersion }}</small></span>
            </div>
            <div v-else-if="error || release?.error || operation?.state === 'failed' || operation?.state === 'rolled_back'" class="version-notice version-notice-error" role="alert">
              <span class="version-notice-icon"><Icon name="x" size="sm" /></span>
              <span class="version-notice-copy"><strong>{{ t('updates.updateFailed') }}</strong><small>{{ error || operation?.error || release?.error || t(`updates.states.${operation?.state}`) }}</small></span>
            </div>
            <div v-else-if="release?.available" class="version-notice version-notice-update" :title="displayVersionDetails(release.latestVersion, release.latestCommit)">
              <span class="version-notice-icon"><Icon name="download" size="sm" /></span>
              <span class="version-notice-copy"><strong>{{ t('updates.newVersion') }}</strong><small>{{ latestVersion }}</small></span>
            </div>

            <p v-if="release?.available && !supported" class="version-hint">{{ t('updates.imageBootstrap') }}</p>
            <button v-if="canUpdate && release?.available && supported && !inProgress" class="version-action" type="button" :disabled="busy" @click="confirm = true"><Icon name="download" size="sm" />{{ busy ? t('updates.updating') : t('updates.updateNow') }}</button>
            <button v-if="(error || release?.error) && release?.channel !== 'legacy' && !loading" class="version-retry" type="button" @click="refresh(true)">{{ t('updates.retry') }}</button>
            <a v-if="release?.releaseUrl" class="version-release-link" :href="release.releaseUrl" rel="noopener noreferrer" target="_blank">{{ t('updates.releaseNotes') }} <Icon name="externalLink" size="xs" /></a>
            <div v-if="canUpdate && release?.available === false && !operation" class="version-rollback">
              <button class="version-rollback-toggle" type="button" :aria-expanded="rollbackDetailsOpen" @click="rollbackDetailsOpen = !rollbackDetailsOpen">
                <span><Icon name="clock" size="xs" />{{ t('updates.rollback') }}</span>
                <Icon name="chevronDown" size="xs" :class="{ 'is-open': rollbackDetailsOpen }" />
              </button>
              <p v-if="rollbackDetailsOpen" class="version-rollback-hint">{{ t('updates.rollbackManualHint') }}</p>
            </div>
          </template>
        </div>
      </div>
    </Transition>
    <ConfirmDialog :show="confirm" :title="t('updates.updateMasterTitle')" :message="t('updates.masterDowntime')" :busy="busy" @cancel="confirm = false" @confirm="start" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { api, type ReleaseSnapshot, type UpdateOperation } from '../../api'
import { displayVersion, displayVersionDetails } from '../../composables/displayVersion'
import { useAuthStore } from '../../stores/auth'
import ConfirmDialog from '../common/ConfirmDialog.vue'
import Icon from '../icons/Icon.vue'

const { t } = useI18n()
const auth = useAuthStore()
const root = ref<HTMLElement | null>(null)
const open = ref(false)
const rollbackDetailsOpen = ref(false)
const confirm = ref(false)
const busy = ref(false)
const loading = ref(false)
const supported = ref(false)
const error = ref('')
const release = ref<ReleaseSnapshot>()
const operation = ref<UpdateOperation>()
const canUpdate = computed(() => auth.user?.role === 'owner')
const inProgress = computed(() => !!operation.value && !['succeeded', 'failed', 'rolled_back'].includes(operation.value.state))
const versionLabels = computed(() => ({ legacy: t('updates.legacyVersion'), development: t('updates.developmentVersion') }))
const currentVersion = computed(() => displayVersion(release.value?.currentVersion, versionLabels.value))
const latestVersion = computed(() => displayVersion(release.value?.latestVersion, versionLabels.value))

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
  loading.value = true
  error.value = ''
  try {
    const result = await api.release(undefined, force)
    release.value = result.release
    supported.value = result.supported
    if (result.operation) { operation.value = result.operation; watchOperation() }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : t('common.error')
  } finally { loading.value = false }
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
const closeOnEscape = (event: KeyboardEvent) => { if (event.key === 'Escape') open.value = false }
onMounted(() => {
  refreshTimer = setInterval(() => { if (auth.user) void refresh() }, 10 * 60_000)
  document.addEventListener('click', closeOutside)
  document.addEventListener('keydown', closeOnEscape)
})
watch(() => auth.user, (user) => { if (user) void refresh() }, { immediate: true })
onBeforeUnmount(() => {
  if (poll) clearInterval(poll)
  if (refreshTimer) clearInterval(refreshTimer)
  document.removeEventListener('click', closeOutside)
  document.removeEventListener('keydown', closeOnEscape)
})
</script>
