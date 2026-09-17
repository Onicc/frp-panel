<template>
  <section>
    <div class="page-header">
      <div>
        <h1>{{ t('clients.title') }}</h1>
        <p>{{ t('clients.description') }}</p>
      </div>
      <div class="header-actions">
        <span class="auto-refresh-hint" :class="{ refreshing }"><span class="live-dot" aria-hidden="true"></span>{{ t('common.autoRefresh') }}</span>
        <button class="button secondary" type="button" @click="load()"><Icon name="refresh" size="sm" />{{ t('common.refresh') }}</button>
        <button class="button primary" type="button" @click="openCreate"><Icon name="plus" size="sm" />{{ t('clients.create') }}</button>
      </div>
    </div>

    <!-- Enrollment card -->
    <div v-if="Object.keys(enrollmentCommands).length" class="install-card">
      <h2>{{ t('clients.installTitle') }}</h2>
      <p>{{ t('clients.commandHint') }}</p>
      <div class="platform-tabs" role="tablist" :aria-label="t('clients.platform')">
        <button
          v-for="item in platforms" :key="item.id"
          class="platform-tab" :class="{active: platform === item.id}"
          type="button" role="tab" :aria-selected="platform === item.id"
          @click="platform = item.id"
        >
          {{ item.label }}
        </button>
      </div>
      <div class="code-block">
        <pre><code>{{ latestInstall }}</code></pre>
        <button class="button secondary code-copy-btn" type="button" @click="copy(latestInstall)">
          <Icon name="copy" size="xs" />{{ copied ? t('common.copied') : t('common.copy') }}
        </button>
      </div>
      <div class="install-actions">
        <button class="text-button" type="button" @click="clearEnrollment">{{ t('common.close') }}</button>
      </div>
    </div>

    <div class="table-page-layout">
      <div class="filters">
        <SearchInput v-model="search" :placeholder="t('common.search')" />
        <Select v-model="stateFilter" :label="t('clients.state')" :options="stateOptions" />
        <Select v-model="statusFilter" :label="t('clients.status')" :options="statusOptions" />
      </div>

      <DataTable>
        <thead>
          <tr>
            <th>{{ t('clients.id') }}</th>
            <th>{{ t('clients.state') }}</th>
            <th>{{ t('clients.status') }}</th>
            <th class="num">{{ t('clients.tunnels') }}</th>
            <th>{{ t('clients.comment') }}</th>
            <th class="actions-col">&nbsp;</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in rows" :key="item.id">
            <td>
              <div class="resource-name">
                <span class="resource-icon client-icon" aria-hidden="true"><Icon name="users" size="xs" /></span>
                <strong>{{ item.id }}</strong>
              </div>
            </td>
            <td><StatusBadge :status="item.configurationState" /></td>
            <td><StatusBadge :status="item.status" /></td>
            <td class="num">{{ item.tunnelCount }}</td>
            <td class="muted">{{ item.comment || '—' }}</td>
            <td class="row-actions">
              <button class="icon-button" type="button" :title="t('common.edit')" @click="openEdit(item)"><Icon name="edit" size="sm" /></button>
              <button class="icon-button" type="button" :title="t('common.rotate')" @click="openRotate(item)"><Icon name="key" size="sm" /></button>
              <button class="icon-button danger-icon" type="button" :title="t('common.delete')" @click="openDelete(item)"><Icon name="trash" size="sm" /></button>
            </td>
          </tr>
        </tbody>
      </DataTable>
      <EmptyState v-if="!loading && !rows.length" :title="t('clients.empty')" :message="t('clients.emptyHint')">
        <button class="button primary" type="button" @click="openCreate">{{ t('clients.create') }}</button>
      </EmptyState>
      <div v-if="loading" class="loading-row"><LoadingSpinner />{{ t('common.loading') }}</div>

      <Pagination :page="page" :page-size="pageSize" :total="total" @update:page="page = $event; load()" />
    </div>

    <!-- Create / Edit dialog -->
    <BaseDialog
      :show="dialog === 'create' || dialog === 'edit'"
      :title="dialog === 'create' ? t('clients.createTitle') : t('clients.editTitle')"
      @close="closeDialog"
    >
      <form @submit.prevent="saveClient">
        <Input v-model="form.clientId" :label="t('clients.id')" :disabled="dialog === 'edit'" required />
        <Input v-model="form.comment" :label="t('clients.comment')" />
        <div v-if="dialog === 'edit'" class="inline-toggle form-toggle">
          <span class="toggle-copy"><strong>{{ t('common.enabled') }}</strong><small>{{ t('clients.enabledHint') }}</small></span>
          <Toggle v-model="form.enabled" />
        </div>
        <p v-if="modalError" class="error" role="alert">{{ modalError }}</p>
      </form>
      <template #footer>
        <button class="button secondary" type="button" @click="closeDialog">{{ t('common.cancel') }}</button>
        <button class="button primary" type="button" :disabled="busy" @click="saveClient">{{ busy ? t('auth.loading') : t('common.save') }}</button>
      </template>
    </BaseDialog>

    <!-- Rotate dialog -->
    <BaseDialog :show="dialog === 'rotate'" :title="t('clients.rotateTitle')" width="narrow" @close="closeDialog">
      <p class="dialog-copy">{{ t('clients.rotateHint') }}</p>
      <label class="check-row">
        <input v-model="form.acknowledge" type="checkbox" />
        {{ t('clients.acknowledge') }}
      </label>
      <p v-if="modalError" class="error" role="alert">{{ modalError }}</p>
      <template #footer>
        <button class="button secondary" type="button" @click="closeDialog">{{ t('common.cancel') }}</button>
        <button
          class="button primary" type="button"
          :disabled="busy || (!form.acknowledge && selected?.configurationState === 'configured')"
          @click="rotate"
        >
          {{ busy ? t('auth.loading') : t('common.rotate') }}
        </button>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="dialog === 'delete'"
      :title="t('clients.deleteTitle')"
      :message="deleteMessage"
      :error="modalError"
      :busy="busy"
      @cancel="closeDialog"
      @confirm="remove"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  BaseDialog, ConfirmDialog, DataTable, EmptyState,
  Input, LoadingSpinner, Pagination, SearchInput, Select, StatusBadge, Toggle,
} from '../components/common'
import Icon from '../components/icons/Icon.vue'
import { api, APIError, type Client } from '../api'
import { useToastStore } from '../stores/toast'
import { useAutoRefresh } from '../composables/useAutoRefresh'

