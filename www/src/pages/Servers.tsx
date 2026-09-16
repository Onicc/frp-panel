import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Server = { id?: string; serverId?: string; ip?: string; online?: boolean }
type PlatformInfo = { clientApiUrl?: string; clientRpcUrl?: string; client_api_url?: string; client_rpc_url?: string }
type Enrollment = { serverId: string; token: string; expiresAt: string }
type StatusResponse = { clients?: Record<string, { status?: number }> }

function yamlString(value: string): string { return JSON.stringify(value) }

export function buildServerCompose(enrollment: Enrollment, publicURL: string): string {
  return `services:
  frps:
    image: onicc/frp-panel:edge
    restart: unless-stopped
    network_mode: host
    command:
      - server
      - --config
      - /data/server.yaml
    environment:
      PUBLIC_URL: ${yamlString(publicURL)}
      SERVER_ENROLLMENT_TOKEN: ${yamlString(enrollment.token)}
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
`
}

export default function Servers() {
  const { t } = useI18n()
  const [servers, setServers] = useState<Server[]>([])
  const [open, setOpen] = useState(false)
  const [serverId, setServerId] = useState('')
  const [serverIp, setServerIp] = useState('')
  const [bindPort, setBindPort] = useState(7000)
  const [deployment, setDeployment] = useState<Enrollment | null>(null)
  const [platform, setPlatform] = useState<PlatformInfo>({})
  const [copied, setCopied] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const refresh = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const listed = (await post<{ servers?: Server[] }>('/api/v1/server/list', { page: 1, pageSize: 100 })).servers ?? []
      const ids = listed.map((server) => server.id ?? server.serverId ?? '').filter(Boolean)
      if (ids.length === 0) { setServers(listed); return }
      const status = await post<StatusResponse>('/api/v1/platform/clientsstatus', { clientType: 2, clientIds: ids }).catch(() => ({} as StatusResponse))
      setServers(listed.map((server) => {
        const id = server.id ?? server.serverId ?? ''
        return { ...server, online: status.clients?.[id]?.status === 1 }
      }))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
    void request<PlatformInfo>('/api/v1/platform/baseinfo').then(setPlatform).catch(() => undefined)
  }, [refresh])

  async function create() {
    const result = await post<Enrollment>('/api/v2/server-enrollments', { serverId, serverIp, bindPort })
    setDeployment(result)
    setServers((current) => current.some((server) => (server.id ?? server.serverId) === result.serverId)
      ? current
      : [...current, { id: result.serverId, ip: serverIp, online: false }])
    setServerId('')
    setServerIp('')
    setBindPort(7000)
  }

  const compose = useMemo(() => {
    if (!deployment) return ''
    const publicURL = platform.clientApiUrl || platform.client_api_url || window.location.origin
    return buildServerCompose(deployment, publicURL)
  }, [deployment, platform])

  return <>
    <header className="page-header"><div><span className="eyebrow">Data plane</span><h1>{t('servers')}</h1><p>{t('serverTopologyHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" onClick={() => setOpen(true)}>＋ {t('createServer')}</button></div></header>
    {deployment && <section className="install-card"><div className="panel-heading"><div><h2>{t('deployServer')}</h2><p>{t('serverTokenHint')}</p></div><button className="icon-button" aria-label="Close" onClick={() => setDeployment(null)}>×</button></div><pre><code>{compose}</code></pre><button className="button secondary" onClick={async () => { await navigator.clipboard.writeText(compose); setCopied(true); window.setTimeout(() => setCopied(false), 1500) }}>{copied ? t('copied') : t('copyCompose')}</button></section>}
    {error && <p className="error route-error" role="alert">{error}</p>}
    <section className="panel table-panel"><table><thead><tr><th>{t('serverId')}</th><th>{t('serverIp')}</th><th>{t('status')}</th></tr></thead><tbody>{servers.map((server) => <tr key={server.id ?? server.serverId}><td><strong>{server.id ?? server.serverId}</strong></td><td>{server.ip || '—'}</td><td><span className={server.online ? 'status online' : 'status offline'}>{server.online ? t('online') : t('offline')}</span></td></tr>)}</tbody></table>{loading && <div className="empty">{t('loading')}</div>}{!loading && servers.length === 0 && <div className="empty">{t('noData')}</div>}</section>
    <Modal open={open} onOpenChange={setOpen} title={t('createServer')} submitLabel={t('createDeployment')} onSubmit={create}><label>{t('serverId')}<input value={serverId} onChange={(event) => setServerId(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label><label>{t('serverIp')}<input value={serverIp} onChange={(event) => setServerIp(event.target.value)} placeholder="frps.example.com" required /></label><label>{t('bindPort')}<input type="number" min="1024" max="65535" value={bindPort} onChange={(event) => setBindPort(Number(event.target.value))} required /></label><p className="field-help">{t('serverCreateHint')}</p></Modal>
  </>
}
