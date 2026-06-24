<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useMtaStore } from './stores/mtaStore'
import type { TrainLocation, Trip } from './types/mta'
import { useTripSSE } from './composables/api/useSSE'

const mtaStore = useMtaStore()
const mapEl = ref<HTMLDivElement | null>(null)
let map: maplibregl.Map | null = null

const { tripEvent } = useTripSSE()
watch(tripEvent, async () => {
  console.log('update train location triggered via golang SSE event')
  updateTrainLocationsOnMap(map!, mtaStore.trainLocations ?? [])
})

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

function addArrowIcon(map: maplibregl.Map) {
  const size = 20
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')!

  // Wider arrowhead: tip near top, base spans almost full width,
  // and the base sits higher up (shorter height) so it reads as "wide" not "tall"
  ctx.beginPath()
  ctx.moveTo(size / 2, size * 0.1) // tip (near top)
  ctx.lineTo(size * 0.95, size * 0.85) // bottom-right corner, pulled out wide
  ctx.lineTo(size / 2, size * 0.6) // notch (inward point for arrowhead look)
  ctx.lineTo(size * 0.05, size * 0.85) // bottom-left corner, pulled out wide
  ctx.closePath()
  ctx.fillStyle = '#ffffff'
  ctx.fill()

  map.addImage(
    'train-arrow',
    {
      width: size,
      height: size,
      data: ctx.getImageData(0, 0, size, size).data,
    },
    { sdf: true },
  )
}

const initTrainTrackingLayers = (map: maplibregl.Map) => {
  map.addSource('train-locations', {
    type: 'geojson',
    data: { type: 'FeatureCollection', features: [] },
  })

  map.addLayer({
    id: 'train-circles',
    type: 'circle',
    source: 'train-locations',
    minzoom: 10,
    layout: {
      // High value renders last (on top)
      'circle-sort-key': ['get', 'circle_priority'],
    },
    paint: {
      'circle-radius': 11,
      'circle-color': ['get', 'color'],
      'circle-stroke-width': 2,
      'circle-stroke-color': '#000000',
    },
  })

  addArrowIcon(map)

  const CIRCLE_RADIUS = 11
  const ARROW_HALF_HEIGHT = 8
  const GAP = 8

  map.addLayer({
    id: 'train-arrows',
    type: 'symbol',
    source: 'train-locations',
    minzoom: 13,
    layout: {
      'icon-image': 'train-arrow',
      'icon-size': 1,
      'icon-rotate': ['get', 'bearing'],
      'icon-rotation-alignment': 'map',
      'icon-allow-overlap': true,
      'icon-ignore-placement': true,
      'icon-offset': [0, -(CIRCLE_RADIUS + GAP - ARROW_HALF_HEIGHT)],
      'symbol-sort-key': ['get', 'circle_priority'],
    },
    paint: {
      'icon-color': ['get', 'color'],
    },
  })

  map.addLayer({
    id: 'train-labels',
    type: 'symbol',
    source: 'train-locations',
    minzoom: 10,
    layout: {
      'text-field': ['get', 'routeId'],
      'text-font': ['Open Sans Bold'],
      'text-size': 11,
      'text-allow-overlap': false,
      'text-ignore-placement': false,
      'text-padding': 2,
      // Low value processes first (wins the text space)
      'symbol-sort-key': ['get', 'label_priority'],
    },
    paint: {
      'text-color': '#ffffff',
    },
  })
}

const updateTrainLocationsOnMap = (map: maplibregl.Map, trainLocations: TrainLocation[]) => {
  const source = map.getSource('train-locations') as maplibregl.GeoJSONSource
  if (!source) return

  const total = trainLocations.length

  const features = trainLocations.map((train, index) => {
    // Top train (end of array) gets high circle priority to render on top
    const circlePriority = index + 1

    // Top train (end of array) gets low label priority (1) to win the text slot
    const labelPriority = total - index

    return {
      type: 'Feature' as const,
      id: `${train.tripId}_${train.nextStopId}`,
      geometry: {
        type: 'Point' as const,
        coordinates: [train.lon, train.lat] as [number, number],
      },
      properties: {
        tripId: train.tripId,
        routeId: train.routeId,
        nextStopId: train.nextStopId,
        bearing: train.bearing,
        color: mtaStore.getRouteColor(train.routeId),
        circle_priority: circlePriority,
        label_priority: labelPriority,
      },
    }
  })

  source.setData({
    type: 'FeatureCollection',
    features,
  })
}

onMounted(async () => {
  await mtaStore.load()
  if (!mapEl.value) return

  map = initMap(mapEl.value)
  map.on('load', () => {
    addRoutes(map!)
    addStops(map!)
    initTrainTrackingLayers(map!)
    updateTrainLocationsOnMap(map!, mtaStore.trainLocations ?? [])
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
