<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useMtaStore } from './stores/mtaStore'
import type { Trip } from './types/mta'

const mtaStore = useMtaStore()
const mapEl = ref<HTMLDivElement | null>(null)
let map: maplibregl.Map | null = null

const initMap = (el: HTMLDivElement): maplibregl.Map => {
  return new maplibregl.Map({
    container: el,
    style: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
    center: [-74.013, 40.706],
    zoom: 13,
    minZoom: 10,
    maxZoom: 17,
    maxBounds: [
      [-76.0, 39.0],
      [-71.0, 42.5],
    ],
  })
}

function isNorthbound(shapeId: string): boolean {
  return /\.+N/.test(shapeId)
}

function isExpressOrZ(routeId: string): boolean {
  return /X\d*$/.test(routeId) || /Z\d*$/.test(routeId)
}

function getEndpointKey(points: [number, number][]): string {
  const start = points[0]
  const end = points[points.length - 1]
  if (!start || !end) return ''
  return `${start[0]},${start[1]}|${end[0]},${end[1]}`
}

function buildShapeOffsets(
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

      // Look at the first 5 track points to evaluate the local terminal corridor space
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

    // Scan the global path ledger for track path conflicts
    while (hasCollision) {
      hasCollision = false

      // If ANY of our line's initial coordinates overlap with a lane slot already claimed
      // by a different route, we must increment the slot lane globally.
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

    // Record this route path's coordinates to block other trains from taking this lane
    for (const key of routeCoordinateKeys) {
      if (!globalCoordinateOffsets.has(key)) {
        globalCoordinateOffsets.set(key, new Set())
      }
      globalCoordinateOffsets.get(key)!.add(candidate)
    }
  }

  return offsetByShape
}

const addRoutes = (map: maplibregl.Map) => {
  const trips = mtaStore.trips ?? []
  const groupedShapes = mtaStore.groupedShapes ?? {}

  const shapeOffsets = buildShapeOffsets(trips, groupedShapes)

  const routeShapesMap = new Map<string, Set<string>>()
  const seenEndpoints = new Map<string, Set<string>>()

  for (const trip of trips) {
    if (!trip.shapeId) continue
    if (!isNorthbound(trip.shapeId) || isExpressOrZ(trip.routeId)) continue

    const points = groupedShapes[trip.shapeId]
    if (!points || points.length === 0) continue

    const endpointKey = getEndpointKey(points)

    if (!seenEndpoints.has(trip.routeId)) {
      seenEndpoints.set(trip.routeId, new Set())
    }
    const routeEndpoints = seenEndpoints.get(trip.routeId)!
    if (routeEndpoints.has(endpointKey)) continue
    routeEndpoints.add(endpointKey)

    if (!routeShapesMap.has(trip.routeId)) {
      routeShapesMap.set(trip.routeId, new Set())
    }
    routeShapesMap.get(trip.routeId)!.add(trip.shapeId)
  }

  const features: Array<{
    type: 'Feature'
    properties: { color: string; offset: number }
    geometry: { type: 'LineString'; coordinates: [number, number][] }
  }> = []

  for (const [routeId, shapeIds] of routeShapesMap) {
    for (const shapeId of shapeIds) {
      const points = groupedShapes[shapeId]
      if (!points) continue

      const offset = shapeOffsets.get(shapeId) ?? 0

      features.push({
        type: 'Feature',
        properties: {
          color: mtaStore.getRouteColor(routeId),
          offset,
        },
        geometry: {
          type: 'LineString',
          coordinates: points,
        },
      })
    }
  }

  map.addSource('shapes', {
    type: 'geojson',
    data: { type: 'FeatureCollection', features },
  })

  map.addLayer({
    id: 'shapes-layer',
    type: 'line',
    source: 'shapes',
    layout: {
      'line-join': 'round',
      'line-cap': 'round',
    },
    paint: {
      'line-color': ['get', 'color'],
      'line-width': 4,
      'line-offset': [
        'interpolate',
        ['linear'],
        ['zoom'],
        10,
        ['*', ['get', 'offset'], 0.3],
        16,
        ['get', 'offset'],
      ],
    },
  })
}

