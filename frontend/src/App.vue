<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useMtaStore } from './stores/mtaStore'

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

const addRoutes = (map: maplibregl.Map) => {
  const features = Object.entries(mtaStore.groupedShapes).map(([shapeId, points]) => ({
    type: 'Feature' as const,
    properties: { color: mtaStore.getShapeColor(shapeId) },
    geometry: {
      type: 'LineString' as const,
      coordinates: points.map(([lat, lon]) => [lon, lat]),
    },
  }))

  map.addSource('routes', {
    type: 'geojson',
    data: { type: 'FeatureCollection', features },
  })

  map.addLayer({
    id: 'routes',
    type: 'line',
    source: 'routes',
    paint: {
      'line-color': ['get', 'color'],
      'line-width': 2,
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
