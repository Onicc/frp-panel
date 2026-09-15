import { useCallback, useEffect, useState } from 'react'
import { post } from '../api'
import { Modal } from '../components/Modal'
import { useI18n } from '../i18n'

type Server = { id?: string; serverId?: string; ip?: string; stopped?: boolean }

export default function Servers() {
  const { t } = useI18n()
  const [servers, setServers] = useState<Server[]>([])
  const [open, setOpen] = useState(false)
  const [serverId, setServerId] = useState('')
  const [serverIp, setServerIp] = useState('')
  const refresh = useCallback(async () => setServers((await post<{ servers?: Server[] }>('/api/v1/server/list', { page: 1, pageSize: 100 })).servers ?? []), [])
  useEffect(() => { void refresh() }, [refresh])
  async function create() {
    const response = await post<{ status?: { code?: number; message?: string } }>('/api/v1/server/init', { serverId, serverIp })
    if (response.status?.code !== undefined && response.status.code !== 0) throw new Error(response.status.message || t('error'))
    setServerId(''); setServerIp(''); await refresh()
  }
  return <>
    <header className="page-header"><div><span className="eyebrow">Ingress</span><h1>{t('servers')}</h1><p>Built-in FRPS stays available as the default server.</p></div><button className="button primary" onClick={() => setOpen(true)}>＋ {t('createServer')}</button></header>
    <section className="panel table-panel"><table><thead><tr><th>{t('serverId')}</th><th>{t('serverIp')}</th><th>Status</th></tr></thead><tbody>{servers.map((server) => <tr key={server.id ?? server.serverId}><td><strong>{server.id ?? server.serverId}</strong></td><td>{server.ip || '—'}</td><td><span className={server.stopped ? 'status offline' : 'status online'}>{server.stopped ? t('offline') : t('online')}</span></td></tr>)}</tbody></table>{servers.length === 0 && <div className="empty">{t('noData')}</div>}</section>
    <Modal open={open} onOpenChange={setOpen} title={t('createServer')} onSubmit={create}><label>{t('serverId')}<input value={serverId} onChange={(event) => setServerId(event.target.value)} pattern="[A-Za-z0-9_-]+" required autoFocus /></label><label>{t('serverIp')}<input value={serverIp} onChange={(event) => setServerIp(event.target.value)} required /></label></Modal>
  </>
}
