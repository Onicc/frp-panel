import { api, type ReleaseSnapshot, type VersionInfo } from '../api'

const cached = new Map<'edge'|'stable', { at: number; value: ReleaseSnapshot }>()
const pending = new Map<'edge'|'stable', Promise<ReleaseSnapshot>>()

export function versionChannel(version?: VersionInfo): 'edge'|'stable'|null {
  if (!version) return null
  if (version.gitVersion === 'main' || version.gitVersion === 'edge' || version.gitVersion.startsWith('edge-')) return 'edge'
  if (/^v\d+\.\d+\.\d+$/.test(version.gitVersion)) return 'stable'
  return null
}

export async function loadRelease(channel: 'edge'|'stable', force = false): Promise<ReleaseSnapshot> {
  const known = cached.get(channel)
  if (!force && known && Date.now() - known.at < 5 * 60_000) return known.value
  const running = pending.get(channel)
  if (running) return running
  const task = api.release(channel, force).then(({ release }) => {
    cached.set(channel, { at: Date.now(), value: release })
    return release
  }).finally(() => pending.delete(channel))
  pending.set(channel, task)
  return task
}

export function cachedRelease(channel: 'edge'|'stable'|null): ReleaseSnapshot | undefined {
  return channel ? cached.get(channel)?.value : undefined
}

export function hasNewRelease(version?: VersionInfo, snapshot?: ReleaseSnapshot): boolean | null {
  if (!version || !snapshot || snapshot.error || !snapshot.latestVersion || versionChannel(version) !== snapshot.channel) return null
  if (snapshot.channel === 'stable') {
    const current = /^v(\d+)\.(\d+)\.(\d+)$/.exec(version.gitVersion)
    const latest = /^v(\d+)\.(\d+)\.(\d+)$/.exec(snapshot.latestVersion)
    if (!current || !latest) return null
    for (let index = 1; index <= 3; index += 1) {
      if (Number(current[index]) < Number(latest[index])) return true
      if (Number(current[index]) > Number(latest[index])) return false
    }
    return false
  }
  if (!version.gitCommit || !snapshot.latestCommit) return null
  if (version.gitCommit === snapshot.latestCommit) return false
  const built = Date.parse(version.buildDate)
  const published = Date.parse(snapshot.publishedAt || '')
  if (!Number.isFinite(built) || !Number.isFinite(published)) return null
  return built < published
}

export function clientUpdateCommand(version: VersionInfo, target: string): string {
  const tag = target === 'edge' || /^v\d+\.\d+\.\d+$/.test(target) ? target : ''
  if (!tag) return ''
  if (version.platform.startsWith('darwin/')) {
    return `sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version ${tag} --restart-service=false && sudo /usr/local/libexec/frp-panel/frp-panel-agent service restart`
  }
  if (version.platform.startsWith('linux/')) {
    return `sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version ${tag} --restart-service`
  }
  if (version.platform.startsWith('windows/')) {
    return `& "$env:ProgramFiles\\frp-panel\\frp-panel-agent.exe" update --version ${tag} --restart-service`
  }
  return ''
}
