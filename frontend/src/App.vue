<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useMtaStore } from './stores/mtaStore'
import { useTripSSE } from './composables/api/useSSE'
import { useMtaMap } from './composables/useMtaMap'
import { endpoints } from './api/endpoints'
import CircleTimer from './components/CircleTimer.vue'
import log from './utils/logger.ts'

const timerRef = ref<InstanceType<typeof CircleTimer> | null>(null)
const showTimer = ref(false)

const mapEl = ref<HTMLDivElement | null>(null)
const mtaStore = useMtaStore()
const { tripEvent } = useTripSSE()
const { initMap, addRoutes, addStops, initTrainTrackingLayers, updateTrainLocationsOnMap } =
  useMtaMap()

let map: maplibregl.Map | null = null

watch(tripEvent, async () => {
  log.debug('update train location triggered via golang SSE event')
  await mtaStore.fetchTrainLocations(endpoints.mta.trips.getLocations)
  updateTrainLocationsOnMap(map!, mtaStore.trainLocations ?? [])
  await startTimer()
})

async function startTimer() {
  showTimer.value = true
  await nextTick()
  timerRef.value?.start()
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
    <div class="overlay">
      <CircleTimer v-if="showTimer" ref="timerRef" />
    </div>
    <div ref="mapEl" class="map-container" />
  </div>
</template>

<style scoped>
.page {
  position: relative;
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

.overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  pointer-events: none;
}
</style>
