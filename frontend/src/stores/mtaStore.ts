import { defineStore } from 'pinia'
import {
  type Stop,
  type ShapePoint,
  type Route,
  type Trip,
  type RouteAPI,
  type TripAPI,
} from '@/types/mta'
import { useFetch } from '@/composables/api/useFetch'
import { computed } from 'vue'
import { normalizeRouteFields, normalizeTripFields } from '@/utils/normalizer'

export const useMtaStore = defineStore('mta', () => {
  const {
    data: stops,
    loading: stopsLoading,
    error: stopsError,
    fetchData: fetchStops,
  } = useFetch<Stop[]>()
  const {
    data: shapes,
    loading: shapesLoading,
    error: shapesError,
    fetchData: fetchShapes,
  } = useFetch<ShapePoint[]>()
  const {
    data: trips,
    loading: tripsLoading,
    error: tripsError,
    fetchData: fetchTrips,
  } = useFetch<TripAPI[], Trip[]>((data) => data.map(normalizeTripFields))
  const {
    data: routes,
    loading: routesLoading,
    error: routesError,
    fetchData: fetchRoutes,
  } = useFetch<RouteAPI[], Route[]>((data) => data.map(normalizeRouteFields))

  const load = async () => {
    await Promise.all([
      fetchStops('/api/mta/stops'),
      fetchShapes('/api/mta/shapes?simplify=true'),
      fetchTrips('/api/mta/trips/today'),
      fetchRoutes('/api/mta/routes'),
    ])
  }

  const groupedShapes = computed(() => {
    return groupShapePoints(shapes.value)
  })

  function groupShapePoints(points: ShapePoint[] | null): Record<string, [number, number][]> {
    const grouped: Record<string, [number, number][]> = {}
    if (points === null) {
      return grouped
    }
    for (const p of points) {
      const pts = grouped[p.id] ?? []
      pts.push([p.lon, p.lat])
      grouped[p.id] = pts
    }
    return grouped
  }

  const shapeColorMap = computed(() => {
    const map = new Map<string, string>()
    for (const trip of trips.value ?? []) {
      if (!trip.shapeId) continue
      const route = routes.value?.find((r) => r.id === trip.routeId)
      map.set(trip.shapeId, route ? `#${route.color}` : '#888888')
    }
    return map
  })

  function getShapeColor(shapeId: string): string {
    return shapeColorMap.value.get(shapeId) ?? '#888888'
  }

  return {
    stops,
    groupedShapes,
    routes,
    trips,
    stopsLoading,
    stopsError,
    shapesLoading,
    shapesError,
    routesLoading,
    routesError,
    tripsLoading,
    tripsError,
    load,
    getShapeColor,
  }
})
