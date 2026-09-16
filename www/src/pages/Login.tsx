import { type FormEvent, useEffect, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { login, register, request, sessionToken } from '../api'
import { useI18n } from '../i18n'

type BootstrapStatus = { registrationEnabled: boolean; ownerExists: boolean; canCreateOwner: boolean }

export default function Login() {
  const { t, language, setLanguage } = useI18n()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [email, setEmail] = useState('')
  const [bootstrap, setBootstrap] = useState(false)
  const [bootstrapStatus, setBootstrapStatus] = useState<BootstrapStatus | null>(null)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    void request<BootstrapStatus>('/api/v2/bootstrap-status')
      .then((status) => {
        setBootstrapStatus(status)
        if (status.canCreateOwner) setBootstrap(true)
      })
      .catch(() => setBootstrapStatus(null))
  }, [])

  if (sessionToken()) return <Navigate to="/" replace />

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      if (bootstrap && password !== confirmPassword) throw new Error(t('passwordMismatch'))
      if (bootstrap) await register(username, email, password)
      await login(username, password)
      navigate('/', { replace: true })
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setPending(false)
    }
  }
  const usernameField = bootstrap ? 'new-username' : 'username'
  const passwordField = bootstrap ? 'new-password' : 'current-password'
  return <main className="login-page">
    <section className="login-card">
      <div className="brand login-brand"><span className="brand-mark">fp</span><div><strong>frp-panel</strong><small>v2 edge</small></div></div>
      <span className="auth-mode">{bootstrap ? t('initialSetup') : t('accountLogin')}</span>
      <h1>{bootstrap ? t('setupTitle') : t('welcome')}</h1>
      <p>{bootstrap ? t('setupHint') : t('loginHint')}</p>
      <form id={bootstrap ? 'owner-setup-form' : 'login-form'} name={bootstrap ? 'owner-setup' : 'login'} onSubmit={submit} autoComplete="on" data-auth-mode={bootstrap ? 'register' : 'login'}>
        <label htmlFor={usernameField}>{bootstrap ? t('username') : t('usernameOrEmail')}<input id={usernameField} name="username" autoComplete="username" pattern={bootstrap ? '[A-Za-z0-9_-]+' : undefined} value={username} onChange={(event) => setUsername(event.target.value)} required autoFocus /></label>
        {bootstrap && <label htmlFor="email">{t('email')}<input id="email" name="email" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required /></label>}
        <label htmlFor={passwordField}>{t('password')}<input id={passwordField} name={passwordField} type="password" minLength={bootstrap ? 12 : undefined} autoComplete={passwordField} value={password} onChange={(event) => setPassword(event.target.value)} required /></label>
        {bootstrap && <label htmlFor="new-password-confirmation">{t('confirmPassword')}<input id="new-password-confirmation" name="new-password-confirmation" type="password" minLength={12} autoComplete="new-password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} required /></label>}
        {error && <p className="error" role="alert">{error}</p>}
        <button type="submit" className="button primary" disabled={pending}>{pending ? t('saving') : (bootstrap ? t('createOwner') : t('login'))}</button>
      </form>
      {bootstrapStatus?.canCreateOwner && <button type="button" className="text-button login-mode" onClick={() => { setBootstrap(!bootstrap); setConfirmPassword(''); setError('') }}>{bootstrap ? t('backToLogin') : t('bootstrapOwner')}</button>}
      {bootstrapStatus && !bootstrapStatus.ownerExists && !bootstrapStatus.registrationEnabled && <p className="setup-warning" role="status">{t('setupDisabled')}</p>}
      <div className="language-switch"><button className={language === 'zh' ? 'active' : ''} onClick={() => setLanguage('zh')}>中文</button><button className={language === 'en' ? 'active' : ''} onClick={() => setLanguage('en')}>English</button></div>
    </section>
  </main>
}
