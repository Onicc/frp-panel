import { type FormEvent, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { login, register, sessionToken } from '../api'
import { useI18n } from '../i18n'

export default function Login() {
  const { t, language, setLanguage } = useI18n()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
	const [email, setEmail] = useState('')
	const [bootstrap, setBootstrap] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  if (sessionToken()) return <Navigate to="/" replace />

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
	  if (bootstrap) await register(username, email, password)
      await login(username, password)
      navigate('/')
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setPending(false)
    }
  }
  return <main className="login-page">
    <section className="login-card">
      <div className="brand login-brand"><span className="brand-mark">fp</span><div><strong>frp-panel</strong><small>v2 edge</small></div></div>
      <h1>{t('welcome')}</h1>
      <p>{t('capabilityHint')}</p>
      <form onSubmit={submit}>
        <label>{t('username')}<input autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} required /></label>
		{bootstrap && <label>{t('email')}<input type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required /></label>}
        <label>{t('password')}<input type="password" minLength={bootstrap ? 12 : undefined} autoComplete={bootstrap ? 'new-password' : 'current-password'} value={password} onChange={(event) => setPassword(event.target.value)} required /></label>
        {error && <p className="error" role="alert">{error}</p>}
		<button className="button primary" disabled={pending}>{pending ? t('saving') : (bootstrap ? t('createOwner') : t('login'))}</button>
      </form>
	  <button className="text-button login-mode" onClick={() => { setBootstrap(!bootstrap); setError('') }}>{bootstrap ? t('backToLogin') : t('bootstrapOwner')}</button>
      <div className="language-switch"><button className={language === 'zh' ? 'active' : ''} onClick={() => setLanguage('zh')}>中文</button><button className={language === 'en' ? 'active' : ''} onClick={() => setLanguage('en')}>English</button></div>
    </section>
  </main>
}
