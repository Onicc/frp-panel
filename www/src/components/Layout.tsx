import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { logout } from '../api'
import { useI18n } from '../i18n'

const nav = [
  { to: '/', key: 'overview', icon: '⌂' },
  { to: '/clients', key: 'clients', icon: 'C' },
  { to: '/servers', key: 'servers', icon: 'S' },
  { to: '/tunnels', key: 'tunnels', icon: 'T' },
] as const

export function Layout() {
  const { t, language, setLanguage } = useI18n()
  const navigate = useNavigate()
  const location = useLocation()
  const section = nav.find((item) => item.to === location.pathname)?.key ?? (location.pathname === '/account' ? 'account' : 'overview')

  return <div className="shell">
    <aside className="sidebar">
      <div className="brand"><span className="brand-mark">fp</span><div><strong>frp-panel</strong><small>{t('controlPlane')}</small></div></div>
      <nav aria-label={t('masterConsole')}>{nav.map(({ to, key, icon }) => <NavLink key={to} to={to} end={to === '/'}><span className="nav-icon" aria-hidden="true">{icon}</span><span>{t(key)}</span></NavLink>)}</nav>
      <div className="sidebar-footnote"><span className="status online"><i />{t('online')}</span><small>Master</small></div>
    </aside>
    <div className="workspace">
      <header className="topbar">
        <div><span>{t('masterConsole')}</span><strong>{t(section)}</strong></div>
        <div className="topbar-actions">
          <label className="language-select"><span>{t('language')}</span><select aria-label={t('language')} value={language} onChange={(event) => setLanguage(event.target.value as 'zh' | 'en')}><option value="zh">{t('chinese')}</option><option value="en">{t('english')}</option></select></label>
          <NavLink to="/account" className="account-link"><span className="account-dot" aria-hidden="true">A</span>{t('account')}</NavLink>
          <button className="signout-button" onClick={() => { void logout().finally(() => navigate('/login')) }}>{t('signOut')}</button>
        </div>
      </header>
      <main className="content"><Outlet /></main>
    </div>
  </div>
}
