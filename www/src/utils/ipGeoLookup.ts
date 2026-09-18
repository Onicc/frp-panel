import { ref, type Ref } from 'vue'

const cacheKey = 'frp-panel:ip-geo-cache:v1'
const cacheTtl = 24 * 60 * 60 * 1000
const requestBatchSize = 8

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

type IpWhoRecord = {
  success?: boolean
  country_code?: string
  country?: string
  region?: string
  city?: string
  latitude?: string | number
  longitude?: string | number
  connection?: { org?: string; isp?: string }
  timezone?: { id?: string }
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

function coordinate(value: string | number | undefined, limit: number): number | undefined {
  if (value === undefined || value === null) return undefined
  if (typeof value === 'string' && ['nil', 'null', 'nan', 'n/a', 'na', ''].includes(value.trim().toLowerCase())) return undefined
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) && Math.abs(parsed) <= limit ? parsed : undefined
}

function normalizeGeoJSRecord(record: GeoJSRecord): IpGeoDetail {
  const latitude = coordinate(record.latitude, 90)
  const longitude = coordinate(record.longitude, 180)
  return {
    countryCode: record.country_code,
    country: record.country,
    region: record.region,
    city: record.city,
    organization: record.organization,
    timezone: record.timezone,
    latitude,
    longitude,
  }
}

function normalizeIpWhoRecord(record: IpWhoRecord): IpGeoDetail {
  return {
    countryCode: record.country_code,
    country: record.country,
    region: record.region,
    city: record.city,
    organization: record.connection?.org || record.connection?.isp,
    timezone: record.timezone?.id,
    latitude: coordinate(record.latitude, 90),
    longitude: coordinate(record.longitude, 180),
  }
}

function hasCoordinates(detail: IpGeoDetail) {
  return detail.latitude !== undefined && detail.longitude !== undefined
}

async function getJSON<T>(url: string): Promise<T> {
  const response = await fetch(url, { headers: { Accept: 'application/json' } })
  if (!response.ok) throw new Error(`Geolocation service returned ${response.status}`)
  return await response.json() as T
}

async function lookupIp(ip: string): Promise<IpGeoEntry> {
  const errors: string[] = []
  try {
    const detail = normalizeGeoJSRecord(await getJSON<GeoJSRecord>(`https://get.geojs.io/v1/ip/geo/${encodeURIComponent(ip)}.json`))
    if (hasCoordinates(detail)) return { status: 'success', detail, fetchedAt: Date.now() }
    errors.push('GeoJS returned no coordinates')
  } catch (error) {
    errors.push(error instanceof Error ? error.message : String(error))
  }
  try {
    const payload = await getJSON<IpWhoRecord>(`https://ipwho.is/${encodeURIComponent(ip)}`)
    if (payload.success === false) throw new Error('ipwho.is could not locate the address')
    const detail = normalizeIpWhoRecord(payload)
    if (hasCoordinates(detail)) return { status: 'success', detail, fetchedAt: Date.now() }
    errors.push('ipwho.is returned no coordinates')
  } catch (error) {
    errors.push(error instanceof Error ? error.message : String(error))
  }
  return { status: 'error', fetchedAt: Date.now(), error: errors.join('; ') || 'missing coordinates' }
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
