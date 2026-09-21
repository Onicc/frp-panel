<template>
  <section>
    <div class="page-header">
      <div>
        <h1>{{ t('tunnels.title') }}</h1>
        <p>{{ t('tunnels.description') }}</p>
      </div>
      <div class="header-actions">
        <span class="auto-refresh-hint" :class="{ refreshing }"><span class="live-dot" aria-hidden="true"></span>{{ t('common.autoRefresh') }}</span>
        <button class="button secondary" type="button" @click="load()"><Icon name="refresh" size="sm" />{{ t('common.refresh') }}</button>
        <button class="button primary" type="button" :disabled="!clients.length || !servers.length" @click="openCreate">
          <Icon name="plus" size="sm" />{{ t('tunnels.create') }}
        </button>
      </div>
    </div>

    <div v-if="!loading && (!clients.length || !servers.length)" class="notice-card">
      <strong>{{ t('tunnels.emptyHint') }}</strong>
      <p>
        <RouterLink to="/clients">{{ t('nav.clients') }}</RouterLink>
        &nbsp;·&nbsp;
        <RouterLink to="/servers">{{ t('nav.servers') }}</RouterLink>
      </p>
    </div>

    <div class="table-page-layout">
      <div class="filters">
        <SearchInput v-model="search" :placeholder="t('common.search')" />
        <Select v-model="statusFilter" :label="t('tunnels.status')" :options="statusOptions" />
      </div>

      <DataTable>
        <thead>
          <tr>
            <th>{{ t('tunnels.name') }}</th>
            <th>{{ t('tunnels.client') }}</th>
            <th>{{ t('tunnels.server') }}</th>
            <th>{{ t('tunnels.type') }}</th>
            <th>{{ t('tunnels.connection') }}</th>
            <th>{{ t('tunnels.remoteConnection') }}</th>
            <th>{{ t('tunnels.status') }}</th>
            <th class="actions-col">&nbsp;</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in rows" :key="item.id">
            <td>
              <div class="resource-name">
                <span class="resource-icon tunnel-icon" aria-hidden="true"><Icon name="arrowsUpDown" size="xs" /></span>
                <strong>{{ item.name }}</strong>
              </div>
            </td>
            <td class="muted">{{ shortId(item.clientId) }}</td>
            <td class="muted">{{ shortId(item.serverId) }}</td>
            <td><span class="protocol-badge">{{ item.type }}</span></td>
            <td><span class="mono">{{ item.localHost }}:{{ item.localPort }}</span></td>
            <td class="mono">{{ remoteConnection(item) }}</td>
            <td><StatusBadge :status="item.status" /></td>
            <td class="row-actions">
              <button class="icon-button" type="button" :title="t('common.edit')" @click="openEdit(item)"><Icon name="edit" size="sm" /></button>
              <button class="icon-button danger-icon" type="button" :title="t('common.delete')" @click="openDelete(item)"><Icon name="trash" size="sm" /></button>
            </td>
          </tr>
        </tbody>
      </DataTable>
      <EmptyState v-if="!loading && !rows.length" :title="t('tunnels.empty')" :message="t('tunnels.emptyHint')" />
      <div v-if="loading" class="loading-row"><LoadingSpinner />{{ t('common.loading') }}</div>

      <Pagination :page="page" :page-size="pageSize" :total="total" @update:page="page = $event; load()" />
    </div>

    <!-- Create / Edit dialog -->
    <BaseDialog
      :show="dialog === 'create' || dialog === 'edit'"
      :title="dialog === 'create' ? t('tunnels.create') : t('tunnels.editTitle')"
      width="wide" @close="closeDialog"
    >
      <form @submit.prevent="save">
        <div class="form-grid">
          <Input v-model="form.name" :label="t('tunnels.name')" required />
          <Select v-model="form.type" :label="t('tunnels.type')" :options="typeOptions" required />
          <Select v-model="form.clientId" :label="t('tunnels.client')" :options="clientOptions" :placeholder="t('tunnels.client')" required />
          <Select v-model="form.serverId" :label="t('tunnels.server')" :options="serverOptions" :placeholder="t('tunnels.server')" required />
          <Input v-model="form.localHost" :label="t('tunnels.localHost')" required />
          <Input v-model.number="form.localPort" :label="t('tunnels.localPort')" type="number" min="1" max="65535" required />
          <Input v-model.number="form.remotePort" :label="t('tunnels.remotePort')" type="number" min="1024" max="65535" required />
        </div>
        <div class="inline-toggle form-toggle">
          <span class="toggle-copy"><strong>{{ t('common.enabled') }}</strong><small>{{ t('tunnels.enabledHint') }}</small></span>
          <Toggle v-model="form.enabled" />
        </div>
        <p v-if="modalError" class="error" role="alert">{{ modalError }}</p>
      </form>
      <template #footer>
        <button class="button secondary" type="button" @click="closeDialog">{{ t('common.cancel') }}</button>
        <button class="button primary" type="button" :disabled="busy" @click="save">{{ busy ? t('auth.loading') : t('common.save') }}</button>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="dialog === 'delete'"
      :title="t('tunnels.deleteTitle')"
      :message="deleteMessage"
      :warning="t('tunnels.deleteWarning')"
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
import { api, APIError, type Client, type Server, type Tunnel } from '../api'
import { useToastStore } from '../stores/toast'
import { useAutoRefresh } from '../composables/useAutoRefresh'

