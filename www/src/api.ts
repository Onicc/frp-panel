const SESSION_KEY = 'frp-panel.authenticated'
const SESSION_EVENT = 'frp-panel:session-change'

type Envelope<T> = {
  code?: number
  msg?: string
  body?: T
  detail?: string
}

export class APIError extends Error {
  constructor(message: string, readonly status: number) {
    super(message)
  }
}

type RPCResult = { status?: { code?: number; message?: string } }
const RPC_SUCCESS = 1

function assertRPCSuccess(result: RPCResult, fallback: string, status: number): void {
  if (result.status?.code !== RPC_SUCCESS) {
    throw new APIError(result.status?.message || fallback, status)
  }
}

export function sessionToken(): string {
  return localStorage.getItem(SESSION_KEY) ?? ''
}

export function clearSession(): void {
  localStorage.removeItem(SESSION_KEY)
  window.dispatchEvent(new Event(SESSION_EVENT))
}

export function onSessionChange(listener: () => void): () => void {
  window.addEventListener(SESSION_EVENT, listener)
  window.addEventListener('storage', listener)
  return () => {
    window.removeEventListener(SESSION_EVENT, listener)
    window.removeEventListener('storage', listener)
  }
}

export async function logout(): Promise<void> {
  try {
    await fetch('/api/v1/auth/logout', { method: 'GET', credentials: 'same-origin' })
  } finally {
    clearSession()
  }
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...options, headers, credentials: 'same-origin' })
  if (response.status === 401) clearSession()
  const payload = (await response.json().catch(() => ({}))) as Envelope<T>
  if (!response.ok || (payload.code !== undefined && payload.code !== 200)) {
    throw new APIError(payload.detail || payload.msg || `Request failed (${response.status})`, response.status)
  }
  return (payload.body ?? payload) as T
}

export async function login(username: string, password: string): Promise<void> {
  const result = await request<RPCResult>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  assertRPCSuccess(result, 'Login failed', 401)
  localStorage.setItem(SESSION_KEY, '1')
  window.dispatchEvent(new Event(SESSION_EVENT))
}

export async function register(username: string, email: string, password: string): Promise<void> {
  const result = await request<RPCResult>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  })
  assertRPCSuccess(result, 'Registration failed', 400)
}

export function post<T>(path: string, body: unknown = {}): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) })
}
