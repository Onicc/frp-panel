import { describe, expect, it } from 'vitest'
import { routeParallelArcs } from './arcGeometry'

describe('routeParallelArcs', () => {
  it('separates parallel tunnels symmetrically', () => {
    const routes = routeParallelArcs([
      { id: 'a', sourceId: 'client-a', targetId: 'server-a' },
      { id: 'b', sourceId: 'client-a', targetId: 'server-a' },
      { id: 'c', sourceId: 'client-a', targetId: 'server-a' },
    ])

    expect(routes.map((route) => route.lane)).toEqual([-1, 0, 1])
    expect(routes.map((route) => route.height)).toEqual([0.77, 1, 1.23])
    expect(routes.every((route) => route.routeCount === 3)).toBe(true)
  })

  it('keeps unrelated endpoint pairs on the center lane', () => {
    const routes = routeParallelArcs([
      { id: 'a', sourceId: 'client-a', targetId: 'server-a' },
      { id: 'b', sourceId: 'client-a', targetId: 'server-b' },
    ])

    expect(routes.map((route) => [route.lane, route.height, route.routeCount])).toEqual([
      [0, 1, 1],
      [0, 1, 1],
    ])
  })
})
