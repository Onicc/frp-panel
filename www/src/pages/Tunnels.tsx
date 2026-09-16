import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Node = { id?: string; clientId?: string }
type Server = { id?: string; serverId?: string; ip?: string }
type Route = { nodeId: string; serverIds: string[] }
type Tunnel = {
  name: string
  nodeId: string
  clientId: string
  serverId: string
  type: 'tcp' | 'udp'
  localHost: string
  localPort: number
  remotePort: number
  stopped: boolean
}

function endpoint(host: string, port: number): string {
  return `${host.includes(':') && !host.startsWith('[') ? `[${host}]` : host}:${port}`
}

export default function Tunnels() {
  const { t } = useI18n()
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [nodes, setNodes] = useState<Node[]>([])
  const [servers, setServers] = useState<Server[]>([])
  const [routes, setRoutes] = useState<Record<string, string[]>>({})
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [nodeId, setNodeId] = useState('')
  const [serverId, setServerId] = useState('')
  const [type, setType] = useState<'tcp' | 'udp'>('tcp')
  const [localHost, setLocalHost] = useState('127.0.0.1')
  const [localPort, setLocalPort] = useState(80)
  const [remotePort, setRemotePort] = useState(8080)
  const [error, setError] = useState('')

  const refresh = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [tunnelList, clientList, serverList, routeList] = await Promise.all([
        request<{ tunnels?: Tunnel[] }>('/api/v2/tunnels'),
        post<{ clients?: Node[] }>('/api/v1/client/list', { page: 1, pageSize: 100 }),
        post<{ servers?: Server[] }>('/api/v1/server/list', { page: 1, pageSize: 100 }),
        request<{ routes?: Route[] }>('/api/v2/node-routes'),
      ])
      setTunnels(tunnelList.tunnels ?? [])
      setNodes(clientList.clients ?? [])
      setServers(serverList.servers ?? [])
      setRoutes(Object.fromEntries((routeList.routes ?? []).map((route) => [route.nodeId, route.serverIds])))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const routedServers = useMemo(() => routes[nodeId] ?? [], [nodeId, routes])
  const serverAddresses = useMemo(() => Object.fromEntries(servers.map((server) => [server.id ?? server.serverId ?? '', server.ip ?? ''])), [servers])
  const routableNodes = useMemo(() => nodes.filter((node) => (routes[node.id ?? node.clientId ?? ''] ?? []).length > 0), [nodes, routes])

  async function create() {
    const created = await post<Tunnel>('/api/v2/tunnels', { name, nodeId, serverId, type, localHost, localPort, remotePort })
    setTunnels((current) => [...current.filter((tunnel) => tunnel.nodeId !== created.nodeId || tunnel.name !== created.name), created])
    setName('')
    setLocalHost('127.0.0.1')
    setLocalPort(80)
    setRemotePort(8080)
  }

  async function remove(tunnel: Tunnel) {
    if (!window.confirm(t('deleteTunnelConfirm'))) return
    setError('')
    try {
      await request('/api/v2/tunnels', { method: 'DELETE', body: JSON.stringify({ name: tunnel.name, nodeId: tunnel.nodeId, serverId: tunnel.serverId }) })
      setTunnels((current) => current.filter((item) => item.nodeId !== tunnel.nodeId || item.name !== tunnel.name))
    } catch (cause) { setError(cause instanceof Error ? cause.message : String(cause)) }
  }

  function beginCreate() {
    const firstNode = routableNodes[0]?.id ?? routableNodes[0]?.clientId ?? ''
    setNodeId(firstNode)
    setServerId((routes[firstNode] ?? [])[0] ?? '')
    setError('')
    setOpen(true)
  }

  return <>
    <header className="page-header"><div><span className="eyebrow">Data plane</span><h1>{t('tunnels')}</h1><p>{t('tunnelTopologyHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" disabled={routableNodes.length === 0} onClick={beginCreate}>＋ {t('createTunnel')}</button></div></header>
    {routableNodes.length === 0 && !loading && <section className="notice-card"><strong>{t('routeRequired')}</strong><p>{t('routeRequiredHint')}</p></section>}
    {error && <p className="error route-error" role="alert">{error}</p>}
    <section className="panel table-panel"><table><thead><tr><th>{t('tunnelName')}</th><th>{t('protocol')}</th><th>{t('node')}</th><th>{t('server')}</th><th>{t('localService')}</th><th>{t('publicEndpoint')}</th><th>{t('actions')}</th></tr></thead><tbody>{tunnels.map((tunnel) => <tr key={`${tunnel.nodeId}-${tunnel.name}`}><td><strong>{tunnel.name}</strong></td><td><span className="protocol-badge">{tunnel.type.toUpperCase()}</span></td><td>{tunnel.nodeId}</td><td>{tunnel.serverId}</td><td><code>{endpoint(tunnel.localHost, tunnel.localPort)}</code></td><td><code>{endpoint(serverAddresses[tunnel.serverId] || tunnel.serverId, tunnel.remotePort)}</code></td><td><button className="button danger" onClick={() => void remove(tunnel)}>{t('delete')}</button></td></tr>)}</tbody></table>{loading && <div className="empty">{t('loading')}</div>}{!loading && tunnels.length === 0 && <div className="empty">{t('noTunnels')}</div>}</section>
    <Modal open={open} onOpenChange={setOpen} title={t('createTunnel')} onSubmit={create}>
      <label>{t('tunnelName')}<input value={name} onChange={(event) => setName(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label>
      <label>{t('node')}<select value={nodeId} onChange={(event) => { const value = event.target.value; setNodeId(value); setServerId((routes[value] ?? [])[0] ?? '') }} required>{routableNodes.map((node) => { const id = node.id ?? node.clientId ?? ''; return <option value={id} key={id}>{id}</option> })}</select></label>
      <label>{t('server')}<select value={serverId} onChange={(event) => setServerId(event.target.value)} required>{routedServers.map((id) => <option value={id} key={id}>{id}</option>)}</select></label>
      <label>{t('protocol')}<select value={type} onChange={(event) => setType(event.target.value as 'tcp' | 'udp')}><option value="tcp">TCP</option><option value="udp">UDP</option></select></label>
      <div className="form-grid"><label>{t('localHost')}<input value={localHost} onChange={(event) => setLocalHost(event.target.value)} required /></label><label>{t('localPort')}<input type="number" min="1" max="65535" value={localPort} onChange={(event) => setLocalPort(Number(event.target.value))} required /></label></div>
      <label>{t('remotePort')}<input type="number" min="1024" max="65535" value={remotePort} onChange={(event) => setRemotePort(Number(event.target.value))} required /></label>
      <p className="field-help">{t('tunnelCreateHint')}</p>
    </Modal>
  </>
}
