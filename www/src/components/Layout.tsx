import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { logout } from '../api'
import { useI18n } from '../i18n'

const nav = [
  ['/', 'overview'], ['/nodes', 'nodes'], ['/servers', 'servers'], ['/tunnels', 'tunnels'],
] as const

export function Layout() {
  const { t, language, setLanguage } = useI18n()
  const navigate = useNavigate()
  return <div className="shell">
    <aside className="sidebar">
      <div className="brand"><span className="brand-mark">fp</span><div><strong>frp-panel</strong><small>{t('controlPlane')}</small></div></div>
      <nav>{nav.map(([to, key]) => <NavLink key={to} to={to} end={to === '/'}>{t(key)}</NavLink>)}</nav>
      <div className="sidebar-footer">
        <div className="language-switch"><button className={language === 'zh' ? 'active' : ''} onClick={() => setLanguage('zh')}>中文</button><button className={language === 'en' ? 'active' : ''} onClick={() => setLanguage('en')}>EN</button></div>
		<button className="text-button" onClick={() => { void logout().finally(() => navigate('/login')) }}>{t('signOut')}</button>
      </div>
    </aside>
    <main className="content"><Outlet /></main>
  </div>
}
