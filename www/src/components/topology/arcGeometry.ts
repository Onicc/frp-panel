export type ArcRoute = {
  lane: number
  routeCount: number
  height: number
}

type ArcEndpointPair = {
  sourceId: string
  targetId: string
}

/**
 * Give parallel Client -> Server routes deterministic heights so that
 * duplicate tunnels remain individually visible instead of being painted on
 * exactly the same pixels.
 */
export function routeParallelArcs<T extends ArcEndpointPair>(items: readonly T[]): Array<T & ArcRoute> {
  const groups = new Map<string, T[]>()
  for (const item of items) {
    const key = `${item.sourceId}\u0000${item.targetId}`
    const group = groups.get(key) || []
    group.push(item)
    groups.set(key, group)
  }

  const routes = new Map<T, ArcRoute>()
  for (const group of groups.values()) {
    const maxLane = (group.length - 1) / 2
    const heightSpan = Math.min(0.55, 0.18 + Math.max(0, group.length - 2) * 0.05)
    group.forEach((item, index) => {
      const lane = index - maxLane
      const height = maxLane === 0 ? 1 : 1 + (lane / maxLane) * heightSpan
      routes.set(item, { lane, routeCount: group.length, height })
    })
  }

  return items.map((item) => ({ ...item, ...routes.get(item)! }))
}
