import { describe, expect, it } from 'vitest'
import { buildServerCompose } from './Servers'

describe('FRPS deployment', () => {
  it('generates a standalone persistent Compose service without a long-lived secret', () => {
    const compose = buildServerCompose(
      { serverId: 'owner.s.edge', token: 'one-use-token', expiresAt: '2030-01-01T00:00:00Z' },
      'https://panel.example.test',
      'wss://panel.example.test',
    )
    expect(compose).toContain('network_mode: host')
    expect(compose).toContain('- server')
    expect(compose).toContain('/data/server.yaml')
    expect(compose).toContain('"one-use-token"')
    expect(compose).toContain('"https://panel.example.test"')
    expect(compose).not.toContain('permanent-secret')
  })
})
