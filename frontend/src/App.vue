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
    maxZoom: 18,
    maxBounds: [
      [-76.0, 39.0],
      [-71.0, 42.5],
    ],
  })
}

function isNorthbound(shapeId: string): boolean {
  return /\.\.N/.test(shapeId)
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
    if (!trip.headsign) continue
    if (!headsignGroups.has(trip.headsign)) {
      headsignGroups.set(trip.headsign, new Set())
    }
    headsignGroups.get(trip.headsign)!.add(trip.routeId)
  }

  const offsetByRoute = new Map<string, number>()

  for (const [, routeIds] of headsignGroups) {
    const routes = [...routeIds]
    if (routes.length < 2) continue
    routes.forEach((routeId, index) => {
      const magnitude = Math.floor(index / 2) * 1 + 4
      const sign = index % 2 === 0 ? 1 : -1
      offsetByRoute.set(routeId, sign * magnitude)
    })
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

    if (!isNorthbound(trip.shapeId)) continue

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

  map.on('zoom', () => {
    console.log('zoom:', map!.getZoom())
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
