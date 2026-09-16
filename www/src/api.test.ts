import { afterEach, describe, expect, it, vi } from 'vitest'
import { APIError, login } from './api'

afterEach(() => {
  vi.unstubAllGlobals()
  localStorage.clear()
})

describe('authentication protocol status', () => {
  it('accepts protobuf success code 1 and records the browser session', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      code: 200, msg: 'ok', body: { status: { code: 1, message: 'ok' } },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } })))

    await login('owner', 'correct-password')

    expect(localStorage.getItem('frp-panel.authenticated')).toBe('1')
  })

  it('surfaces a nested authentication failure instead of the envelope message', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      code: 200, msg: 'ok', body: { status: { code: 2, message: 'invalid username or password' } },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } })))

    await expect(login('owner', 'wrong-password')).rejects.toEqual(expect.objectContaining<Partial<APIError>>({
      message: 'invalid username or password', status: 401,
    }))
    expect(localStorage.getItem('frp-panel.authenticated')).toBeNull()
  })
})
