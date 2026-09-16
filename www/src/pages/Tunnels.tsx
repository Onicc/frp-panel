import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Client = { id?: string; clientId?: string }
type Server = { id?: string; serverId?: string; ip?: string }
type Tunnel = {
  name: string
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
  const [clients, setClients] = useState<Client[]>([])
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [clientId, setClientId] = useState('')
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
      const [tunnelList, clientList, serverList] = await Promise.all([
        request<{ tunnels?: Tunnel[] }>('/api/v2/tunnels'),
        post<{ clients?: Client[] }>('/api/v1/client/list', { page: 1, pageSize: 100 }),
        post<{ servers?: Server[] }>('/api/v1/server/list', { page: 1, pageSize: 100 }),
      ])
      setTunnels(tunnelList.tunnels ?? [])
      setClients(clientList.clients ?? [])
      setServers(serverList.servers ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const serverAddresses = useMemo(() => Object.fromEntries(servers.map((server) => [server.id ?? server.serverId ?? '', server.ip ?? ''])), [servers])
  const canCreate = clients.length > 0 && servers.length > 0

  async function create() {
    const created = await post<Tunnel>('/api/v2/tunnels', { name, clientId, serverId, type, localHost, localPort, remotePort })
    setTunnels((current) => [...current.filter((tunnel) => tunnel.clientId !== created.clientId || tunnel.name !== created.name), created])
    setName('')
    setLocalHost('127.0.0.1')
    setLocalPort(80)
    setRemotePort(8080)
  }

  async function remove(tunnel: Tunnel) {
    if (!window.confirm(t('deleteTunnelConfirm'))) return
    setError('')
    try {
      await request('/api/v2/tunnels', { method: 'DELETE', body: JSON.stringify({ name: tunnel.name, clientId: tunnel.clientId, serverId: tunnel.serverId }) })
      setTunnels((current) => current.filter((item) => item.clientId !== tunnel.clientId || item.name !== tunnel.name))
    } catch (cause) { setError(cause instanceof Error ? cause.message : String(cause)) }
  }

  function beginCreate() {
    setClientId(clients[0]?.id ?? clients[0]?.clientId ?? '')
    setServerId(servers[0]?.id ?? servers[0]?.serverId ?? '')
    setError('')
    setOpen(true)
  }

  return <>
    <header className="page-header"><div><span className="eyebrow">{t('dataPlane')}</span><h1>{t('tunnels')}</h1><p>{t('tunnelTopologyHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" disabled={!canCreate} onClick={beginCreate}>＋ {t('createTunnel')}</button></div></header>
    {!canCreate && !loading && <section className="notice-card"><strong>{t('tunnelPrerequisites')}</strong><p>{clients.length === 0 ? t('createClientFirst') : t('createServerFirst')}</p></section>}
    {error && <p className="error alert" role="alert">{error}</p>}
    <section className="panel table-panel"><div className="table-title"><div><h2>{t('tunnelInventory')}</h2><p>{t('tunnelInventoryHint')}</p></div><span className="count-badge">{tunnels.length}</span></div><div className="table-scroll"><table><thead><tr><th>{t('tunnelName')}</th><th>{t('protocol')}</th><th>{t('client')}</th><th>{t('server')}</th><th>{t('localService')}</th><th>{t('publicEndpoint')}</th><th>{t('actions')}</th></tr></thead><tbody>{tunnels.map((tunnel) => <tr key={`${tunnel.clientId}-${tunnel.name}`}><td><div className="resource-name"><span className="resource-icon tunnel-icon">T</span><strong>{tunnel.name}</strong></div></td><td><span className="protocol-badge">{tunnel.type.toUpperCase()}</span></td><td>{tunnel.clientId}</td><td>{tunnel.serverId}</td><td><code>{endpoint(tunnel.localHost, tunnel.localPort)}</code></td><td><code>{endpoint(serverAddresses[tunnel.serverId] || tunnel.serverId, tunnel.remotePort)}</code></td><td><button className="button danger" onClick={() => void remove(tunnel)}>{t('delete')}</button></td></tr>)}</tbody></table></div>{loading && <div className="empty">{t('loading')}</div>}{!loading && tunnels.length === 0 && <div className="empty"><strong>{t('noTunnels')}</strong><span>{t('noTunnelsHint')}</span></div>}</section>
    <Modal open={open} onOpenChange={setOpen} title={t('createTunnel')} onSubmit={create}>
      <label>{t('tunnelName')}<input value={name} onChange={(event) => setName(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label>
      <div className="form-grid"><label>{t('client')}<select value={clientId} onChange={(event) => setClientId(event.target.value)} required>{clients.map((client) => { const id = client.id ?? client.clientId ?? ''; return <option value={id} key={id}>{id}</option> })}</select></label><label>{t('server')}<select value={serverId} onChange={(event) => setServerId(event.target.value)} required>{servers.map((server) => { const id = server.id ?? server.serverId ?? ''; return <option value={id} key={id}>{id}</option> })}</select></label></div>
      <label>{t('protocol')}<select value={type} onChange={(event) => setType(event.target.value as 'tcp' | 'udp')}><option value="tcp">TCP</option><option value="udp">UDP</option></select></label>
      <div className="form-grid"><label>{t('localHost')}<input value={localHost} onChange={(event) => setLocalHost(event.target.value)} required /></label><label>{t('localPort')}<input type="number" min="1" max="65535" value={localPort} onChange={(event) => setLocalPort(Number(event.target.value))} required /></label></div>
      <label>{t('remotePort')}<input type="number" min="1024" max="65535" value={remotePort} onChange={(event) => setRemotePort(Number(event.target.value))} required /></label>
      <p className="field-help">{t('tunnelCreateHint')}</p>
    </Modal>
  </>
}