const addStops = (map: maplibregl.Map) => {
  map.addSource('stops', {
    type: 'geojson',
    data: {
      type: 'FeatureCollection',
      features: (mtaStore.stops ?? []).map((stop) => ({
        type: 'Feature' as const,
        properties: { name: stop.name },
        geometry: {
          type: 'Point' as const,
          coordinates: [stop.lon, stop.lat],
        },
      })),
    },
  })

  map.addLayer({
    id: 'stops',
    type: 'circle',
    source: 'stops',
    paint: {
      'circle-radius': 1,
      'circle-color': '#ffffff',
      'circle-stroke-color': '#333333',
      'circle-stroke-width': 1.5,
    },
  })

  map.addLayer({
    id: 'stop-labels',
    type: 'symbol',
    source: 'stops',
    layout: {
      'text-field': ['get', 'name'],
      'text-font': ['Open Sans Bold'],
      'text-size': 11,
      'text-offset': [1, 0],
      'text-anchor': 'left',
      'text-optional': true,
    },
    paint: {
      'text-color': '#333333',
      'text-halo-color': '#ffffff',
      'text-halo-width': 1.5,
    },
  })
}

const addTrainLocations = (map: maplibregl.Map) => {
  const lookup = mtaStore.stopLocationLookup

  const features: Array<{
    type: 'Feature'
    geometry: { type: 'Point'; coordinates: [number, number] }
    properties: {
      stopId: string
      routeId: string
      color: string
      priority: number
    }
  }> = []
  let priority = 1

  for (const nextStop of mtaStore.nextStops ?? []) {
    if (nextStop.routeId === 'FS') {
      console.log(lookup.get(nextStop.stopId))
    }
    const stopLocation = lookup.get(nextStop.stopId)
    if (stopLocation === undefined) continue
    features.push({
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [stopLocation.lon, stopLocation.lat] },
      properties: {
        stopId: nextStop.stopId,
        routeId: nextStop.routeId,
        color: mtaStore.getRouteColor(nextStop.routeId),
        priority: priority++,
      },
    })
  }
  map.addSource('next-stops', {
    type: 'geojson',
    data: {
      type: 'FeatureCollection',
      features,
    },
  })
  map.addLayer({
    id: 'next-stops-circles',
    type: 'circle',
    source: 'next-stops',
    minzoom: 11,
    layout: {
      'circle-sort-key': ['get', 'priority'],
    },
    paint: {
      'circle-radius': 8,
      'circle-color': ['get', 'color'],
      'circle-stroke-width': 1.5,
      'circle-stroke-color': '#000000',
    },
  })

  map.addLayer({
    id: 'next-stops-labels',
    type: 'symbol',
    source: 'next-stops',
    minzoom: 11,
    layout: {
      'text-field': ['get', 'routeId'],
      'text-font': ['Open Sans Bold'],
      'text-size': 12,
      'text-allow-overlap': false,
      'symbol-sort-key': ['get', 'priority'],
    },
    paint: {
      'text-color': '#ffffff',
    },
  })
}

onMounted(async () => {
  await mtaStore.load()
  if (!mapEl.value) return

  map = initMap(mapEl.value)
  map.on('load', () => {
    addRoutes(map!)
    addStops(map!)
    addTrainLocations(map!)
  })
})

onUnmounted(() => {
  map?.remove()
  map = null
})
</script>

<template>
  <div class="page">
    <div ref="mapEl" class="map-container" />
  </div>
</template>

<style scoped>
.page {
  height: 100%;
  width: 100%;
}
.map-container {
  height: 100vh;
  width: 100vw;
  position: fixed;
  top: 0;
  left: 0;
}
</style>
