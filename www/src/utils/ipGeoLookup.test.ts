import { afterEach, describe, expect, it, vi } from 'vitest'
import { createIpGeoLookup, isPrivateOrLocalIp } from './ipGeoLookup'

afterEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
})

describe('IP geolocation lookup', () => {
  it('recognizes local and private endpoint addresses without network requests', () => {
    expect(isPrivateOrLocalIp('127.0.0.1')).toBe(true)
    expect(isPrivateOrLocalIp('10.0.0.8')).toBe(true)
    expect(isPrivateOrLocalIp('172.16.10.4')).toBe(true)
    expect(isPrivateOrLocalIp('192.168.1.20')).toBe(true)
    expect(isPrivateOrLocalIp('118.25.94.27')).toBe(false)
    expect(isPrivateOrLocalIp('43.134.184.42')).toBe(false)
  })

  it('falls back when GeoJS returns a nil coordinate', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ ip: '1.1.1.1', latitude: 'nil', longitude: 'nil' }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ success: true, country_code: 'AU', city: 'Brisbane', latitude: -27.4679, longitude: 153.028 }), { status: 200 })))

    const lookup = createIpGeoLookup()
    await lookup.lookupMany(['1.1.1.1'])

    expect(lookup.entries.value['1.1.1.1']).toMatchObject({
      status: 'success',
      detail: { countryCode: 'AU', city: 'Brisbane', latitude: -27.4679, longitude: 153.028 },
    })
    expect(fetch).toHaveBeenCalledTimes(2)
  })
})
