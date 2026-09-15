import { useCallback, useEffect, useMemo, useState } from 'react'
import { post, request } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Node = { id?: string; clientId?: string; stopped?: boolean; ephemeral?: boolean }
type ClientList = { clients?: Node[] }
type PlatformInfo = { clientApiUrl?: string; clientRpcUrl?: string }
type Enrollment = { nodeId: string; token: string; expiresAt: string }
type OS = 'linux' | 'macos' | 'windows'

function shellQuote(value: string): string { return `'${value.replaceAll("'", `'"'"'`)}'` }
function psQuote(value: string): string { return `'${value.replaceAll("'", "''")}'` }

export default function Nodes() {
  const { t } = useI18n()
  const [nodes, setNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [nodeId, setNodeId] = useState('')
  const [install, setInstall] = useState<Enrollment | null>(null)
  const [platform, setPlatform] = useState<PlatformInfo>({})
  const [os, setOS] = useState<OS>('linux')
  const [copied, setCopied] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    try { setNodes((await post<ClientList>('/api/v1/client/list', { page: 1, pageSize: 100 })).clients ?? []) } finally { setLoading(false) }
  }, [])
  useEffect(() => { void refresh(); void request<PlatformInfo>('/api/v1/platform/baseinfo').then(setPlatform).catch(() => undefined) }, [refresh])

  async function createEnrollment() {
    const result = await post<Enrollment>('/api/v2/enrollments', { nodeId })
    setInstall(result)
    setNodeId('')
  }

  const command = useMemo(() => {
    if (!install) return ''
    const api = platform.clientApiUrl || window.location.origin
    const rpc = platform.clientRpcUrl || `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`
    if (os === 'windows') {
      return `$p=Join-Path $env:TEMP 'frp-panel-install.ps1'; irm 'https://raw.githubusercontent.com/Onicc/frp-panel/main/install.ps1' -OutFile $p; & $p -AgentArguments @('--node-id',${psQuote(install.nodeId)},'--enrollment-token',${psQuote(install.token)},'--api-url',${psQuote(api)},'--rpc-url',${psQuote(rpc)}); Remove-Item $p -Force`
    }
    return `curl -fsSL https://raw.githubusercontent.com/Onicc/frp-panel/main/install.sh | sudo bash -s -- --node-id ${shellQuote(install.nodeId)} --enrollment-token ${shellQuote(install.token)} --api-url ${shellQuote(api)} --rpc-url ${shellQuote(rpc)}`
  }, [install, os, platform])

  return <>
    <header className="page-header"><div><span className="eyebrow">Infrastructure</span><h1>{t('nodes')}</h1><p>{t('capabilityHint')}</p></div><div className="header-actions"><button className="button secondary" onClick={() => void refresh()}>{t('refresh')}</button><button className="button primary" onClick={() => setCreateOpen(true)}>＋ {t('createNode')}</button></div></header>
    {install && <section className="install-card"><div className="panel-heading"><div><h2>{t('installAgent')}</h2><p>Enrollment token expires in 10 minutes.</p></div><button className="icon-button" onClick={() => setInstall(null)}>×</button></div><div className="tabs">{(['linux', 'macos', 'windows'] as OS[]).map((item) => <button className={os === item ? 'active' : ''} onClick={() => setOS(item)} key={item}>{t(item)}</button>)}</div><pre><code>{command}</code></pre><button className="button secondary" onClick={async () => { await navigator.clipboard.writeText(command); setCopied(true); window.setTimeout(() => setCopied(false), 1500) }}>{copied ? t('copied') : t('copy')}</button></section>}
    <section className="panel table-panel"><table><thead><tr><th>{t('nodeId')}</th><th>Status</th><th>Lifecycle</th></tr></thead><tbody>{nodes.map((node) => <tr key={node.id ?? node.clientId}><td><strong>{node.id ?? node.clientId}</strong></td><td><span className={node.stopped ? 'status offline' : 'status online'}>{node.stopped ? t('offline') : t('online')}</span></td><td>{node.ephemeral ? 'Ephemeral' : 'Managed'}</td></tr>)}</tbody></table>{!loading && nodes.length === 0 && <div className="empty">{t('noData')}</div>}{loading && <div className="empty">Loading…</div>}</section>
    <Modal open={createOpen} onOpenChange={setCreateOpen} title={t('createNode')} onSubmit={createEnrollment}><label>{t('nodeId')}<input value={nodeId} onChange={(event) => setNodeId(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label><p className="field-help">A short-lived enrollment token is generated; the permanent credential is written only to the protected agent config.</p></Modal>
  </>
}