const { t } = useI18n()
const toast = useToastStore()

const rows    = ref<Tunnel[]>([])
const clients = ref<Client[]>([])
const servers = ref<Server[]>([])
const busy    = ref(false)
const page     = ref(1)
const pageSize = 25
const total    = ref(0)
const search       = ref('')
const statusFilter = ref('')
const dialog   = ref<'create'|'edit'|'delete'|''>('')
const selected = ref<Tunnel>()
const modalError   = ref('')
const statuses = ['pending', 'online', 'offline', 'error', 'disabled']
const statusOptions = computed(() => [
  { value: '', label: t('tunnels.allStatus') },
  ...statuses.map((value) => ({ value, label: t(`common.${value}`) })),
])
const typeOptions = computed(() => [
  { value: 'tcp', label: 'TCP', description: 'Transmission Control Protocol' },
  { value: 'udp', label: 'UDP', description: 'User Datagram Protocol' },
])
const clientOptions = computed(() => clients.value.map((item) => ({
  value: item.id,
  label: item.id,
  description: `${t('clients.state')}: ${t(`common.${item.configurationState}`)} · ${t('clients.status')}: ${t(`common.${item.status}`)}`,
})))
const serverOptions = computed(() => servers.value.map((item) => ({
  value: item.id,
  label: item.id,
  description: `${t('servers.state')}: ${t(`common.${item.configurationState}`)} · ${t('servers.status')}: ${t(`common.${item.status}`)}`,
})))
const form = reactive({ name: '', clientId: '', serverId: '', type: 'tcp', localHost: '127.0.0.1', localPort: 80, remotePort: 8080, enabled: true })

const deleteMessage = computed(() => selected.value ? `${t('tunnels.deleteMessage')} ${selected.value.name}` : '')

const shortId = (id: string) => id.split('.').pop() || id
const remoteConnection = (item: Tunnel) => {
  const address = item.serverAddress || shortId(item.serverId)
  const displayAddress = address.includes(':') && !address.startsWith('[') ? `[${address}]` : address
  return `${displayAddress}:${item.remotePort}`
}

const loadData = async () => {
  try {
    const [tunnelPage, clientPage, serverPage] = await Promise.all([
      api.tunnels({ page: page.value, pageSize, search: search.value, status: statusFilter.value }),
      api.clients({ pageSize: 100 }),
      api.servers({ pageSize: 100 }),
    ])
    rows.value    = tunnelPage.items
    total.value   = tunnelPage.total
    clients.value = clientPage.items
    servers.value = serverPage.items
  } catch (e) {
    toast.show(e instanceof APIError ? e.message : t('common.error'), 'error')
  }
}
const { loading, refreshing, refresh: load } = useAutoRefresh(loadData)
watch([search, statusFilter], () => { page.value = 1; void load() })

const reset = () => {
  form.name = ''; form.clientId = clients.value[0]?.id || ''; form.serverId = servers.value[0]?.id || ''
  form.type = 'tcp'; form.localHost = '127.0.0.1'; form.localPort = 80; form.remotePort = 8080; form.enabled = true
  modalError.value = ''
}
const closeDialog = () => { dialog.value = ''; reset() }
const openCreate  = () => { reset(); dialog.value = 'create' }
const openEdit    = (item: Tunnel) => {
  selected.value = item
  form.name = item.name; form.clientId = item.clientId; form.serverId = item.serverId
  form.type = item.type; form.localHost = item.localHost; form.localPort = item.localPort
  form.remotePort = item.remotePort; form.enabled = item.enabled
  modalError.value = ''; dialog.value = 'edit'
}
const openDelete = (item: Tunnel) => { selected.value = item; dialog.value = 'delete' }

const save = async () => {
  modalError.value = ''
  if (!form.name || !form.clientId || !form.serverId) { modalError.value = t('problems.required'); return }
  busy.value = true
  try {
    const body = { name: form.name, clientId: form.clientId, serverId: form.serverId, type: form.type, localHost: form.localHost, localPort: form.localPort, remotePort: form.remotePort, enabled: form.enabled }
    if (dialog.value === 'create') await api.createTunnel(body)
    else if (selected.value) await api.updateTunnel(selected.value.id, body)
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
    await api.deleteTunnel(selected.value.id)
    toast.show(t('common.success'))
    closeDialog()
    await load()
  } catch (e) {
    modalError.value = e instanceof APIError ? e.message : t('common.error')
  } finally {
    busy.value = false
  }
}
</script>
