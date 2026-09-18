<template>
  <section>
    <!-- Status bar — three groups with breakdowns -->
    <div class="page-header">
      <div>
        <h1>{{ t('overview.title') }}</h1>
        <p>{{ t('overview.description') }}</p>
      </div>
      <div class="header-actions">
        <span class="auto-refresh-hint" :class="{ refreshing }"><span class="live-dot" aria-hidden="true"></span>{{ t('common.autoRefresh') }}</span>
        <button class="button secondary" type="button" :disabled="busy" @click="load()">
          <Icon name="refresh" size="sm" />{{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div class="status-bar">
      <div class="status-bar-item">
        <h3>{{ t('nav.clients') }}</h3>
        <span class="status-bar-count">{{ data.clients }}</span>
        <div class="status-breakdown">
          <span v-for="item in data.clientBreakdown" :key="item.status" class="breakdown-item">
            <span class="dot" :class="item.status"></span>
            <strong>{{ item.count }}</strong> {{ t(`common.${item.status}`) }}
          </span>
        </div>
      </div>
      <div class="status-bar-item">
        <h3>{{ t('nav.servers') }}</h3>
        <span class="status-bar-count">{{ data.servers }}</span>
        <div class="status-breakdown">
          <span v-for="item in data.serverBreakdown" :key="item.status" class="breakdown-item">
            <span class="dot" :class="item.status"></span>
            <strong>{{ item.count }}</strong> {{ t(`common.${item.status}`) }}
          </span>
        </div>
      </div>
      <div class="status-bar-item">
        <h3>{{ t('nav.tunnels') }}</h3>
        <span class="status-bar-count">{{ data.tunnels }}</span>
        <div class="status-breakdown">
          <span v-for="item in data.tunnelBreakdown" :key="item.status" class="breakdown-item">
            <span class="dot" :class="item.status"></span>
            <strong>{{ item.count }}</strong> {{ t(`common.${item.status}`) }}
          </span>
        </div>
      </div>
    </div>

    <!-- Topology panel -->
    <div class="panel topology-panel">
      <h2>{{ t('overview.topologyTitle') }}</h2>

      <NetworkTopologyMap :nodes="topology.nodes" :links="topology.links" />

      <!-- Getting-started flow -->
      <div class="next-flow">
        <h3>{{ t('overview.next') }}</h3>
        <div class="flow-steps">
          <RouterLink to="/servers"><Icon name="server" size="sm" />{{ t('nav.servers') }}</RouterLink>
          <span class="arrow">→</span>
          <RouterLink to="/clients"><Icon name="users" size="sm" />{{ t('nav.clients') }}</RouterLink>
          <span class="arrow">→</span>
          <RouterLink to="/tunnels"><Icon name="arrowsUpDown" size="sm" />{{ t('nav.tunnels') }}</RouterLink>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '../components/icons/Icon.vue'
import { api } from '../api'
import { useToastStore } from '../stores/toast'
import type { TopologyLink, TopologyNode } from '../api'
import NetworkTopologyMap from '../components/topology/NetworkTopologyMap.vue'
import { useAutoRefresh } from '../composables/useAutoRefresh'

const { t } = useI18n()
const toast = useToastStore()

const data = reactive({
  clients: 0,
  servers: 0,
  tunnels: 0,
  clientBreakdown: [] as { status: string; count: number }[],
  serverBreakdown: [] as { status: string; count: number }[],
  tunnelBreakdown: [] as { status: string; count: number }[],
})
const topology = reactive({ nodes: [] as TopologyNode[], links: [] as TopologyLink[] })

const breakdown = (items: { status: string }[]) => {
  const map = new Map<string, number>()
  for (const item of items) map.set(item.status, (map.get(item.status) || 0) + 1)
  return Array.from(map.entries())
    .map(([status, count]) => ({ status, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 4)
}

const loadData = async () => {
  try {
    const [overview, clientPage, serverPage, topologyResult, tunnelPage] = await Promise.all([
      api.overview(),
      api.clients({ pageSize: 100 }),
      api.servers({ pageSize: 100 }),
      api.topology(),
      api.tunnels({ pageSize: 100 }),
    ])
    Object.assign(data, {
      clients: overview.clients,
      servers: overview.servers,
      tunnels: overview.tunnels,
      clientBreakdown: breakdown(clientPage.items),
      serverBreakdown: breakdown(serverPage.items),
      tunnelBreakdown: breakdown(tunnelPage.items),
    })
    topology.nodes.splice(0, topology.nodes.length, ...topologyResult.nodes)
    topology.links.splice(0, topology.links.length, ...topologyResult.links)
  } catch {
    toast.show(t('common.error'), 'error')
  }
}

const { loading: busy, refreshing, refresh: load } = useAutoRefresh(loadData)
</script>
