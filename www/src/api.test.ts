import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from './api'

afterEach(() => vi.restoreAllMocks())

describe('v2 API client', () => {
  it('surfaces RFC 9457 details instead of treating a failed login as ok', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ title: 'Sign-in failed', detail: 'the username or password is incorrect' }), { status: 401, headers: { 'Content-Type': 'application/problem+json' } })))
    await expect(api.login('owner', 'bad')).rejects.toMatchObject({ status: 401, message: 'the username or password is incorrect' })
  })

  it('uses resource-oriented endpoints and returns the page envelope', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ items: [], total: 0, page: 1, pageSize: 25 }), { status: 200 })))
    const result = await api.clients({ page: 1, pageSize: 25 })
    expect(result.items).toEqual([])
    expect(String((fetch as ReturnType<typeof vi.fn>).mock.calls[0][0])).toContain('/api/v2/clients')
  })
})
