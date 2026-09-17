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

      <svg
        v-if="data.clients > 0 || data.servers > 0"
        class="topology-svg" viewBox="0 0 600 200" aria-label="Network topology diagram"
      >
        <!-- Master node -->
        <g transform="translate(300,30)">
          <circle r="14" fill="var(--accent-lo)" stroke="var(--accent)" stroke-width="1.5" />
          <text y="28" text-anchor="middle" font-size="10" fill="var(--text-muted)" font-family="var(--font-sans)">Master</text>
        </g>

        <!-- Server nodes -->
        <g
          v-for="(srv, i) in topoServers" :key="srv.id"
          :transform="`translate(${serverX(i, topoServers.length)},100)`"
        >
          <line
            :x1="0" :y1="-14" :x2="300 - serverX(i, topoServers.length)" :y2="-72"
            stroke="var(--border)" stroke-width="1"
          />
          <circle r="10" fill="var(--ok-lo)" :stroke="srv.status === 'online' ? 'var(--ok)' : 'var(--border)'" stroke-width="1.5" />
          <text y="22" text-anchor="middle" font-size="9" fill="var(--text-muted)" font-family="var(--font-sans)">{{ shortId(srv.id) }}</text>
        </g>

        <!-- Client nodes -->
        <g
          v-for="(cli, i) in topoClients" :key="cli.id"
          :transform="`translate(${clientX(i, topoClients.length)},170)`"
        >
          <circle r="8" fill="var(--info-lo)" :stroke="cli.status === 'online' ? 'var(--info)' : 'var(--border)'" stroke-width="1.5" />
          <text y="20" text-anchor="middle" font-size="9" fill="var(--text-muted)" font-family="var(--font-sans)">{{ shortId(cli.id) }}</text>
        </g>

        <!-- Empty call-to-action -->
        <text
          v-if="data.clients === 0 && data.servers === 0"
          x="300" y="110" text-anchor="middle" font-size="13"
          fill="var(--text-muted)" font-family="var(--font-sans)"
        >
          {{ t('overview.topologyEmpty') }}
        </text>
      </svg>

      <div v-if="data.clients === 0 && data.servers === 0" style="padding:32px 0 8px;text-align:center;color:var(--text-muted);font-size:var(--text-sm);">
        {{ t('overview.topologyEmpty') }}
      </div>

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
import type { Client, Server } from '../api'
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
const topoClients = reactive<Client[]>([])
const topoServers = reactive<Server[]>([])

const shortId = (id: string) => id.split('.').pop()?.slice(0, 8) || id.slice(0, 8)

const serverX = (i: number, total: number) => {
  if (total === 0) return 300
  const step = Math.min(500, total * 80)
  const start = 300 - step / 2
  return start + (i / Math.max(total - 1, 1)) * step
}
const clientX = (i: number, total: number) => serverX(i, total)

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
    const [overview, clientPage, serverPage] = await Promise.all([
      api.overview(),
      api.clients({ pageSize: 12 }),
      api.servers({ pageSize: 8 }),
    ])
    Object.assign(data, {
      clients: overview.clients,
      servers: overview.servers,
      tunnels: overview.tunnels,
      clientBreakdown: breakdown(clientPage.items),
      serverBreakdown: breakdown(serverPage.items),
      tunnelBreakdown: [],
    })
    topoClients.splice(0, topoClients.length, ...clientPage.items.slice(0, 8))
    topoServers.splice(0, topoServers.length, ...serverPage.items.slice(0, 5))
  } catch {
    toast.show(t('common.error'), 'error')
  }
}

const { loading: busy, refreshing, refresh: load } = useAutoRefresh(loadData)
</script>
