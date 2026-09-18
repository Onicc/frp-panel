import { ref, type Ref } from 'vue'

const cacheKey = 'frp-panel:ip-geo-cache:v1'
const cacheTtl = 24 * 60 * 60 * 1000
const requestBatchSize = 50

export type IpGeoStatus = 'loading' | 'success' | 'error' | 'private'

export type IpGeoDetail = {
  countryCode?: string
  country?: string
  region?: string
  city?: string
  organization?: string
  timezone?: string
  latitude?: number
  longitude?: number
}

export type IpGeoEntry = {
  status: IpGeoStatus
  detail?: IpGeoDetail
  fetchedAt?: number
  error?: string
}

type GeoJSRecord = {
  country_code?: string
  country?: string
  region?: string
  city?: string
  organization?: string
  timezone?: string
  latitude?: string | number
  longitude?: string | number
}

function readCache(): Record<string, IpGeoEntry> {
  try {
    const raw = localStorage.getItem(cacheKey)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as Record<string, IpGeoEntry>
    const now = Date.now()
    return Object.fromEntries(Object.entries(parsed).filter(([, value]) =>
      value.status === 'private' || (value.fetchedAt !== undefined && now - value.fetchedAt < cacheTtl),
    ))
  } catch {
    return {}
  }
}

function saveCache(entries: Record<string, IpGeoEntry>) {
  try {
    localStorage.setItem(cacheKey, JSON.stringify(entries))
  } catch {
    // A blocked or full browser cache must not affect the topology view.
  }
}

function ipv4Parts(ip: string): number[] | undefined {
  const parts = ip.split('.')
  if (parts.length !== 4 || parts.some((part) => !/^\d{1,3}$/.test(part))) return undefined
  const values = parts.map(Number)
  return values.every((part) => part >= 0 && part <= 255) ? values : undefined
}

export function isPrivateOrLocalIp(value: string): boolean {
  const ip = value.trim().replace(/^\[|\]$/g, '').toLowerCase()
  const v4 = ipv4Parts(ip)
  if (v4) {
    const [a, b] = v4
    return a === 0 || a === 10 || a === 127 || (a === 100 && b >= 64 && b <= 127) ||
      (a === 169 && b === 254) || (a === 172 && b >= 16 && b <= 31) ||
      (a === 192 && b === 168) || (a === 192 && b === 0) || (a === 198 && (b === 18 || b === 19)) ||
      a >= 224
  }
  return ip === '::1' || ip === '::' || ip.startsWith('fc') || ip.startsWith('fd') ||
    ip.startsWith('fe8') || ip.startsWith('fe9') || ip.startsWith('fea') || ip.startsWith('feb') ||
    ip.startsWith('ff')
}

function normalizeRecord(record: GeoJSRecord): IpGeoDetail {
  const latitude = Number(record.latitude)
  const longitude = Number(record.longitude)
  return {
    countryCode: record.country_code,
    country: record.country,
    region: record.region,
    city: record.city,
    organization: record.organization,
    timezone: record.timezone,
    latitude: Number.isFinite(latitude) ? latitude : undefined,
    longitude: Number.isFinite(longitude) ? longitude : undefined,
  }
}

async function lookupIp(ip: string): Promise<IpGeoEntry> {
  try {
    const response = await fetch(`https://get.geojs.io/v1/ip/geo/${encodeURIComponent(ip)}.json`, {
      headers: { Accept: 'application/json' },
    })
    if (!response.ok) throw new Error(`GeoJS returned ${response.status}`)
    const payload = await response.json() as GeoJSRecord
    const detail = normalizeRecord(payload)
    if (detail.latitude === undefined || detail.longitude === undefined) {
      return { status: 'error', fetchedAt: Date.now(), error: 'missing coordinates' }
    }
    return { status: 'success', detail, fetchedAt: Date.now() }
  } catch (error) {
    return { status: 'error', fetchedAt: Date.now(), error: error instanceof Error ? error.message : String(error) }
  }
}

export function createIpGeoLookup() {
  const entries: Ref<Record<string, IpGeoEntry>> = ref(readCache())

  const lookupMany = async (values: string[]) => {
    const ips = [...new Set(values.map((value) => value.trim()).filter(Boolean))]
    const pending: string[] = []
    const next = { ...entries.value }
    for (const ip of ips) {
      if (isPrivateOrLocalIp(ip)) {
        next[ip] = { status: 'private' }
      } else if (next[ip]?.status !== 'success' || !next[ip]?.fetchedAt || Date.now() - next[ip].fetchedAt! >= cacheTtl) {
        next[ip] = { status: 'loading' }
        pending.push(ip)
      }
    }
    entries.value = next
    saveCache(next)

    for (let index = 0; index < pending.length; index += requestBatchSize) {
      const batch = pending.slice(index, index + requestBatchSize)
      const result = await Promise.all(batch.map(async (ip) => [ip, await lookupIp(ip)] as const))
      const updated = { ...entries.value }
      for (const [ip, entry] of result) updated[ip] = entry
      entries.value = updated
      saveCache(updated)
    }
  }

  return { entries, lookupMany }
}
