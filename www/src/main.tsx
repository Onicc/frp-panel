import { StrictMode, Suspense, lazy } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { sessionToken } from './api'
import { Layout } from './components/Layout'
import { I18nProvider } from './i18n'
import './styles.css'

const Login = lazy(() => import('./pages/Login'))
const Overview = lazy(() => import('./pages/Overview'))
const Nodes = lazy(() => import('./pages/Nodes'))
const Servers = lazy(() => import('./pages/Servers'))
const Placeholder = lazy(() => import('./pages/Placeholder'))

function ProtectedLayout() {
  return sessionToken() ? <Layout /> : <Navigate to="/login" replace />
}

function App() {
  return <I18nProvider><BrowserRouter><Suspense fallback={<div className="app-loading">frp-panel</div>}><Routes>
    <Route path="/login" element={<Login />} />
    <Route element={<ProtectedLayout />}>
      <Route index element={<Overview />} />
      <Route path="nodes" element={<Nodes />} />
      <Route path="servers" element={<Servers />} />
      {['tunnels', 'networks', 'functions', 'audit', 'users', 'settings'].map((path) => <Route key={path} path={path} element={<Placeholder />} />)}
    </Route>
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes></Suspense></BrowserRouter></I18nProvider>
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
