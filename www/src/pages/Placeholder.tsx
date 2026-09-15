import { useLocation } from 'react-router-dom'
import { useI18n } from '../i18n'

const keys = { tunnels: 'tunnels', networks: 'networks', functions: 'functions', audit: 'audit', users: 'users', settings: 'settings' } as const

export default function Placeholder() {
  const { t } = useI18n()
  const segment = useLocation().pathname.slice(1) as keyof typeof keys
  const key = keys[segment] ?? 'settings'
  return <><header className="page-header"><div><span className="eyebrow">frp-panel v2</span><h1>{t(key)}</h1></div></header><section className="panel empty-module"><span>◇</span><h2>{t(key)}</h2><p>{t('comingSoon')}</p></section></>
}
