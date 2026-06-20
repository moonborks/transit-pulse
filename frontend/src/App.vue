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
  return /\.\.N/.test(shapeId)
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

function buildRouteOffsets(trips: Trip[]): Map<string, number> {
  const headsignGroups = new Map<string, Set<string>>()
  for (const trip of trips) {
    if (!trip.headsign || isExpressOrZ(trip.routeId)) continue
    if (!headsignGroups.has(trip.headsign)) {
      headsignGroups.set(trip.headsign, new Set())
    }
    headsignGroups.get(trip.headsign)!.add(trip.routeId)
  }
  const sortedGroups = [...headsignGroups.entries()].sort((a, b) => {
    if (b[1].size !== a[1].size) return b[1].size - a[1].size
    return a[0].localeCompare(b[0])
  })
  const offsetByRoute = new Map<string, number>()
  const step = 4
  const base = 2
  function offsetForSlot(slot: number, step: number, base: number): number {
    const magnitude = Math.floor(slot / 2) * step + base
    return slot % 2 === 0 ? magnitude : -magnitude
  }
  function slotFromOffset(offset: number, step: number, base: number): number {
    const group = (Math.abs(offset) - base) / step
    return group * 2 + (offset < 0 ? 1 : 0)
  }
  for (const [, routes] of sortedGroups) {
    const usedOffsets = new Set<number>()

    let maxSlot = -1

    for (const routeId of routes) {
      const existing = offsetByRoute.get(routeId)
      if (existing === undefined) continue

      const slot = slotFromOffset(existing, step, base)

      if (!Number.isNaN(slot)) {
        maxSlot = Math.max(maxSlot, slot)
      }
    }

    let slot = maxSlot + 1

    for (const routeId of routes) {
      const existing = offsetByRoute.get(routeId)
      if (existing !== undefined) usedOffsets.add(existing)
    }

    for (const routeId of routes) {
      if (offsetByRoute.has(routeId)) continue

      let candidate = offsetForSlot(slot, step, base)
      while (usedOffsets.has(candidate)) {
        slot++
        candidate = offsetForSlot(slot, step, base)
      }
      offsetByRoute.set(routeId, candidate)
      usedOffsets.add(candidate)
      slot++
    }
  }

  return offsetByRoute
}

const addRoutes = (map: maplibregl.Map) => {
  const trips = mtaStore.trips ?? []
  const routeOffsets = buildRouteOffsets(trips)
  const routeShapesMap = new Map<string, Set<string>>()
  const seenEndpoints = new Map<string, Set<string>>()

  for (const trip of mtaStore.trips ?? []) {
    if (!trip.shapeId) continue

    if (!isNorthbound(trip.shapeId) || isExpressOrZ(trip.routeId)) continue

    const points = mtaStore.groupedShapes[trip.shapeId]
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
    if (!routeOffsets.has(routeId)) continue
    const offset = routeOffsets.get(routeId) ?? 0
    for (const shapeId of shapeIds) {
      const points = mtaStore.groupedShapes[shapeId]
      if (!points) continue
      features.push({
        type: 'Feature',
        properties: { color: mtaStore.getShapeColor(shapeId), offset },
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
      'line-join': 'round', // Smooths the sharp elbow corners
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
      'text-size': 11,
      'text-offset': [0.8, 0],
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

onMounted(async () => {
  await mtaStore.load()
  if (!mapEl.value) return

  map = initMap(mapEl.value)
  map.on('load', () => {
    addRoutes(map!)
    addStops(map!)
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
