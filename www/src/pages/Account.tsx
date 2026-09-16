import { type FormEvent, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearSession, post, request } from '../api'
import { useI18n } from '../i18n'

type AccountInfo = { username: string; email: string; role: string }

export default function Account() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [account, setAccount] = useState<AccountInfo | null>(null)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    void request<AccountInfo>('/api/v2/account').then(setAccount).catch((cause) => setError(cause instanceof Error ? cause.message : String(cause)))
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    if (newPassword !== confirmPassword) {
      setError(t('passwordMismatch'))
      return
    }
    setPending(true)
    try {
      await post('/api/v2/account/password', { currentPassword, newPassword })
      localStorage.setItem('frp-panel.password-changed', '1')
      navigate('/login', { replace: true })
      clearSession()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally { setPending(false) }
  }

  return <>
    <header className="page-header"><div><span className="eyebrow">{t('security')}</span><h1>{t('account')}</h1><p>{t('accountHint')}</p></div></header>
    <div className="settings-grid">
      <section className="panel profile-card"><div className="panel-heading"><div><h2>{t('profile')}</h2><p>{t('profileHint')}</p></div><span className="profile-avatar">{account?.username?.slice(0, 1).toUpperCase() || 'U'}</span></div><dl><div><dt>{t('username')}</dt><dd>{account?.username || '—'}</dd></div><div><dt>{t('email')}</dt><dd>{account?.email || '—'}</dd></div><div><dt>{t('role')}</dt><dd><span className="role-badge">{account?.role || '—'}</span></dd></div></dl></section>
      <section className="panel password-card"><div className="panel-heading"><div><h2>{t('changePassword')}</h2><p>{t('changePasswordHint')}</p></div></div><form onSubmit={submit} autoComplete="on"><label htmlFor="account-current-password">{t('currentPassword')}<input id="account-current-password" name="current-password" type="password" autoComplete="current-password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} required /></label><label htmlFor="account-new-password">{t('newPassword')}<input id="account-new-password" name="new-password" type="password" autoComplete="new-password" minLength={12} value={newPassword} onChange={(event) => setNewPassword(event.target.value)} required /></label><label htmlFor="account-confirm-password">{t('confirmPassword')}<input id="account-confirm-password" name="new-password-confirmation" type="password" autoComplete="new-password" minLength={12} value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} required /></label><p className="field-help">{t('passwordPolicy')}</p>{error && <p className="error alert" role="alert">{error}</p>}<button className="button primary" disabled={pending}>{pending ? t('saving') : t('updatePassword')}</button></form></section>
    </div>
  </>
}
