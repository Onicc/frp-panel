import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Node = { id?: string; clientId?: string; stopped?: boolean; ephemeral?: boolean; online?: boolean }
type Server = { id?: string; serverId?: string; ip?: string }
type ClientList = { clients?: Node[] }
type PlatformInfo = { clientApiUrl?: string; clientRpcUrl?: string; client_api_url?: string; client_rpc_url?: string }
type Enrollment = { nodeId: string; token: string; expiresAt: string }
type Route = { nodeId: string; serverIds: string[] }
type StatusResponse = { clients?: Record<string, { status?: number }> }
type OS = 'linux' | 'macos' | 'windows'

function shellQuote(value: string): string { return `'${value.replaceAll("'", `'"'"'`)}'` }
function psQuote(value: string): string { return `'${value.replaceAll("'", "''")}'` }

export default function Nodes() {
  const { t } = useI18n()
  const [nodes, setNodes] = useState<Node[]>([])
  const [servers, setServers] = useState<Server[]>([])
  const [routes, setRoutes] = useState<Record<string, string[]>>({})
  const [routeDrafts, setRouteDrafts] = useState<Record<string, string>>({})
  const [routeError, setRouteError] = useState('')
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [nodeId, setNodeId] = useState('')
  const [install, setInstall] = useState<Enrollment | null>(null)
  const [platform, setPlatform] = useState<PlatformInfo>({})
  const [os, setOS] = useState<OS>('linux')
  const [copied, setCopied] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const [clientList, serverList, routeList] = await Promise.all([
        post<ClientList>('/api/v1/client/list', { page: 1, pageSize: 100 }),
        post<{ servers?: Server[] }>('/api/v1/server/list', { page: 1, pageSize: 100 }),
        request<{ routes?: Route[] }>('/api/v2/node-routes'),
      ])
      const listedNodes = clientList.clients ?? []
      const ids = listedNodes.map((node) => node.id ?? node.clientId ?? '').filter(Boolean)
      const status = ids.length > 0
        ? await post<StatusResponse>('/api/v1/platform/clientsstatus', { clientType: 1, clientIds: ids }).catch(() => ({} as StatusResponse))
        : {}
      setNodes(listedNodes.map((node) => {
        const id = node.id ?? node.clientId ?? ''
        return { ...node, online: status.clients?.[id]?.status === 1 }
      }))
      setServers(serverList.servers ?? [])
      setRoutes(Object.fromEntries((routeList.routes ?? []).map((route) => [route.nodeId, route.serverIds])))
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { void refresh(); void request<PlatformInfo>('/api/v1/platform/baseinfo').then(setPlatform).catch(() => undefined) }, [refresh])

  async function createEnrollment() {
    const result = await post<Enrollment>('/api/v2/enrollments', { nodeId })
    setInstall(result)
    setNodeId('')
  }

  async function addRoute(id: string) {
    const serverId = routeDrafts[id]
    if (!serverId) return
    setRouteError('')
    try { await post('/api/v2/node-routes', { nodeId: id, serverId }); await refresh() }
    catch (cause) { setRouteError(cause instanceof Error ? cause.message : String(cause)) }
  }

  async function removeRoute(node: string, server: string) {
    setRouteError('')
    try {
      await request('/api/v2/node-routes', { method: 'DELETE', body: JSON.stringify({ nodeId: node, serverId: server }) })
      await refresh()
    } catch (cause) { setRouteError(cause instanceof Error ? cause.message : String(cause)) }
  }

  const command = useMemo(() => {
    if (!install) return ''
    const api = platform.clientApiUrl || platform.client_api_url || window.location.origin
    const rpc = platform.clientRpcUrl || platform.client_rpc_url || `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`
    if (os === 'windows') {
      return `$p=Join-Path $env:TEMP 'frp-panel-install.ps1'; irm 'https://raw.githubusercontent.com/Onicc/frp-panel/main/install.ps1' -OutFile $p; & $p -AgentArguments @('--node-id',${psQuote(install.nodeId)},'--enrollment-token',${psQuote(install.token)},'--api-url',${psQuote(api)},'--rpc-url',${psQuote(rpc)}); Remove-Item $p -Force`
    }
    return `curl -fsSL https://raw.githubusercontent.com/Onicc/frp-panel/main/install.sh | sudo bash -s -- --node-id ${shellQuote(install.nodeId)} --enrollment-token ${shellQuote(install.token)} --api-url ${shellQuote(api)} --rpc-url ${shellQuote(rpc)}`
  }, [install, os, platform])

  return <>
    <header className="page-header"><div><span className="eyebrow">Infrastructure</span><h1>{t('nodes')}</h1><p>{t('nodeTopologyHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" onClick={() => setCreateOpen(true)}>＋ {t('createNode')}</button></div></header>
    {install && <section className="install-card"><div className="panel-heading"><div><h2>{t('installAgent')}</h2><p>{t('agentTokenHint')}</p></div><button className="icon-button" aria-label="Close" onClick={() => setInstall(null)}>×</button></div><div className="tabs">{(['linux', 'macos', 'windows'] as OS[]).map((item) => <button className={os === item ? 'active' : ''} onClick={() => setOS(item)} key={item}>{t(item)}</button>)}</div><pre><code>{command}</code></pre><button className="button secondary" onClick={async () => { await navigator.clipboard.writeText(command); setCopied(true); window.setTimeout(() => setCopied(false), 1500) }}>{copied ? t('copied') : t('copy')}</button></section>}
    {routeError && <p className="error route-error" role="alert">{routeError}</p>}
    <section className="panel table-panel"><table><thead><tr><th>{t('nodeId')}</th><th>{t('status')}</th><th>{t('lifecycle')}</th><th>{t('frpsRoutes')}</th><th>{t('addRoute')}</th></tr></thead><tbody>{nodes.map((node) => {
      const id = node.id ?? node.clientId ?? ''
      const assigned = routes[id] ?? []
      const available = servers.filter((server) => !assigned.includes(server.id ?? server.serverId ?? ''))
      return <tr key={id}><td><strong>{id}</strong></td><td><span className={node.online ? 'status online' : 'status offline'}>{node.online ? t('online') : t('offline')}</span></td><td>{node.ephemeral ? t('ephemeral') : t('managed')}</td><td><div className="route-list">{assigned.map((server) => <span className="route-chip" key={server}>{server}<button aria-label={`${t('removeRoute')} ${server}`} onClick={() => void removeRoute(id, server)}>×</button></span>)}{assigned.length === 0 && <span className="muted">{t('unassigned')}</span>}</div></td><td><div className="route-control"><select aria-label={t('frpsRoutes')} value={routeDrafts[id] ?? ''} onChange={(event) => setRouteDrafts((current) => ({ ...current, [id]: event.target.value }))}><option value="">{servers.length === 0 ? t('createServerFirst') : t('selectServer')}</option>{available.map((server) => { const value = server.id ?? server.serverId ?? ''; return <option value={value} key={value}>{value}</option> })}</select><button className="button secondary" disabled={!routeDrafts[id]} onClick={() => void addRoute(id)}>{t('add')}</button></div></td></tr>
    })}</tbody></table>{!loading && nodes.length === 0 && <div className="empty">{t('noData')}</div>}{loading && <div className="empty">{t('loading')}</div>}</section>
    <Modal open={createOpen} onOpenChange={setCreateOpen} title={t('createNode')} onSubmit={createEnrollment}><label>{t('nodeId')}<input value={nodeId} onChange={(event) => setNodeId(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label><p className="field-help">{t('nodeCreateHint')}</p></Modal>
  </>
}