const { t } = useI18n()
const toast = useToastStore()

const rows     = ref<Client[]>([])
const busy     = ref(false)
const page     = ref(1)
const pageSize = 25
const total    = ref(0)
const search       = ref('')
const stateFilter  = ref('')
const statusFilter = ref('')
const dialog   = ref<'create'|'edit'|'rotate'|'delete'|''>('')
const selected = ref<Client>()
const modalError = ref('')
const copied   = ref(false)
const enrollmentCommands = ref<Partial<Record<'linux'|'darwin'|'windows', string>>>({})
const platform = ref<'linux'|'darwin'|'windows'>('linux')
const platforms = [
  { id: 'linux'   as const, label: 'Linux' },
  { id: 'darwin'  as const, label: 'macOS' },
  { id: 'windows' as const, label: 'Windows' },
]
const latestInstall = computed(() =>
  enrollmentCommands.value[platform.value] || enrollmentCommands.value.linux || ''
)
const statuses = ['pending', 'online', 'offline', 'error', 'disabled']
const stateOptions = computed(() => [
  { value: '', label: t('clients.allConfiguration') },
  { value: 'configured', label: t('common.configured') },
  { value: 'unconfigured', label: t('common.unconfigured') },
])
const statusOptions = computed(() => [
  { value: '', label: t('clients.allStatus') },
  ...statuses.map((value) => ({ value, label: t(`common.${value}`) })),
])
const form = reactive({ clientId: '', comment: '', enabled: true, acknowledge: false })

const deleteMessage = computed(() =>
  selected.value ? `${t('clients.deleteHint')} ${selected.value.id}` : t('clients.deleteHint')
)

const loadData = async () => {
  try {
    const result = await api.clients({ page: page.value, pageSize, search: search.value, configurationState: stateFilter.value, status: statusFilter.value })
    rows.value  = result.items
    total.value = result.total
  } catch (e) {
    toast.show(e instanceof APIError ? e.message : t('common.error'), 'error')
  }
}
const { loading, refreshing, refresh: load } = useAutoRefresh(loadData)
watch([search, stateFilter, statusFilter], () => { page.value = 1; void load() })

const reset = () => { form.clientId = ''; form.comment = ''; form.enabled = true; form.acknowledge = false; modalError.value = '' }
const closeDialog    = () => { dialog.value = ''; reset() }
const clearEnrollment = () => { enrollmentCommands.value = {}; copied.value = false }
const setEnrollment  = (value: { installCommands?: Record<'linux'|'darwin'|'windows', string>; installCommand?: string }) => {
  enrollmentCommands.value = value.installCommands || { linux: value.installCommand || '' }
  platform.value = 'linux'
}
const openCreate = () => { reset(); dialog.value = 'create' }
const openEdit   = (item: Client) => { reset(); selected.value = item; form.clientId = item.id; form.comment = item.comment; form.enabled = item.enabled; dialog.value = 'edit' }
const openRotate = (item: Client) => { reset(); selected.value = item; dialog.value = 'rotate' }
const openDelete = (item: Client) => { selected.value = item; dialog.value = 'delete' }

const saveClient = async () => {
  modalError.value = ''
  if (!form.clientId) { modalError.value = t('problems.required'); return }
  busy.value = true
  try {
    if (dialog.value === 'create') {
      const result = await api.createClient({ clientId: form.clientId, comment: form.comment })
      setEnrollment(result.enrollment)
      toast.show(t('common.success'))
    } else if (selected.value) {
      await api.updateClient(selected.value.id, { comment: form.comment, enabled: form.enabled })
      toast.show(t('common.success'))
    }
    closeDialog()
    await load()
  } catch (e) {
    modalError.value = e instanceof APIError ? e.message : t('common.error')
  } finally {
    busy.value = false
  }
}

const rotate = async () => {
  if (!selected.value) return
  modalError.value = ''
  busy.value = true
  try {
    const result = await api.rotateClient(selected.value.id, form.acknowledge)
    setEnrollment(result.enrollment)
    toast.show(t('common.success'))
    closeDialog()
    await load()
  } catch (e) {
    modalError.value = e instanceof APIError ? e.message : t('common.error')
  } finally {
    busy.value = false
  }
}

const remove = async () => {
  if (!selected.value) return
  busy.value = true
  try {
    await api.deleteClient(selected.value.id)
    toast.show(t('common.success'))
    closeDialog()
    await load()
  } catch (e) {
    modalError.value = e instanceof APIError ? e.message : t('problems.dependencies')
  } finally {
    busy.value = false
  }
}

const copy = async (value: string) => {
  await navigator.clipboard?.writeText(value)
  copied.value = true
  toast.show(t('common.copied'))
  window.setTimeout(() => copied.value = false, 1800)
}
</script>
