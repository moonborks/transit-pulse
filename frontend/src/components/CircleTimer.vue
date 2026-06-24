<script setup lang="ts">
import { now } from 'maplibre-gl'
import { computed, onBeforeUnmount, ref, type Ref } from 'vue'

const intervalMS: number = 30_000
const radius: number = 24
const center: number = 60

const progress: Ref<number> = ref<number>(0)

const circumference = computed(() => 2 * Math.PI * radius)
const strokeDasharray = computed(() => `${circumference.value}`)
const strokeDashoffset = computed(() => {
  return circumference.value * (1 - progress.value / 100)
})

let rafID: number | null = null
let startTime: number = 0

function stop() {
  if (rafID) cancelAnimationFrame(rafID)
  rafID = null
}

function start() {
  stop()
  startTime = now()
  progress.value = 0
  rafID = requestAnimationFrame(tick)
}

function tick() {
  const elapsedTime = now() - startTime
  const p = Math.min(1, elapsedTime / intervalMS)
  progress.value = p * 100

  if (p < 1) rafID = requestAnimationFrame(tick)
}

defineExpose({ start })

onBeforeUnmount(stop)
</script>

<template>
  <div class="ring">
    <svg viewBox="0 0 120 120" class="ring-svg">
      <circle class="ring-bg" :cx="center" :cy="center" :r="radius"></circle>
      <circle
        class="ring-fg"
        :cx="center"
        :cy="center"
        :r="radius"
        :style="{ strokeDasharray, strokeDashoffset }"
      ></circle>
    </svg>
  </div>
</template>

<style lang="css" scoped>
.ring {
  position: relative;
  width: 120px;
  height: 120px;
}

.ring-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-bg {
  fill: none;
  stroke: #e6e6e6;
  stroke-width: 10;
}

.ring-fg {
  fill: none;
  stroke: #3b82f6;
  stroke-width: 10;
  stroke-linecap: round;
}
</style>
