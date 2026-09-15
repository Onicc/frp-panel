const SESSION_KEY = 'frp-panel.authenticated'

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

export function sessionToken(): string {
  return localStorage.getItem(SESSION_KEY) ?? ''
}

export function clearSession(): void {
  localStorage.removeItem(SESSION_KEY)
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
  const result = await request<{ status?: { code?: number; message?: string } }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  if (result.status?.code !== undefined && result.status.code !== 0) {
    throw new APIError(result.status?.message || 'Login failed', 401)
  }
  localStorage.setItem(SESSION_KEY, '1')
}

export async function register(username: string, email: string, password: string): Promise<void> {
  const result = await request<{ status?: { code?: number; message?: string } }>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  })
  if (result.status?.code !== undefined && result.status.code !== 0) {
    throw new APIError(result.status.message || 'Registration failed', 400)
  }
}

export function post<T>(path: string, body: unknown = {}): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) })
}
