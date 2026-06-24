import type { Trip } from '../types/mta'

export function isNorthbound(shapeId: string): boolean {
  return /\.+N/.test(shapeId)
}

export function isExpressOrZ(routeId: string): boolean {
  return /X\d*$/.test(routeId) || /Z\d*$/.test(routeId)
}

export function getEndpointKey(points: [number, number][]): string {
  const start = points[0]
  const end = points[points.length - 1]
  if (!start || !end) return ''
  return `${start[0]},${start[1]}|${end[0]},${end[1]}`
}

export function buildShapeOffsets(
  trips: Trip[],
  groupedShapes: Record<string, [number, number][]>,
): Map<string, number> {
  const routeShapesMap = new Map<string, Set<string>>()
  for (const trip of trips) {
    if (!trip.shapeId || !isNorthbound(trip.shapeId) || isExpressOrZ(trip.routeId)) continue
    if (!routeShapesMap.has(trip.routeId)) {
      routeShapesMap.set(trip.routeId, new Set())
    }
    routeShapesMap.get(trip.routeId)!.add(trip.shapeId)
  }

  const offsetByShape = new Map<string, number>()
  const globalCoordinateOffsets = new Map<string, Set<number>>()
  const step = 5
  const base = 3

  function offsetForSlot(slot: number, step: number, base: number): number {
    const group = Math.floor(slot / 4)
    const withinGroup = slot % 4
    const isPositive = withinGroup < 2
    const stepMultiplier = group * 2 + (withinGroup % 2)
    const magnitude = stepMultiplier * step + base
    return isPositive ? magnitude : -magnitude
  }

  for (const [, shapeIds] of routeShapesMap) {
    const routeCoordinateKeys: string[] = []

    for (const shapeId of shapeIds) {
      const points = groupedShapes[shapeId]
      if (!points || points.length === 0) continue

      const pointsToAnalyze = points.slice(0, 5)
      for (const [lon, lat] of pointsToAnalyze) {
        routeCoordinateKeys.push(`${lat.toFixed(4)},${lon.toFixed(4)}`)
      }

      if (points.length > 5) {
        const endPoints = points.slice(-5)
        for (const [lon, lat] of endPoints) {
          routeCoordinateKeys.push(`${lat.toFixed(4)},${lon.toFixed(4)}`)
        }
      }
    }

    let slot = 0
    let candidate = offsetForSlot(slot, step, base)
    let hasCollision = true

    while (hasCollision) {
      hasCollision = false
      for (const key of routeCoordinateKeys) {
        if (globalCoordinateOffsets.get(key)?.has(candidate)) {
          hasCollision = true
          break
        }
      }
      if (hasCollision) {
        slot++
        candidate = offsetForSlot(slot, step, base)
      }
    }

    for (const shapeId of shapeIds) {
      offsetByShape.set(shapeId, candidate)
    }

    for (const key of routeCoordinateKeys) {
      if (!globalCoordinateOffsets.has(key)) {
        globalCoordinateOffsets.set(key, new Set())
      }
      globalCoordinateOffsets.get(key)!.add(candidate)
    }
  }

  return offsetByShape
}
