import { useEffect, useState } from 'react'
import { post, request } from '../api'
import { useI18n } from '../i18n'

type ListResult = { clients?: unknown[]; servers?: unknown[]; tunnels?: unknown[] }

export default function Overview() {
  const { t } = useI18n()
  const [stats, setStats] = useState({ clients: 0, servers: 0, tunnels: 0 })
  useEffect(() => {
    Promise.allSettled([
      post<ListResult>('/api/v1/client/list', { page: 1, pageSize: 100 }),
      post<ListResult>('/api/v1/server/list', { page: 1, pageSize: 100 }),
      request<ListResult>('/api/v2/tunnels'),
    ]).then(([clients, servers, tunnels]) => setStats({
      clients: clients.status === 'fulfilled' ? clients.value.clients?.length ?? 0 : 0,
      servers: servers.status === 'fulfilled' ? servers.value.servers?.length ?? 0 : 0,
      tunnels: tunnels.status === 'fulfilled' ? tunnels.value.tunnels?.length ?? 0 : 0,
    }))
  }, [])
  return <>
    <header className="page-header"><div><span className="eyebrow">frp-panel</span><h1>{t('overview')}</h1><p>{t('capabilityHint')}</p></div><span className="health"><i />Master {t('online')}</span></header>
    <section className="stats-grid">
      <article><div className="stat-heading"><span>{t('clients')}</span><span className="resource-icon">C</span></div><strong>{stats.clients}</strong><small>{t('managedClients')}</small></article>
      <article><div className="stat-heading"><span>{t('servers')}</span><span className="resource-icon server-icon">S</span></div><strong>{stats.servers}</strong><small>{t('managedIngress')}</small></article>
      <article><div className="stat-heading"><span>{t('tunnels')}</span><span className="resource-icon tunnel-icon">T</span></div><strong>{stats.tunnels}</strong><small>{t('desiredTunnels')}</small></article>
    </section>
    <section className="panel"><div className="panel-heading"><div><h2>{t('controlPlane')}</h2><p>SQLite / PostgreSQL · TLS · RBAC</p></div><span className="status online"><i />{t('online')}</span></div><div className="activity"><div className="activity-mark">✓</div><div><strong>{t('secureDefaults')}</strong><p>{t('secureDefaultsHint')}</p></div></div></section>
  </>
}
