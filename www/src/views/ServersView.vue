<template>
  <section>
    <div class="page-header">
      <div>
        <h1>{{ t('servers.title') }}</h1>
        <p>{{ t('servers.description') }}</p>
      </div>
      <div class="header-actions">
        <span class="auto-refresh-hint" :class="{ refreshing }"><span class="live-dot" aria-hidden="true"></span>{{ t('common.autoRefresh') }}</span>
        <button class="button secondary" type="button" @click="load()"><Icon name="refresh" size="sm" />{{ t('common.refresh') }}</button>
        <button class="button primary" type="button" @click="openCreate"><Icon name="plus" size="sm" />{{ t('servers.create') }}</button>
      </div>
    </div>

    <!-- Enrollment card -->
    <div v-if="latestCompose" class="install-card">
      <h2>{{ t('servers.deployTitle') }}</h2>
      <p>{{ t('servers.commandHint') }}</p>
      <div class="code-block">
        <pre><code>{{ latestCompose }}</code></pre>
        <button class="button secondary code-copy-btn" type="button" @click="copy(latestCompose)">
          <Icon name="copy" size="xs" />{{ copied ? t('common.copied') : t('common.copy') }}
        </button>
      </div>
      <div class="install-actions">
        <button class="text-button" type="button" @click="latestCompose = ''">{{ t('common.close') }}</button>
      </div>
    </div>

    <div class="table-page-layout">
      <!-- Filters -->
      <div class="filters">
        <SearchInput v-model="search" :placeholder="t('common.search')" />
        <Select v-model="stateFilter" :label="t('servers.state')" :options="stateOptions" />
        <Select v-model="statusFilter" :label="t('servers.status')" :options="statusOptions" />
      </div>

      <DataTable>
        <thead>
          <tr>
            <th>{{ t('servers.id') }}</th>
            <th>{{ t('servers.address') }}</th>
            <th class="num">{{ t('servers.bindPort') }}</th>
            <th class="num">{{ t('servers.serverApiPort') }}</th>
            <th>{{ t('servers.state') }}</th>
            <th>{{ t('servers.status') }}</th>
            <th class="num">{{ t('servers.tunnels') }}</th>
            <th class="actions-col">&nbsp;</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in rows" :key="item.id">
            <td>
              <div class="resource-name">
                <span class="resource-icon server-icon" aria-hidden="true"><Icon name="server" size="xs" /></span>
                <strong>{{ item.id }}</strong>
              </div>
            </td>
            <td>{{ item.address }}</td>
            <td class="num mono">{{ item.bindPort }}</td>
            <td class="num mono">{{ item.serverApiPort }}</td>
            <td><StatusBadge :status="item.configurationState" /></td>
            <td><StatusBadge :status="item.status" /></td>
            <td class="num">{{ item.tunnelCount }}</td>
            <td class="row-actions">
              <button class="icon-button" type="button" :title="t('common.edit')" @click="openEdit(item)"><Icon name="edit" size="sm" /></button>
              <button class="icon-button" type="button" :title="t('common.rotate')" @click="openRotate(item)"><Icon name="key" size="sm" /></button>
              <button class="icon-button danger-icon" type="button" :title="t('common.delete')" @click="openDelete(item)"><Icon name="trash" size="sm" /></button>
            </td>
          </tr>
        </tbody>
      </DataTable>
      <EmptyState v-if="!loading && !rows.length" :title="t('servers.empty')" :message="t('servers.emptyHint')">
        <button class="button primary" type="button" @click="openCreate">{{ t('servers.create') }}</button>
      </EmptyState>
      <div v-if="loading" class="loading-row"><LoadingSpinner />{{ t('common.loading') }}</div>

      <Pagination :page="page" :page-size="pageSize" :total="total" @update:page="page = $event; load()" />
    </div>

    <!-- Create / Edit dialog -->
    <BaseDialog
      :show="dialog === 'create' || dialog === 'edit'"
      :title="dialog === 'create' ? t('servers.createTitle') : t('servers.editTitle')"
      @close="closeDialog"
    >
      <form @submit.prevent="saveServer">
        <Input v-model="form.serverId" :label="t('servers.id')" :disabled="dialog === 'edit'" required />
        <Input v-model="form.address" :label="t('servers.address')" required />
        <div class="form-grid">
          <Input v-model.number="form.bindPort" :label="t('servers.bindPort')" type="number" min="1024" max="65535" required />
          <Input
            v-model.number="form.serverApiPort"
            :label="t('servers.serverApiPort')"
            :hint="t('servers.portHint')"
            :error="portConflict ? t('servers.portConflict') : undefined"
            type="number"
            min="1024"
            max="65535"
            required
          />
        </div>
        <Input v-model="form.comment" :label="t('servers.comment')" />
        <p v-if="modalError" class="error" role="alert">{{ modalError }}</p>
      </form>
      <template #footer>
        <button class="button secondary" type="button" @click="closeDialog">{{ t('common.cancel') }}</button>
        <button class="button primary" type="button" :disabled="busy" @click="saveServer">{{ busy ? t('auth.loading') : t('common.save') }}</button>
      </template>
    </BaseDialog>

    <!-- Rotate dialog -->
    <BaseDialog :show="dialog === 'rotate'" :title="t('servers.rotateTitle')" width="narrow" @close="closeDialog">
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
      :title="t('servers.deleteTitle')"
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
  Input, LoadingSpinner, Pagination, SearchInput, Select, StatusBadge,
} from '../components/common'
import Icon from '../components/icons/Icon.vue'
import { api, APIError, type Server } from '../api'
import { useToastStore } from '../stores/toast'
import { useAutoRefresh } from '../composables/useAutoRefresh'

