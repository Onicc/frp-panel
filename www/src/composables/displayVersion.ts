const stableVersion = /^v?(\d+\.\d+\.\d+)$/
type VersionLabels = { legacy?: string; development?: string }

/** Never invent a stable version for a legacy or local build. */
export function displayVersion(raw?: string, labels: VersionLabels = {}): string {
  if (!raw) return '—'
  const stable = stableVersion.exec(raw)
  if (stable) return `v${stable[1]}`
  if (raw === 'main' || raw === 'edge' || raw.startsWith('edge-')) return labels.legacy || 'Legacy build'
  return labels.development || 'Development build'
}

export function displayVersionDetails(raw?: string, commit?: string, labels: VersionLabels = {}): string {
  const label = displayVersion(raw, labels)
  return commit ? `${label} · ${commit.slice(0, 7)}` : label
}
