<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { useMtaStore } from "./stores/mtaStore";

const mtaStore = useMtaStore();

const mapEl = ref<HTMLDivElement | null>(null);

let map: L.Map | null = null;

const BOUNDS = L.latLngBounds(
  [39.0, -76.0], // SW
  [42.5, -71.0], // NE
);

onMounted(async () => {
  await mtaStore.load();
  if (!mapEl.value) return;
  map = L.map(mapEl.value, {
    center: [40.706, -74.013],
    zoom: 13,
    minZoom: 12,
    maxZoom: 18,
    maxBounds: BOUNDS,
    maxBoundsViscosity: 0.8,
    zoomSnap: 0.5,
    zoomDelta: 0.5,
  }).setView([40.706, -74.013], 13);

  L.tileLayer("https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png", {
    attribution: "© CartoDB © OpenStreetMap contributors",
  }).addTo(map);

  const shapes = mtaStore.groupedShapes;
  for (const [shapeId, points] of Object.entries(shapes)) {
    L.polyline(points, { color: mtaStore.getShapeColor(shapeId), weight: 3 }).addTo(map);
  }
  for (const stop of mtaStore.stops ?? []) {
    L.circle([stop.lat, stop.lon], {
      radius: 1,
      color: "#333333",
      fillColor: "#333333",
      fillOpacity: 1,
      weight: 1,
    })
      .addTo(map)
      .bindTooltip(stop.name, {
        permanent: true,
        direction: "right",
        className: "stop-label",
        offset: [0, 0],
      });
  }
  if (map !== null) {
    map.on("zoom", () => {
      console.log("zoom:", map!.getZoom());
    });
  }
});

onUnmounted(() => {
  map?.remove();
  map = null;
});
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

<style>
.leaflet-tooltip.stop-label {
  background-color: transparent;
  border: none;
  box-shadow: none;
  font-family: Arial, sans-serif;
  font-size: 11px;
  font-weight: bold;
  color: #333333; /* Dark gray text color */
  display: block !important;
  white-space: normal !important;
  width: 200px !important; /* Hard limit: change this number to wrap earlier/later */
  line-height: 1.1;
}

/* Removes the tiny arrow pointing to the marker */
.leaflet-tooltip-left.stop-label::before,
.leaflet-tooltip-right.stop-label::before {
  border: none;
}
</style>