const { t } = useI18n()
const toast = useToastStore()

const rows     = ref<Server[]>([])
const busy     = ref(false)
const page     = ref(1)
const pageSize = 25
const total    = ref(0)
const search       = ref('')
const stateFilter  = ref('')
const statusFilter = ref('')
const dialog   = ref<'create'|'edit'|'rotate'|'delete'|''>('')
const selected = ref<Server>()
const modalError   = ref('')
const copied       = ref(false)
const latestCompose = ref('')
const statuses = ['pending', 'online', 'offline', 'error', 'disabled']
const stateOptions = computed(() => [
  { value: '', label: t('servers.allConfiguration') },
  { value: 'configured', label: t('common.configured') },
  { value: 'unconfigured', label: t('common.unconfigured') },
])
const statusOptions = computed(() => [
  { value: '', label: t('servers.allStatus') },
  ...statuses.map((value) => ({ value, label: t(`common.${value}`) })),
])
const form = reactive({ serverId: '', address: '', bindPort: 7000, serverApiPort: 8999, comment: '', acknowledge: false })

const portConflict = computed(() => Number(form.bindPort) > 0 && Number(form.bindPort) === Number(form.serverApiPort))

const deleteMessage = computed(() =>
  selected.value ? `${t('servers.deleteHint')} ${selected.value.id}` : t('servers.deleteHint')
)

const loadData = async () => {
  try {
    const result = await api.servers({ page: page.value, pageSize, search: search.value, configurationState: stateFilter.value, status: statusFilter.value })
    rows.value  = result.items
    total.value = result.total
  } catch (e) {
    toast.show(e instanceof APIError ? e.message : t('common.error'), 'error')
  }
}
const { loading, refreshing, refresh: load } = useAutoRefresh(loadData)
watch([search, stateFilter, statusFilter], () => { page.value = 1; void load() })

const reset = () => { form.serverId = ''; form.address = ''; form.bindPort = 7000; form.serverApiPort = 8999; form.comment = ''; form.acknowledge = false; modalError.value = '' }
const closeDialog = () => { dialog.value = ''; reset() }
const openCreate  = () => { reset(); dialog.value = 'create' }
const openEdit    = (item: Server) => { reset(); selected.value = item; form.serverId = item.id; form.address = item.address; form.bindPort = item.bindPort; form.serverApiPort = item.serverApiPort || 8999; form.comment = item.comment; dialog.value = 'edit' }
const openRotate  = (item: Server) => { reset(); selected.value = item; dialog.value = 'rotate' }
const openDelete  = (item: Server) => { selected.value = item; dialog.value = 'delete' }

const saveServer = async () => {
  modalError.value = ''
  if (!form.serverId || !form.address) { modalError.value = t('problems.required'); return }
  if (portConflict.value) { modalError.value = t('servers.portConflict'); return }
  busy.value = true
  try {
    if (dialog.value === 'create') {
      const result = await api.createServer({ serverId: form.serverId, address: form.address, bindPort: form.bindPort, serverApiPort: form.serverApiPort, comment: form.comment })
      latestCompose.value = result.enrollment.composeYaml || ''
      toast.show(t('common.success'))
    } else if (selected.value) {
      await api.updateServer(selected.value.id, { address: form.address, bindPort: form.bindPort, serverApiPort: form.serverApiPort, comment: form.comment })
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
    const result = await api.rotateServer(selected.value.id, form.acknowledge)
    latestCompose.value = result.enrollment.composeYaml || ''
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
    await api.deleteServer(selected.value.id)
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
