import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Client = { id?: string; clientId?: string; stopped?: boolean; ephemeral?: boolean; online?: boolean; lastSeenAt?: number }
type ClientList = { clients?: Client[] }
type PlatformInfo = { clientApiUrl?: string; clientRpcUrl?: string; client_api_url?: string; client_rpc_url?: string }
type Enrollment = { clientId: string; token: string; expiresAt: string }
type StatusResponse = { clients?: Record<string, { status?: number }> }
type OS = 'linux' | 'macos' | 'windows'

function shellQuote(value: string): string { return `'${value.replaceAll("'", `'"'"'`)}'` }
function psQuote(value: string): string { return `'${value.replaceAll("'", "''")}'` }

export default function Clients() {
  const { t, language } = useI18n()
  const [clients, setClients] = useState<Client[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [clientId, setClientId] = useState('')
  const [install, setInstall] = useState<Enrollment | null>(null)
  const [platform, setPlatform] = useState<PlatformInfo>({})
  const [os, setOS] = useState<OS>('linux')
  const [copied, setCopied] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const listed = (await post<ClientList>('/api/v1/client/list', { page: 1, pageSize: 100 })).clients ?? []
      const ids = listed.map((client) => client.id ?? client.clientId ?? '').filter(Boolean)
      const status = ids.length > 0
        ? await post<StatusResponse>('/api/v1/platform/clientsstatus', { clientType: 1, clientIds: ids }).catch(() => ({} as StatusResponse))
        : {}
      setClients(listed.map((client) => {
        const id = client.id ?? client.clientId ?? ''
        return { ...client, online: status.clients?.[id]?.status === 1 }
      }))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally { setLoading(false) }
  }, [])

  useEffect(() => {
    void refresh()
    void request<PlatformInfo>('/api/v1/platform/baseinfo').then(setPlatform).catch(() => undefined)
  }, [refresh])

  async function createEnrollment() {
    const result = await post<Enrollment>('/api/v2/enrollments', { clientId })
    setInstall(result)
    setClients((current) => current.some((client) => (client.id ?? client.clientId) === result.clientId)
      ? current
      : [...current, { id: result.clientId, online: false }])
    setClientId('')
  }

  const command = useMemo(() => {
    if (!install) return ''
    const api = platform.clientApiUrl || platform.client_api_url || window.location.origin
    const rpc = platform.clientRpcUrl || platform.client_rpc_url || `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`
    if (os === 'windows') {
      return `$p=Join-Path $env:TEMP 'frp-panel-install.ps1'; irm 'https://raw.githubusercontent.com/Onicc/frp-panel/main/install.ps1' -OutFile $p; & $p -AgentArguments @('--client-id',${psQuote(install.clientId)},'--enrollment-token',${psQuote(install.token)},'--api-url',${psQuote(api)},'--rpc-url',${psQuote(rpc)}); Remove-Item $p -Force`
    }
    return `curl -fsSL https://raw.githubusercontent.com/Onicc/frp-panel/main/install.sh | sudo bash -s -- --client-id ${shellQuote(install.clientId)} --enrollment-token ${shellQuote(install.token)} --api-url ${shellQuote(api)} --rpc-url ${shellQuote(rpc)}`
  }, [install, os, platform])

  function formatLastSeen(value?: number): string {
    if (!value) return '—'
    return new Intl.DateTimeFormat(language === 'zh' ? 'zh-CN' : 'en', { dateStyle: 'medium', timeStyle: 'short' }).format(value)
  }

  return <>
    <header className="page-header"><div><span className="eyebrow">{t('infrastructure')}</span><h1>{t('clients')}</h1><p>{t('clientTopologyHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" onClick={() => setCreateOpen(true)}>＋ {t('createClient')}</button></div></header>
    {install && <section className="install-card" aria-live="polite"><div className="panel-heading"><div><span className="step-badge">{t('nextStep')}</span><h2>{t('installAgent')}</h2><p>{t('agentTokenHint')}</p></div><button className="icon-button" aria-label={t('close')} onClick={() => setInstall(null)}>×</button></div><div className="tabs" role="tablist" aria-label={t('operatingSystem')}>{(['linux', 'macos', 'windows'] as OS[]).map((item) => <button role="tab" aria-selected={os === item} className={os === item ? 'active' : ''} onClick={() => setOS(item)} key={item}>{t(item)}</button>)}</div><pre><code>{command}</code></pre><div className="install-actions"><button className="button secondary" onClick={async () => { await navigator.clipboard.writeText(command); setCopied(true); window.setTimeout(() => setCopied(false), 1500) }}>{copied ? t('copied') : t('copy')}</button><small>{t('installCommandHint')}</small></div></section>}
    {error && <p className="error alert" role="alert">{error}</p>}
    <section className="panel table-panel"><div className="table-title"><div><h2>{t('clientInventory')}</h2><p>{t('clientInventoryHint')}</p></div><span className="count-badge">{clients.length}</span></div><div className="table-scroll"><table><thead><tr><th>{t('clientId')}</th><th>{t('agentStatus')}</th><th>{t('lifecycle')}</th><th>{t('lastSeen')}</th></tr></thead><tbody>{clients.map((client) => {
      const id = client.id ?? client.clientId ?? ''
      return <tr key={id}><td><div className="resource-name"><span className="resource-icon">C</span><strong>{id}</strong></div></td><td><span className={client.online ? 'status online' : 'status offline'}><i />{client.online ? t('online') : t('offline')}</span></td><td>{client.ephemeral ? t('ephemeral') : t('managed')}</td><td className="muted">{formatLastSeen(client.lastSeenAt)}</td></tr>
    })}</tbody></table></div>{!loading && clients.length === 0 && <div className="empty"><strong>{t('noClients')}</strong><span>{t('noClientsHint')}</span></div>}{loading && <div className="empty">{t('loading')}</div>}</section>
    <Modal open={createOpen} onOpenChange={setCreateOpen} title={t('createClient')} onSubmit={createEnrollment}><label>{t('clientId')}<input value={clientId} onChange={(event) => setClientId(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label><p className="field-help">{t('clientCreateHint')}</p></Modal>
  </>
}
