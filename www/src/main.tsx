import { StrictMode, Suspense, lazy, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { onSessionChange, sessionToken } from './api'
import { Layout } from './components/Layout'
import { I18nProvider } from './i18n'
import './styles.css'

const Login = lazy(() => import('./pages/Login'))
const Overview = lazy(() => import('./pages/Overview'))
const Clients = lazy(() => import('./pages/Clients'))
const Servers = lazy(() => import('./pages/Servers'))
const Tunnels = lazy(() => import('./pages/Tunnels'))
const Account = lazy(() => import('./pages/Account'))

function ProtectedLayout() {
  const [authenticated, setAuthenticated] = useState(() => Boolean(sessionToken()))
  useEffect(() => onSessionChange(() => setAuthenticated(Boolean(sessionToken()))), [])
  return authenticated ? <Layout /> : <Navigate to="/login" replace />
}

function App() {
  return <I18nProvider><BrowserRouter><Suspense fallback={<div className="app-loading">frp-panel</div>}><Routes>
    <Route path="/login" element={<Login />} />
    <Route element={<ProtectedLayout />}>
      <Route index element={<Overview />} />
      <Route path="clients" element={<Clients />} />
      <Route path="servers" element={<Servers />} />
      <Route path="tunnels" element={<Tunnels />} />
      <Route path="account" element={<Account />} />
    </Route>
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes></Suspense></BrowserRouter></I18nProvider>
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
