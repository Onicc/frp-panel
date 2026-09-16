import { useEffect, useState } from 'react'
import { post } from '../api'
import { useI18n } from '../i18n'

type ListResult = { clients?: unknown[]; servers?: unknown[]; proxyConfigs?: unknown[] }

export default function Overview() {
  const { t } = useI18n()
  const [stats, setStats] = useState({ nodes: 0, servers: 0, tunnels: 0 })
  useEffect(() => {
    Promise.allSettled([
      post<ListResult>('/api/v1/client/list', { page: 1, pageSize: 100 }), post<ListResult>('/api/v1/server/list', { page: 1, pageSize: 100 }), post<ListResult>('/api/v1/proxy/list_configs', { page: 1, pageSize: 100 }),
    ]).then(([nodes, servers, tunnels]) => setStats({
      nodes: nodes.status === 'fulfilled' ? nodes.value.clients?.length ?? 0 : 0,
      servers: servers.status === 'fulfilled' ? servers.value.servers?.length ?? 0 : 0,
      tunnels: tunnels.status === 'fulfilled' ? tunnels.value.proxyConfigs?.length ?? 0 : 0,
    }))
  }, [])
  return <>
    <header className="page-header"><div><span className="eyebrow">frp-panel v2</span><h1>{t('overview')}</h1><p>{t('capabilityHint')}</p></div><span className="health"><i />{t('online')}</span></header>
    <section className="stats-grid">
      <article><span>{t('nodes')}</span><strong>{stats.nodes}</strong><small>Managed FRPC agents</small></article>
      <article><span>{t('servers')}</span><strong>{stats.servers}</strong><small>Managed ingress</small></article>
      <article><span>{t('tunnels')}</span><strong>{stats.tunnels}</strong><small>Desired state</small></article>
    </section>
    <section className="panel"><div className="panel-heading"><div><h2>{t('controlPlane')}</h2><p>SQLite / PostgreSQL · TLS · RBAC</p></div></div><div className="activity"><div className="activity-mark">✓</div><div><strong>Secure defaults are active</strong><p>Remote shell, functions and WireGuard require explicit enablement.</p></div></div></section>
  </>
}
